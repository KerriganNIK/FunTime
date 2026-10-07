package game

const CurrentRulesVersion = 2

type Rules struct {
	Version              int   `json:"version"`
	StartingBalanceCents int64 `json:"startingBalanceCents"`
	UpgradeCostCents     int64 `json:"upgradeCostCents"`
	ShieldBaseCostCents  int64 `json:"shieldBaseCostCents"`
	BombBaseCostCents    int64 `json:"bombBaseCostCents"`
	NuclearCostCents     int64 `json:"nuclearCostCents"`
	EcologyCostCents     int64 `json:"ecologyCostCents"`
	LevelDiscountCents   int64 `json:"levelDiscountCents"`
	TourismIncomeCents   int64 `json:"tourismIncomeCents"`
	CityCount            int   `json:"cityCount"`
}

func (m *Match) Rules() Rules {
	if m.RulesVersion < 2 {
		return Rules{Version: 1, StartingBalanceCents: 100000, UpgradeCostCents: 15000, ShieldBaseCostCents: 30000, BombBaseCostCents: 30000, NuclearCostCents: 50000, EcologyCostCents: 20000, CityCount: 4}
	}
	return Rules{Version: 2, StartingBalanceCents: 70000, UpgradeCostCents: 20000, ShieldBaseCostCents: 30000, BombBaseCostCents: 30000, NuclearCostCents: 50000, EcologyCostCents: 20000, LevelDiscountCents: 5000, TourismIncomeCents: 7500, CityCount: 5}
}

func (c *Country) CityByRole(role string) *City {
	for i := range c.Cities {
		if c.Cities[i].Role == role {
			return &c.Cities[i]
		}
	}
	return nil
}
func (c *Country) Level(role string) int {
	city := c.CityByRole(role)
	if city == nil || city.Destroyed {
		return 0
	}
	return max(0, city.Level)
}
func (m *Match) Prices(c *Country) map[string]int64 {
	rules := m.Rules()
	return map[string]int64{"upgrade": rules.UpgradeCostCents, "shield": max(0, rules.ShieldBaseCostCents-int64(c.Level("security"))*rules.LevelDiscountCents), "bomb": max(0, rules.BombBaseCostCents-int64(c.Level("military"))*rules.LevelDiscountCents), "nuclear": rules.NuclearCostCents, "ecology": rules.EcologyCostCents}
}
func (m *Match) PlanCost(c *Country, p Plan) int64 {
	prices := m.Prices(c)
	cost := int64(len(p.Upgrades))*prices["upgrade"] + int64(len(p.Shields))*prices["shield"] + int64(p.Bombs)*prices["bomb"]
	if p.Nuclear {
		cost += prices["nuclear"]
	}
	if p.Ecology {
		cost += prices["ecology"]
	}
	for _, d := range p.Donations {
		cost += d.AmountCents
	}
	return cost
}
func (m *Match) NextIncome(c *Country) int64 {
	if !c.Alive() || c.Eliminated {
		return 0
	}
	if m.RulesVersion < 2 {
		return Income(c, m.Ecology)
	}
	// The sheet leaves sanction loss manual. Retain the agreed 15% of income per country.
	income := max(0, m.BaseIncomeCents)*int64(max(0, 100-m.Pollution)) + int64(c.Level("tourism"))*m.Rules().TourismIncomeCents*100
	return (income*int64(max(0, 100-15*len(c.SanctionedBy))) + 5000) / 10000
}

var cityRoles = []string{"capital", "military", "security", "intelligence", "tourism"}

func currentDefinitions() []Definition {
	return []Definition{
		{ID: "australia", Name: "Австралия", Cities: []string{"Канберра", "Аделаида", "Сидней", "Перт", "Брисбен"}, StartingDevelopment: []int{80, 60, 70, 90, 90}},
		{ID: "germany", Name: "Германия", Cities: []string{"Берлин", "Штутгарт", "Франкфурт", "Гамбург", "Мюнхен"}, StartingDevelopment: []int{70, 70, 90, 70, 70}},
		{ID: "kazakhstan", Name: "Казахстан", Cities: []string{"Астана", "Караганда", "Шымкент", "Актау", "Алматы"}, StartingDevelopment: []int{75, 80, 65, 100, 65}},
		{ID: "canada", Name: "Канада", Cities: []string{"Оттава", "Эдмонтон", "Калгари", "Торонто", "Ванкувер"}, StartingDevelopment: []int{80, 65, 100, 70, 80}},
		{ID: "mexico", Name: "Мексика", Cities: []string{"Мехико", "Монтеррей", "Гвадалахара", "Тихуана", "Канкун"}, StartingDevelopment: []int{70, 65, 60, 65, 100}},
		{ID: "norway", Name: "Норвегия", Cities: []string{"Осло", "Берген", "Тронхейм", "Тромсе", "Ставангер"}, StartingDevelopment: []int{85, 60, 100, 60, 70}},
		{ID: "russia", Name: "Россия", Cities: []string{"Москва", "Челябинск", "Тюмень", "Калининград", "Сочи"}, StartingDevelopment: []int{80, 90, 70, 65, 60}},
		{ID: "saudi-arabia", Name: "Саудовская Аравия", Cities: []string{"Эр-Рияд", "Табук", "Таиф", "Даммам", "Джидда"}, StartingDevelopment: []int{100, 70, 80, 60, 70}},
		{ID: "north-korea", Name: "Северная Корея", Cities: []string{"Пхеньян", "Хамхын", "Кэсон", "Чхонджин", "Вонсан"}, StartingDevelopment: []int{65, 100, 60, 70, 60}},
		{ID: "france", Name: "Франция", Cities: []string{"Париж", "Бордо", "Лион", "Тулуза", "Ницца"}, StartingDevelopment: []int{80, 80, 80, 80, 80}},
	}
}
