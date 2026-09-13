package game

import "time"

// Construct projections explicitly. Never serialize the persisted match for players.
func (m *Match) View(a Actor, now time.Time) map[string]any {
	countries := []any{}
	for _, c := range m.Countries {
		cities := []any{}
		private := a.Host || a.CountryID == c.ID
		for _, city := range c.Cities {
			v := map[string]any{"id": city.ID, "name": city.Name, "development": city.Development, "destroyed": city.Destroyed}
			if private {
				v["shield"] = city.Shield
			}
			cities = append(cities, v)
		}
		v := map[string]any{"id": c.ID, "name": c.Name, "cities": cities, "score": c.Score(), "eliminated": c.Eliminated, "retaliationRound": c.RetaliationRound}
		if private {
			v["balanceCents"] = c.BalanceCents
			v["incomeCents"] = c.IncomeCents
			v["bombs"] = c.Bombs
			v["nuclearRound"] = c.NuclearRound
			v["plan"] = c.Plan
			v["planCostCents"] = Cost(c.Plan)
			v["sanctionedBy"] = c.SanctionedBy
			v["doubleRound"] = c.DoubleRound
		}
		countries = append(countries, v)
	}
	events := []Event{}
	for _, e := range m.Events {
		if a.Host || a.CountryID != "" && e.CountryID == a.CountryID {
			events = append(events, e)
			continue
		}
		switch e.Kind {
		case "attack", "upgrade", "production", "sanction":
			if m.Phase == "finished" || e.Kind == "sanction" && e.TargetID == a.CountryID {
				events = append(events, e)
			}
		case "donation":
			if a.CountryID != "" && e.TargetID == a.CountryID {
				e.CountryID = ""
				e.Text = "Анонимное пожертвование"
				events = append(events, e)
			}
		}
	}
	meetings := []Meeting{}
	for _, mt := range m.Meetings {
		if a.Host || a.CountryID != "" && (mt.From == a.CountryID || mt.To == a.CountryID) {
			meetings = append(meetings, mt)
		}
	}
	v := map[string]any{"revision": m.Revision, "round": m.Round, "phase": m.Phase, "phaseSeconds": m.PhaseSeconds, "deadline": m.Deadline, "paused": m.Paused, "remaining": m.Remaining, "serverTime": now.UnixMilli(), "ecology": m.Ecology, "countries": countries, "events": events, "history": m.History, "meetings": meetings}
	if a.Host {
		v["formulas"] = map[string]int{"ecologyImprovement": 20, "bombProductionEcology": 4, "launchGlobalDamage": 3}
	}
	return v
}
