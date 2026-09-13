package game

import "time"

type Definition struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Cities []string `json:"cities"`
}

var Definitions = []Definition{
	{"australia", "Австралия", []string{"Канберра", "Сидней", "Мельбурн", "Перт"}},
	{"germany", "Германия", []string{"Берлин", "Мюнхен", "Гамбург", "Франкфурт"}},
	{"kazakhstan", "Казахстан", []string{"Астана", "Алматы", "Шымкент", "Актау"}},
	{"canada", "Канада", []string{"Оттава", "Ванкувер", "Монреаль", "Калгари"}},
	{"mexico", "Мексика", []string{"Мехико", "Гвадалахара", "Монтеррей", "Тихуана"}},
	{"norway", "Норвегия", []string{"Осло", "Берген", "Тронхейм", "Ставангер"}},
	{"russia", "Россия", []string{"Москва", "Санкт-Петербург", "Тюмень", "Екатеринбург"}},
	{"north-korea", "Северная Корея", []string{"Пхеньян", "Разон", "Чхонджин", "Нампо"}},
	{"saudi-arabia", "Саудовская Аравия", []string{"Эр-Рияд", "Джидда", "Медина", "Даммам"}},
	{"france", "Франция", []string{"Париж", "Лион", "Марсель", "Бордо"}},
}

type City struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Development int    `json:"development"`
	Destroyed   bool   `json:"destroyed"`
	Shield      bool   `json:"shield"`
}
type Donation struct {
	CountryID   string `json:"countryId"`
	AmountCents int64  `json:"amountCents"`
}
type Plan struct {
	Version   int        `json:"version"`
	Upgrades  []string   `json:"upgrades"`
	Shields   []string   `json:"shields"`
	Nuclear   bool       `json:"nuclear"`
	Bombs     int        `json:"bombs"`
	Ecology   bool       `json:"ecology"`
	Sanctions []string   `json:"sanctions"`
	Donations []Donation `json:"donations"`
	Launches  []string   `json:"launches"`
}

func EmptyPlan(version int) Plan {
	return Plan{Version: version, Upgrades: []string{}, Shields: []string{}, Sanctions: []string{}, Donations: []Donation{}, Launches: []string{}}
}

type Country struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Cities           []City   `json:"cities"`
	BalanceCents     int64    `json:"balanceCents"`
	IncomeCents      int64    `json:"incomeCents"`
	NuclearRound     int      `json:"nuclearRound"`
	Bombs            int      `json:"bombs"`
	Plan             Plan     `json:"plan"`
	SanctionedBy     []string `json:"sanctionedBy"`
	RetaliationRound int      `json:"retaliationRound"`
	Eliminated       bool     `json:"eliminated"`
	DoubleRound      int      `json:"doubleRound"`
}
type Event struct {
	Round     int    `json:"round"`
	Kind      string `json:"kind"`
	CountryID string `json:"countryId,omitempty"`
	TargetID  string `json:"targetId,omitempty"`
	CityID    string `json:"cityId,omitempty"`
	Amount    int64  `json:"amount,omitempty"`
	Text      string `json:"text"`
}
type RoundScore struct {
	Round   int                `json:"round"`
	Ecology int                `json:"ecology"`
	Scores  map[string]float64 `json:"scores"`
}
type Meeting struct {
	ID       string    `json:"id"`
	Round    int       `json:"round"`
	From     string    `json:"from"`
	To       string    `json:"to"`
	Status   string    `json:"status"`
	Double   bool      `json:"double"`
	Messages []Message `json:"messages"`
}
type Message struct {
	CountryID string `json:"countryId"`
	Name      string `json:"name"`
	Text      string `json:"text"`
	At        int64  `json:"at"`
}
type Match struct {
	ID           string       `json:"id"`
	Revision     int          `json:"revision"`
	Round        int          `json:"round"`
	Phase        string       `json:"phase"`
	PhaseSeconds int          `json:"phaseSeconds"`
	Deadline     int64        `json:"deadline"`
	Paused       bool         `json:"paused"`
	Remaining    int64        `json:"remaining"`
	Ecology      int          `json:"ecology"`
	Countries    []Country    `json:"countries"`
	Events       []Event      `json:"events"`
	History      []RoundScore `json:"history"`
	Meetings     []Meeting    `json:"meetings"`
	Requests     []string     `json:"requests"`
}
type Actor struct {
	ID, Name, CountryID string
	Host                bool
}

func New(id string, ids []string, seconds int, now time.Time) (*Match, error) {
	if len(ids) < 2 || len(ids) > 10 {
		return nil, invalid("Нужны от 2 до 10 стран")
	}
	if seconds < 10 || seconds > 3600 {
		return nil, invalid("Длительность фазы: от 10 до 3600 секунд")
	}
	m := &Match{ID: id, Revision: 1, Round: 1, Phase: "council", PhaseSeconds: seconds, Deadline: now.Add(time.Duration(seconds) * time.Second).UnixMilli(), Ecology: 100, Events: []Event{}, Meetings: []Meeting{}, History: []RoundScore{}, Requests: []string{}}
	seen := map[string]bool{}
	for _, id := range ids {
		if seen[id] {
			return nil, invalid("Страна повторяется")
		}
		seen[id] = true
		var def *Definition
		for i := range Definitions {
			if Definitions[i].ID == id {
				def = &Definitions[i]
			}
		}
		if def == nil {
			return nil, invalid("Неизвестная страна")
		}
		c := Country{ID: id, Name: def.Name, BalanceCents: 100000, Plan: EmptyPlan(0), SanctionedBy: []string{}}
		for i, name := range def.Cities {
			c.Cities = append(c.Cities, City{ID: id + "-" + string(rune('1'+i)), Name: name, Development: []int{100, 80, 60, 60}[i]})
		}
		m.Countries = append(m.Countries, c)
	}
	m.recordScore(0)
	return m, nil
}
func (m *Match) Country(id string) *Country {
	for i := range m.Countries {
		if m.Countries[i].ID == id {
			return &m.Countries[i]
		}
	}
	return nil
}
func (m *Match) City(id string) (*Country, *City) {
	for i := range m.Countries {
		for j := range m.Countries[i].Cities {
			if m.Countries[i].Cities[j].ID == id {
				return &m.Countries[i], &m.Countries[i].Cities[j]
			}
		}
	}
	return nil, nil
}
func (c *Country) Score() float64 {
	sum := 0
	for _, city := range c.Cities {
		sum += city.Development
	}
	return float64(sum) / 4
}
func (c *Country) Alive() bool {
	for _, city := range c.Cities {
		if !city.Destroyed {
			return true
		}
	}
	return false
}
func (m *Match) recordScore(round int) {
	s := RoundScore{Round: round, Ecology: m.Ecology, Scores: map[string]float64{}}
	for _, c := range m.Countries {
		s.Scores[c.ID] = c.Score()
	}
	m.History = append(m.History, s)
}
