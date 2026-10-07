package game

import (
	"encoding/json"
	"strings"
	"testing"
)

func currentMatch(t *testing.T) *Match {
	t.Helper()
	m, err := New("NEW234", []string{"norway", "germany", "france"}, 720, testNow)
	if err != nil {
		t.Fatal(err)
	}
	return m
}
func TestNewSheetBudgetIncomeAndLevels(t *testing.T) {
	m := currentMatch(t)
	c := m.Country("norway")
	if len(c.Cities) != 5 || c.Score() != 75 || c.BalanceCents != 70000 {
		t.Fatalf("initial country: %+v", c)
	}
	m.Advance(testNow)
	if c.IncomeCents != 0 || c.BalanceCents != 70000 {
		t.Fatal("first phase must not award income")
	}
	c.Plan.Upgrades = []string{"norway-2", "norway-5"}
	m.Country("germany").Plan.Nuclear = true
	m.Country("france").Plan.Sanctions = []string{"norway"}
	m.Advance(testNow)
	// 700 - 400 + (1000 * .94 + 75) * .85 = 1162.75.
	if c.BalanceCents != 116275 || c.IncomeCents != 86275 || m.Pollution != 6 {
		t.Fatalf("budget=%d income=%d pollution=%d", c.BalanceCents, c.IncomeCents, m.Pollution)
	}
	if c.Level("military") != 1 || c.Level("tourism") != 1 || m.Prices(c)["bomb"] != 25000 {
		t.Fatal("city levels and discount")
	}
	m.Advance(testNow)
	if c.BalanceCents != 116275 {
		t.Fatal("income credited twice")
	}
	c.NuclearRound = 2
	c.Plan.Bombs = 2
	c.Plan.Upgrades = []string{"norway-2"}
	c.Plan.Ecology = true
	m.Advance(testNow)
	// Discount uses level before this round's upgrade: 2 * 250 + 200 + 200.
	if c.BalanceCents != 133775 || c.Level("military") != 2 || m.Pollution != 0 {
		t.Fatalf("discount snapshot or cleanup: %+v", c)
	}
	c.CityByRole("military").Level = 100
	c.CityByRole("security").Level = 100
	if m.Prices(c)["bomb"] != 0 || m.Prices(c)["shield"] != 0 {
		t.Fatal("discount prices must be nonnegative")
	}
	m.Pollution = 110
	c.SanctionedBy = []string{}
	if m.NextIncome(c) != 7500 {
		t.Fatal("pollution clamps base income, preserves tourism")
	}
	c.SanctionedBy = make([]string, 7)
	m.Pollution = 0
	if m.NextIncome(c) != 0 {
		t.Fatal("sanctions cannot produce negative income")
	}
}
func TestNewIntelligenceSnapshotAndPrivacy(t *testing.T) {
	m := currentMatch(t)
	m.Phase = "headquarters"
	c := m.Country("norway")
	p := c.Plan
	p.Upgrades = []string{"norway-4"}
	p.Spies = []string{"germany"}
	if err := execute(t, m, Actor{CountryID: c.ID}, "plan", p); err != nil {
		t.Fatal(err)
	}
	m.Country("germany").Plan.Nuclear = true
	m.Advance(testNow)
	if c.Level("intelligence") != 0 || len(c.Intelligence) != 1 {
		t.Fatal("one level buys one report")
	}
	report := c.Intelligence[0]
	if report.BalanceCents != 114000 || report.NuclearRound != 2 || len(report.Cities) != 5 {
		t.Fatalf("report %+v", report)
	}
	m.Country("germany").BalanceCents++
	if c.Intelligence[0].BalanceCents != 114000 {
		t.Fatal("reports must be historical snapshots")
	}
	for _, a := range []Actor{{}, {CountryID: "germany"}} {
		view := m.View(a, testNow)
		countries := view["countries"].([]any)
		if _, ok := countries[0].(map[string]any)["intelligence"]; ok {
			t.Fatal("foreign report leaked")
		}
		b, _ := json.Marshal(view)
		if strings.Contains(string(b), "\"spy\"") {
			t.Fatal("foreign report leaked")
		}
	}
	view := m.View(Actor{CountryID: c.ID}, testNow)
	b, _ := json.Marshal(view)
	if !strings.Contains(string(b), "countryName") {
		t.Fatal("own report missing")
	}
}
func TestSpyCannotSurviveCancelledUpgrade(t *testing.T) {
	m := currentMatch(t)
	m.Phase = "headquarters"
	c := m.Country("norway")
	c.BalanceCents = 0
	c.Plan.Upgrades = []string{"norway-4"}
	c.Plan.Spies = []string{"germany"}
	m.Advance(testNow)
	if len(c.Intelligence) != 0 || c.Level("intelligence") < 0 {
		t.Fatal("unfunded spy order executed")
	}
}
func TestCapitalLevelSpentOnlyOnAcceptedExtraMeeting(t *testing.T) {
	m := currentMatch(t)
	m.Phase = "headquarters"
	sender := Actor{CountryID: "norway"}
	c := m.Country(sender.CountryID)
	c.CityByRole("capital").Level = 1
	for _, target := range []string{"germany", "france"} {
		if err := execute(t, m, sender, "meeting.request", map[string]any{"countryId": target, "double": true}); err != nil {
			t.Fatal(err)
		}
		mt := m.Meetings[len(m.Meetings)-1]
		if target == "france" {
			if c.Level("capital") != 1 {
				t.Fatal("pending visit spent level")
			}
			if err := execute(t, m, Actor{CountryID: target}, "meeting.respond", map[string]any{"id": mt.ID, "accept": false}); err != nil {
				t.Fatal(err)
			}
			if c.Level("capital") != 1 {
				t.Fatal("rejection spent level")
			}
			if err := execute(t, m, sender, "meeting.request", map[string]any{"countryId": target, "double": true}); err != nil {
				t.Fatal(err)
			}
			mt = m.Meetings[len(m.Meetings)-1]
		}
		if err := execute(t, m, Actor{CountryID: target}, "meeting.respond", map[string]any{"id": mt.ID, "accept": true}); err != nil {
			t.Fatal(err)
		}
	}
	if c.Level("capital") != 0 {
		t.Fatal("extra accepted visit must cost one level")
	}
}
func TestLastRoundDoesNotCreditSeventhIncome(t *testing.T) {
	m := currentMatch(t)
	for m.Phase != "finished" {
		m.Advance(testNow)
	}
	if m.Country("norway").BalanceCents != 570000 || len(m.History) != 7 {
		t.Fatal("six rounds produce five incomes")
	}
}

func TestShieldPriceUsesPreUpgradeSecurityAndDestroyedLevelsReset(t *testing.T) {
	m := currentMatch(t)
	m.Phase = "headquarters"
	c := m.Country("norway")
	c.Plan.Upgrades = []string{"norway-3"}
	c.Plan.Shields = []string{"norway-3"}
	m.Advance(testNow)
	if c.BalanceCents != 120000 || c.Level("security") != 1 || !c.Cities[2].Shield {
		t.Fatal("same-round security upgrade must not discount existing purchases")
	}
	m.Advance(testNow)
	enemy := m.Country("germany")
	enemy.Bombs = 2
	enemy.Plan.Launches = []string{"norway-3", "norway-3"}
	m.Advance(testNow)
	if !c.Cities[2].Destroyed || c.Cities[2].Level != 0 || m.Prices(c)["shield"] != 30000 || c.Cities[0].Development != 79 {
		t.Fatal("destruction must reset bonus; shielded launch must also cause global damage")
	}
}
func TestFreePurchasesNotTrimmedAndOldUnversionedMatchKeepsRules(t *testing.T) {
	m := currentMatch(t)
	c := m.Country("norway")
	c.BalanceCents = 0
	c.NuclearRound = 1
	c.CityByRole("military").Level = 6
	c.CityByRole("security").Level = 6
	p := c.Plan
	p.Upgrades = []string{"norway-1"}
	p.Shields = []string{"norway-1"}
	p.Bombs = 1
	m.trimBudget(c, &p)
	if len(p.Upgrades) != 0 || len(p.Shields) != 1 || p.Bombs != 1 {
		t.Fatal("free orders should not be cancelled for overspending")
	}
	old := newMatch(t)
	old.RulesVersion = 0
	old.BaseIncomeCents = 0
	old.Advance(testNow)
	old.Country("norway").Plan.Upgrades = []string{"norway-1"}
	old.Advance(testNow)
	old.Advance(testNow)
	if old.Rules().Version != 1 || len(old.Country("norway").Cities) != 4 || old.Country("norway").BalanceCents != 182500 {
		t.Fatal("unversioned persisted games must retain original economics")
	}
}
