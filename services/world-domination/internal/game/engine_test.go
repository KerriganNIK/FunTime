package game

import (
	"encoding/json"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	"testing"
	"time"
)

var testNow = time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)

func newMatch(t *testing.T) *Match {
	t.Helper()
	m, err := newMatchWithRules("ABC234", []string{"norway", "germany"}, 720, testNow, 1)
	if err != nil {
		t.Fatal(err)
	}
	return m
}
func execute(t *testing.T, m *Match, a Actor, kind string, v any) error {
	t.Helper()
	b, _ := json.Marshal(v)
	return m.Command(a, kind, b, testNow)
}
func TestEconomyBoundariesAndFirstBudget(t *testing.T) {
	for _, tc := range []struct{ n, want int }{{60, 20}, {99, 20}, {100, 25}, {199, 25}, {200, 100}, {250, 100}} {
		if got := Upgrade(tc.n); got != tc.want {
			t.Errorf("upgrade(%d)=%d", tc.n, got)
		}
	}
	m := newMatch(t)
	m.Advance(testNow)
	c := m.Country("norway")
	if c.BalanceCents != 100000 || c.IncomeCents != 0 {
		t.Fatal("first budget must be 1000")
	}
	if got := Income(c, 100); got != 90000 {
		t.Fatalf("income=%d", got)
	}
	c.SanctionedBy = []string{"a", "b"}
	if got := Income(c, 120); got != 75600 {
		t.Fatalf("stacked ecology income=%d", got)
	}
	c.SanctionedBy = make([]string, 7)
	if Income(c, 100) != 0 {
		t.Fatal("negative sanction multiplier")
	}
	if Income(c, 0) != 0 {
		t.Fatal("zero ecology")
	}
}
func TestSimultaneousShieldsAndGlobalDamage(t *testing.T) {
	for _, hits := range []int{1, 2} {
		m := newMatch(t)
		m.Phase = "headquarters"
		a := m.Country("norway")
		b := m.Country("germany")
		a.Bombs = hits
		a.Plan.Launches = []string{"germany-1"}
		if hits == 2 {
			a.Plan.Launches = append(a.Plan.Launches, "germany-1")
		}
		b.Plan.Shields = []string{"germany-1"}
		b.Bombs = 1
		b.Plan.Launches = []string{"norway-1"}
		m.Advance(testNow)
		if !a.Cities[0].Destroyed {
			t.Fatal("destroyed attacker must still fire")
		}
		if b.Cities[0].Shield {
			t.Fatal("shield consumed")
		}
		if b.Cities[0].Destroyed != (hits == 2) {
			t.Fatalf("hits=%d destroyed=%v", hits, b.Cities[0].Destroyed)
		}
		if b.Cities[1].Development != 80-(hits+1)*3 {
			t.Fatal("global penalty must include every launch, including absorbed")
		}
		if a.Bombs != 0 || b.Bombs != 0 {
			t.Fatal("ordnance not consumed")
		}
	}
}
func TestNuclearMaturityEcologyAndTransfers(t *testing.T) {
	m := newMatch(t)
	m.Advance(testNow)
	a := m.Country("norway")
	actor := Actor{CountryID: a.ID}
	p := EmptyPlan(0)
	p.Nuclear = true
	p.Bombs = 1
	if err := execute(t, m, actor, "plan", p); err == nil {
		t.Fatal("bombs allowed too early")
	}
	p.Bombs = 0
	p.Ecology = true
	p.Donations = []Donation{{CountryID: "germany", AmountCents: 10000}}
	p.Sanctions = []string{"germany"}
	if err := execute(t, m, actor, "plan", p); err != nil {
		t.Fatal(err)
	}
	m.Advance(testNow)
	if a.BalanceCents != 20000 || m.Country("germany").BalanceCents != 110000 || m.Ecology != 120 || a.NuclearRound != 2 {
		t.Fatalf("settlement: %+v", m)
	}
	m.Advance(testNow)
	if m.Country("germany").IncomeCents != 91800 {
		t.Fatal("sanctions must apply next round")
	}
	p = EmptyPlan(a.Plan.Version)
	p.Bombs = 1
	p.Launches = []string{"germany-1"}
	if err := execute(t, m, actor, "plan", p); err != nil {
		t.Fatal(err)
	}
	m.Advance(testNow)
	if m.Ecology != 116 {
		t.Fatal("production penalty")
	}
	if len(m.Country("germany").SanctionedBy) != 0 {
		t.Fatal("sanctions not expired")
	}
}
func TestOverspendIsRandomlyTrimmedAndIncomingFundsNotSpendable(t *testing.T) {
	for range 30 {
		m := newMatch(t)
		m.Phase = "headquarters"
		a := m.Country("norway")
		b := m.Country("germany")
		a.BalanceCents = 0
		a.Plan.Upgrades = []string{"norway-1"}
		b.Plan.Donations = []Donation{{CountryID: a.ID, AmountCents: 20000}}
		b.Plan.Nuclear = true
		b.Plan.Ecology = true
		b.Plan.Shields = []string{"germany-1", "germany-2"}
		m.Advance(testNow)
		if a.Cities[0].Development != 100 {
			t.Fatal("incoming donation spent in same settlement")
		}
		if a.BalanceCents < 0 || b.BalanceCents < 0 {
			t.Fatal("negative budget")
		}
		n := 0
		for _, e := range m.Events {
			if e.Kind == "cancelled" {
				n++
			}
		}
		if n == 0 {
			t.Fatal("missing cancellation audit")
		}
	}
}
func TestRetaliationAndSixRoundEnd(t *testing.T) {
	m := newMatch(t)
	m.Phase = "headquarters"
	a := m.Country("norway")
	b := m.Country("germany")
	a.Bombs = 4
	a.Plan.Launches = []string{"germany-1", "germany-2", "germany-3", "germany-4"}
	b.Bombs = 1
	m.Advance(testNow)
	if b.RetaliationRound != 2 || b.Eliminated || b.Score() != 0 {
		t.Fatal("retaliation not queued")
	}
	m.Advance(testNow)
	if b.IncomeCents != 0 {
		t.Fatal("destroyed country earned income")
	}
	p := EmptyPlan(b.Plan.Version)
	p.Ecology = true
	if err := execute(t, m, Actor{CountryID: b.ID}, "plan", p); err == nil {
		t.Fatal("retaliation purchases allowed")
	}
	p.Ecology = false
	p.Launches = []string{"norway-1"}
	if err := execute(t, m, Actor{CountryID: b.ID}, "plan", p); err != nil {
		t.Fatal(err)
	}
	m.Advance(testNow)
	if !b.Eliminated || !a.Cities[0].Destroyed {
		t.Fatal("retaliation failed")
	}
	for m.Phase != "finished" {
		m.Advance(testNow)
	}
	if m.Round != 6 || len(m.History) != 7 {
		t.Fatal("must end after six rounds")
	}
	m.Advance(testNow)
	if m.Round != 6 {
		t.Fatal("seventh round created")
	}
}
func TestPublicAndOpponentProjectionsNeverContainSecrets(t *testing.T) {
	m := newMatch(t)
	m.Country("norway").Bombs = 17
	m.Events = []Event{{Kind: "attack", CountryID: "norway", TargetID: "germany"}, {Kind: "donation", CountryID: "norway", TargetID: "germany", Amount: 77700}}
	for _, a := range []Actor{{}, {CountryID: "germany"}} {
		b, _ := json.Marshal(m.View(a, testNow))
		var v struct {
			Countries []map[string]any
			Events    []Event
		}
		_ = json.Unmarshal(b, &v)
		for _, c := range v.Countries {
			if c["id"] == a.CountryID {
				continue
			}
			for _, secret := range []string{"balanceCents", "plan", "bombs", "nuclearRound", "sanctionedBy"} {
				if _, ok := c[secret]; ok {
					t.Fatalf("secret %s leaked", secret)
				}
			}
		}
		for _, e := range v.Events {
			if e.CountryID == "norway" {
				t.Fatal("attacker or donor exposed early")
			}
		}
	}
	m.Phase = "finished"
	v := m.View(Actor{CountryID: "germany"}, testNow)
	b, _ := json.Marshal(v["events"])
	if !strings.Contains(string(b), `"kind":"attack"`) {
		t.Fatal("final attacks missing")
	}
	var events []Event
	_ = json.Unmarshal(b, &events)
	for _, e := range events {
		if e.Kind == "donation" && e.CountryID != "" {
			t.Fatal("donor revealed at final")
		}
	}
}
func TestPlanConflictsHostPermissionsAndPause(t *testing.T) {
	m := newMatch(t)
	if status.Code(execute(t, m, Actor{}, "advance", map[string]string{"expectedPhase": "1:council"})) != codes.PermissionDenied {
		t.Fatal("player advanced phase")
	}
	m.Advance(testNow)
	p := EmptyPlan(0)
	if err := execute(t, m, Actor{CountryID: "norway"}, "plan", p); err != nil {
		t.Fatal(err)
	}
	if status.Code(execute(t, m, Actor{CountryID: "norway"}, "plan", p)) != codes.Aborted {
		t.Fatal("stale plan overwrote teammate")
	}
	if err := execute(t, m, Actor{Host: true}, "pause", map[string]string{"expectedPhase": "1:headquarters"}); err != nil {
		t.Fatal(err)
	}
	if !m.Paused || m.Remaining != 720000 {
		t.Fatal("pause timer wrong")
	}
	if err := execute(t, m, Actor{CountryID: "norway"}, "plan", EmptyPlan(1)); err == nil {
		t.Fatal("orders accepted during pause")
	}
}
func TestDiplomacyRejectedVisitDoesNotConsumeQuota(t *testing.T) {
	m := newMatch(t)
	m.Phase = "headquarters"
	a, b := Actor{CountryID: "norway", Name: "A"}, Actor{CountryID: "germany", Name: "B"}
	request := map[string]any{"countryId": "germany"}
	if err := execute(t, m, a, "meeting.request", request); err != nil {
		t.Fatal(err)
	}
	if err := execute(t, m, b, "meeting.respond", map[string]any{"id": "meeting-1", "accept": false}); err != nil {
		t.Fatal(err)
	}
	if err := execute(t, m, a, "meeting.request", request); err != nil {
		t.Fatal(err)
	}
	if err := execute(t, m, b, "meeting.respond", map[string]any{"id": "meeting-2", "accept": true}); err != nil {
		t.Fatal(err)
	}
	if err := execute(t, m, a, "meeting.message", map[string]string{"id": "meeting-2", "text": "Союз?"}); err != nil {
		t.Fatal(err)
	}
	if len(m.View(Actor{}, testNow)["meetings"].([]Meeting)) != 0 {
		t.Fatal("diplomacy exposed to projector")
	}
	if err := execute(t, m, a, "meeting.request", request); err == nil {
		t.Fatal("repeat accepted destination allowed")
	}
}
