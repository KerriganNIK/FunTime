package game

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math/big"
	"slices"
	"strings"
	"time"
)

func invalid(s string) error { return status.Error(codes.InvalidArgument, s) }
func denied() error {
	return status.Error(codes.PermissionDenied, "Это действие недоступно")
}
func Upgrade(development int) int {
	if development < 100 {
		return 20
	}
	if development < 200 {
		return 25
	}
	return 100
}
func Income(c *Country, ecology int) int64 {
	sum := 0
	for _, city := range c.Cities {
		sum += city.Development
	}
	return (int64(sum)*3*int64(ecology)*int64(max(0, 100-15*len(c.SanctionedBy))) + 50) / 100
}
func (m *Match) validate(c *Country, p Plan) error {
	if c.Eliminated {
		return denied()
	}
	if len(p.Upgrades) > len(c.Cities) || len(p.Shields) > len(c.Cities) || p.Bombs < 0 || p.Bombs > 1000 || len(p.Launches) > 1000 || len(p.Donations) > 10 || len(p.Sanctions) > 9 || len(p.Spies) > 9 {
		return invalid("Слишком много действий")
	}
	if c.RetaliationRound > 0 && (len(p.Upgrades)+len(p.Shields)+p.Bombs+len(p.Donations)+len(p.Sanctions)+len(p.Spies) > 0 || p.Nuclear || p.Ecology) {
		return invalid("Акт возмездия: доступны только оставшиеся бомбы")
	}
	for _, list := range [][]string{p.Upgrades, p.Shields, p.Sanctions, p.Spies} {
		seen := map[string]bool{}
		for _, id := range list {
			if seen[id] {
				return invalid("Действие повторяется")
			}
			seen[id] = true
		}
	}
	for _, id := range append(slices.Clone(p.Upgrades), p.Shields...) {
		owner, city := m.City(id)
		if city == nil || owner.ID != c.ID || city.Destroyed {
			return invalid("Город недоступен")
		}
	}
	for _, id := range p.Shields {
		_, city := m.City(id)
		if city.Shield {
			return invalid("На городе уже установлен щит")
		}
	}
	if p.Nuclear && c.NuclearRound != 0 {
		return invalid("Ядерная программа уже запущена")
	}
	if p.Bombs > 0 && (c.NuclearRound == 0 || m.Round < c.NuclearRound) {
		return invalid("Покупка бомб доступна со следующего раунда после запуска программы")
	}
	if len(p.Launches) > c.Bombs+p.Bombs {
		return invalid("Недостаточно бомб")
	}
	for _, id := range p.Launches {
		_, city := m.City(id)
		if city == nil || city.Destroyed {
			return invalid("Цель недоступна")
		}
	}
	for _, id := range p.Sanctions {
		other := m.Country(id)
		if other == nil || other.ID == c.ID || !other.Alive() {
			return invalid("Неверная цель санкций")
		}
	}
	seen := map[string]bool{}
	spyLevel := c.Level("intelligence")
	if city := c.CityByRole("intelligence"); city != nil && slices.Contains(p.Upgrades, city.ID) {
		spyLevel++
	}
	if len(p.Spies) > 0 && (m.RulesVersion < 2 || len(p.Spies) > spyLevel) {
		return invalid("Недостаточно уровней города разведки")
	}
	for _, id := range p.Spies {
		other := m.Country(id)
		if other == nil || other.ID == c.ID || !other.Alive() {
			return invalid("Неверная цель разведки")
		}
	}
	for _, d := range p.Donations {
		other := m.Country(d.CountryID)
		if other == nil || other.ID == c.ID || !other.Alive() || d.AmountCents <= 0 || d.AmountCents > 100000000000 || seen[d.CountryID] {
			return invalid("Неверное пожертвование")
		}
		seen[d.CountryID] = true
	}
	return nil
}
func decode(b []byte, v any) error {
	if len(b) == 0 {
		b = []byte(`{}`)
	}
	if err := json.Unmarshal(b, v); err != nil {
		return invalid("Некорректные данные")
	}
	return nil
}
func (m *Match) Command(a Actor, kind string, payload []byte, now time.Time) error {
	if kind == "income.adjust" {
		if !a.Host {
			return denied()
		}
		if m.RulesVersion < 2 || m.Phase == "finished" {
			return invalid("Изменение дохода недоступно")
		}
		var p struct {
			BaseIncomeCents int64  `json:"baseIncomeCents"`
			ExpectedPhase   string `json:"expectedPhase"`
		}
		if err := decode(payload, &p); err != nil {
			return err
		}
		if p.ExpectedPhase != fmt.Sprintf("%d:%s", m.Round, m.Phase) {
			return status.Error(codes.Aborted, "Фаза уже изменилась")
		}
		if p.BaseIncomeCents < 0 || p.BaseIncomeCents > 100000000 {
			return invalid("Базовый доход: от 0 до 1 000 000 монет")
		}
		m.BaseIncomeCents = p.BaseIncomeCents
		return nil
	}
	if kind == "advance" || kind == "pause" || kind == "resume" {
		if !a.Host {
			return denied()
		}
		if m.Phase == "finished" {
			return invalid("Игра завершена")
		}
		var p struct {
			ExpectedPhase string `json:"expectedPhase"`
		}
		if err := decode(payload, &p); err != nil {
			return err
		}
		if p.ExpectedPhase != fmt.Sprintf("%d:%s", m.Round, m.Phase) {
			return status.Error(codes.Aborted, "Фаза уже изменилась")
		}
		switch kind {
		case "advance":
			m.Advance(now)
		case "pause":
			if !m.Paused {
				m.Remaining = max(0, m.Deadline-now.UnixMilli())
				m.Paused = true
			}
		case "resume":
			if m.Paused {
				m.Deadline = now.UnixMilli() + m.Remaining
				m.Paused = false
			}
		}
		return nil
	}
	if m.Phase != "headquarters" || m.Paused {
		return invalid("Действия доступны во время работы в штабе")
	}
	c := m.Country(a.CountryID)
	if c == nil || c.Eliminated {
		return denied()
	}
	switch kind {
	case "plan":
		p := EmptyPlan(0)
		if err := decode(payload, &p); err != nil {
			return err
		}
		if p.Upgrades == nil {
			p.Upgrades = []string{}
		}
		if p.Shields == nil {
			p.Shields = []string{}
		}
		if p.Launches == nil {
			p.Launches = []string{}
		}
		if p.Sanctions == nil {
			p.Sanctions = []string{}
		}
		if p.Donations == nil {
			p.Donations = []Donation{}
		}
		if p.Version != c.Plan.Version {
			return status.Error(codes.Aborted, "План изменил другой участник. Загрузите его изменения")
		}
		if err := m.validate(c, p); err != nil {
			return err
		}
		p.Version++
		c.Plan = p
		return nil
	case "meeting.request", "meeting.respond", "meeting.message":
		return m.diplomacy(a, c, kind, payload, now)
	default:
		return invalid("Неизвестное действие")
	}
}
func randomIndex(n int) int {
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		panic(err)
	}
	return int(v.Int64())
}
func (m *Match) trimBudget(c *Country, p *Plan) {
	prices := m.Prices(c)
	for m.PlanCost(c, *p) > c.BalanceCents {
		actions := []func(){}
		for i := range p.Upgrades {
			actions = append(actions, func() { p.Upgrades = append(p.Upgrades[:i], p.Upgrades[i+1:]...) })
		}
		for i := range p.Shields {
			if prices["shield"] > 0 {
				actions = append(actions, func() { p.Shields = append(p.Shields[:i], p.Shields[i+1:]...) })
			}
		}
		for range p.Bombs {
			if prices["bomb"] > 0 {
				actions = append(actions, func() { p.Bombs-- })
			}
		}
		if p.Nuclear {
			actions = append(actions, func() { p.Nuclear = false })
		}
		if p.Ecology {
			actions = append(actions, func() { p.Ecology = false })
		}
		for i := range p.Donations {
			actions = append(actions, func() { p.Donations = append(p.Donations[:i], p.Donations[i+1:]...) })
		}
		if len(actions) == 0 {
			break
		}
		actions[randomIndex(len(actions))]()
		m.Events = append(m.Events, Event{Round: m.Round, Kind: "cancelled", CountryID: c.ID, Text: "Случайная отмена покупки из-за превышения бюджета"})
	}
	for len(p.Launches) > c.Bombs+p.Bombs {
		i := randomIndex(len(p.Launches))
		p.Launches = append(p.Launches[:i], p.Launches[i+1:]...)
		m.Events = append(m.Events, Event{Round: m.Round, Kind: "cancelled", CountryID: c.ID, Text: "Запуск отменён: покупка бомбы не вошла в бюджет"})
	}
}
func (m *Match) settle() {
	plans := map[string]Plan{}
	launches := []Event{}
	produced, ecology, programs := 0, 0, 0
	for i := range m.Countries {
		c := &m.Countries[i]
		p := c.Plan
		if c.Eliminated {
			p = EmptyPlan(p.Version)
		}
		m.trimBudget(c, &p)
		plans[c.ID] = p
		c.BalanceCents -= m.PlanCost(c, p)
		for _, id := range p.Upgrades {
			_, city := m.City(id)
			city.Development += Upgrade(city.Development)
			if m.RulesVersion >= 2 {
				city.Level++
			}
			m.Events = append(m.Events, Event{Round: m.Round, Kind: "upgrade", CountryID: c.ID, CityID: id, Text: "Развитие города"})
		}
		for _, id := range p.Shields {
			_, city := m.City(id)
			city.Shield = true
		}
		if p.Nuclear {
			c.NuclearRound = m.Round + 1
			programs++
		}
		c.Bombs += p.Bombs
		produced += p.Bombs
		if p.Bombs > 0 {
			m.Events = append(m.Events, Event{Round: m.Round, Kind: "production", CountryID: c.ID, Amount: int64(p.Bombs), Text: "Создание бомб"})
		}
		if p.Ecology {
			ecology++
			m.Events = append(m.Events, Event{Round: m.Round, Kind: "ecology", CountryID: c.ID, Text: "Улучшение экологии"})
		}
		if m.RulesVersion >= 2 {
			if len(p.Spies) > c.Level("intelligence") {
				p.Spies = p.Spies[:c.Level("intelligence")]
				m.Events = append(m.Events, Event{Round: m.Round, Kind: "cancelled", CountryID: c.ID, Text: "Разведка отменена: улучшение города не вошло в бюджет"})
			}
			if city := c.CityByRole("intelligence"); city != nil {
				city.Level -= len(p.Spies)
			}
			plans[c.ID] = p
		}
		for _, id := range p.Launches {
			owner, _ := m.City(id)
			launches = append(launches, Event{Round: m.Round, Kind: "attack", CountryID: c.ID, TargetID: owner.ID, CityID: id, Text: "Ядерный удар"})
			c.Bombs--
		}
	}
	for i := range m.Countries {
		m.Countries[i].SanctionedBy = []string{}
	}
	for _, c := range m.Countries {
		p := plans[c.ID]
		for _, d := range p.Donations {
			m.Country(d.CountryID).BalanceCents += d.AmountCents
			m.Events = append(m.Events, Event{Round: m.Round, Kind: "donation", CountryID: c.ID, TargetID: d.CountryID, Amount: d.AmountCents, Text: "Пожертвование"})
		}
		for _, id := range p.Sanctions {
			m.Country(id).SanctionedBy = append(m.Country(id).SanctionedBy, c.ID)
			m.Events = append(m.Events, Event{Round: m.Round, Kind: "sanction", CountryID: c.ID, TargetID: id, Text: "Санкции на следующий раунд"})
		}
	}
	// All orders exist before damage: a destroyed country still fires this round.
	for _, hit := range launches {
		_, city := m.City(hit.CityID)
		if city.Shield {
			city.Shield = false
			hit.Text = "Щит поглотил удар"
		} else {
			city.Destroyed = true
			city.Development = 0
			hit.Text = "Удар по городу без щита"
		}
		m.Events = append(m.Events, hit)
	}
	if m.RulesVersion >= 2 {
		m.Pollution = max(0, m.Pollution+programs*6+produced*3-ecology*20)
		m.Ecology = max(0, 100-m.Pollution)
	} else {
		m.Ecology = max(0, m.Ecology+ecology*20-produced*4)
	}
	for i := range m.Countries {
		c := &m.Countries[i]
		for j := range c.Cities {
			city := &c.Cities[j]
			city.Development = max(0, city.Development-len(launches)*3)
			if city.Development == 0 {
				city.Destroyed = true
				city.Shield = false
				city.Level = 0
			}
		}
		if c.RetaliationRound > 0 && c.RetaliationRound <= m.Round {
			c.Eliminated = true
		}
		if !c.Alive() && c.RetaliationRound == 0 {
			c.RetaliationRound = m.Round + 1
		}
		c.Plan = EmptyPlan(c.Plan.Version + 1)
	}
	if m.RulesVersion >= 2 {
		for i := range m.Countries {
			c := &m.Countries[i]
			c.IncomeCents = 0
			if m.Round < 6 {
				c.IncomeCents = m.NextIncome(c)
				c.BalanceCents += c.IncomeCents
			}
		}
		for i := range m.Countries {
			c := &m.Countries[i]
			for _, id := range plans[c.ID].Spies {
				other := m.Country(id)
				c.Intelligence = append(c.Intelligence, IntelligenceReport{Round: m.Round, CountryID: id, CountryName: other.Name, BalanceCents: other.BalanceCents, Bombs: other.Bombs, NuclearRound: other.NuclearRound, Cities: slices.Clone(other.Cities)})
				m.Events = append(m.Events, Event{Round: m.Round, Kind: "spy", CountryID: c.ID, TargetID: id, Text: "Получен разведывательный отчёт"})
			}
		}
	}
	m.recordScore(m.Round)
}
func (m *Match) Advance(now time.Time) {
	if m.Phase == "finished" {
		return
	}
	if m.Phase == "council" {
		m.Phase = "headquarters"
		for i := range m.Countries {
			c := &m.Countries[i]
			if m.RulesVersion >= 2 {
				continue
			}
			c.IncomeCents = 0
			if m.Round > 1 && c.Alive() && !c.Eliminated {
				c.IncomeCents = Income(c, m.Ecology)
				c.BalanceCents += c.IncomeCents
			}
		}
	} else {
		m.settle()
		if m.Round == 6 {
			m.Phase = "finished"
		} else {
			m.Round++
			m.Phase = "council"
		}
	}
	m.Paused = false
	m.Remaining = 0
	m.Deadline = now.Add(time.Duration(m.PhaseSeconds) * time.Second).UnixMilli()
	if m.Phase == "finished" {
		m.Deadline = 0
	}
}
func (m *Match) diplomacy(a Actor, c *Country, kind string, b []byte, now time.Time) error {
	if !c.Alive() {
		return denied()
	}
	var p struct {
		CountryID string `json:"countryId"`
		ID        string `json:"id"`
		Accept    bool   `json:"accept"`
		Double    bool   `json:"double"`
		Text      string `json:"text"`
	}
	if err := decode(b, &p); err != nil {
		return err
	}
	accepted := func(from string) int {
		n := 0
		for _, mt := range m.Meetings {
			if mt.Round == m.Round && mt.From == from && mt.Status == "accepted" {
				n++
			}
		}
		return n
	}
	if kind == "meeting.request" {
		other := m.Country(p.CountryID)
		if other == nil || other.ID == c.ID || !other.Alive() {
			return invalid("Страна недоступна")
		}
		if len(m.Meetings) > 500 {
			return invalid("Лимит переговоров достигнут")
		}
		for _, mt := range m.Meetings {
			if mt.Round == m.Round && mt.From == c.ID && (mt.Status == "pending" || mt.To == other.ID && mt.Status == "accepted") {
				return invalid("Дождитесь ответа или выберите другую страну")
			}
		}
		n := accepted(c.ID)
		if m.RulesVersion >= 2 && n >= 1 && (!p.Double || c.Level("capital") < 1) || m.RulesVersion < 2 && (n >= 2 || n == 1 && (!p.Double || c.DoubleRound != 0 && c.DoubleRound != m.Round)) {
			return invalid("Лимит дипломатических визитов исчерпан")
		}
		m.Meetings = append(m.Meetings, Meeting{ID: fmt.Sprintf("meeting-%d", len(m.Meetings)+1), Round: m.Round, From: c.ID, To: other.ID, Status: "pending", Double: p.Double, Messages: []Message{}})
		return nil
	}
	for i := range m.Meetings {
		mt := &m.Meetings[i]
		if mt.ID != p.ID {
			continue
		}
		if mt.Round != m.Round {
			return invalid("Переговоры этого раунда завершены")
		}
		if kind == "meeting.respond" {
			if mt.To != c.ID || mt.Status != "pending" {
				return denied()
			}
			if p.Accept {
				sender := m.Country(mt.From)
				n := accepted(mt.From)
				if !sender.Alive() || m.RulesVersion >= 2 && n >= 1 && (!mt.Double || sender.Level("capital") < 1) || m.RulesVersion < 2 && (n >= 2 || n == 1 && (!mt.Double || sender.DoubleRound != 0 && sender.DoubleRound != m.Round)) {
					return invalid("Лимит дипломатических визитов исчерпан")
				}
				if m.RulesVersion >= 2 && n >= 1 {
					sender.CityByRole("capital").Level--
				} else if n == 1 {
					sender.DoubleRound = m.Round
				}
				mt.Status = "accepted"
			} else {
				mt.Status = "rejected"
			}
			return nil
		}
		if (mt.From != c.ID && mt.To != c.ID) || mt.Status != "accepted" {
			return denied()
		}
		p.Text = strings.TrimSpace(p.Text)
		if len([]rune(p.Text)) < 1 || len([]rune(p.Text)) > 1000 || len(mt.Messages) >= 100 {
			return invalid("Сообщение: 1–1000 символов, до 100 за встречу")
		}
		mt.Messages = append(mt.Messages, Message{CountryID: c.ID, Name: a.Name, Text: p.Text, At: now.UnixMilli()})
		return nil
	}
	return invalid("Встреча не найдена")
}
