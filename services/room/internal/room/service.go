package room

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	roomv1 "funtime.local/contracts/room/v1"
	runtimev1 "funtime.local/contracts/runtime/v1"
	"funtime.local/platform/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Participant struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CountryID string `json:"countryId"`
	Role      string `json:"role"`
	Host      bool   `json:"host"`
	TokenHash string `json:"tokenHash"`
}
type Room struct {
	Code         string        `json:"code"`
	GameID       string        `json:"gameId"`
	Status       string        `json:"status"`
	Revision     int           `json:"revision"`
	PhaseSeconds int           `json:"phaseSeconds"`
	Participants []Participant `json:"participants"`
	CountryIDs   []string      `json:"countryIds"`
}
type Definition struct {
	ID        string `json:"id"`
	Countries []struct {
		ID     string   `json:"id"`
		Name   string   `json:"name"`
		Cities []string `json:"cities"`
	} `json:"countries"`
	Roles []string `json:"roles"`
}
type Service struct {
	roomv1.UnimplementedRoomServiceServer
	mu       sync.Mutex
	store    store.Store
	runtimes map[string]runtimev1.GameRuntimeClient
}

func New(s store.Store, runtimes map[string]runtimev1.GameRuntimeClient) *Service {
	return &Service{store: s, runtimes: runtimes}
}
func token() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
func hash(s string) string { v := sha256.Sum256([]byte(s)); return hex.EncodeToString(v[:]) }
func name(s string) (string, error) {
	s = strings.TrimSpace(s)
	if len([]rune(s)) < 1 || len([]rune(s)) > 32 {
		return "", status.Error(codes.InvalidArgument, "Имя: от 1 до 32 символов")
	}
	return s, nil
}
func internal(err error) error {
	slog.Error("room persistence failure", "error", err)
	return status.Error(codes.Unavailable, "Не удалось сохранить комнату")
}
func (s *Service) load(ctx context.Context, code string) (*Room, error) {
	var r Room
	err := s.store.Load(ctx, strings.ToUpper(code), &r)
	if errors.Is(err, store.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "Комната не найдена")
	}
	if err != nil {
		return nil, internal(err)
	}
	return &r, nil
}
func session(r *Room, t string) (*Participant, error) {
	if t != "" {
		h := hash(t)
		for i := range r.Participants {
			if r.Participants[i].TokenHash == h {
				return &r.Participants[i], nil
			}
		}
	}
	return nil, status.Error(codes.Unauthenticated, "Присоединитесь к комнате")
}
func (s *Service) Create(ctx context.Context, q *roomv1.CreateRequest) (*roomv1.SessionReply, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n, err := name(q.Name)
	if err != nil {
		return nil, err
	}
	if s.runtimes[q.GameId] == nil {
		return nil, status.Error(codes.InvalidArgument, "Игра недоступна")
	}
	alphabet := "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	var code string
	for {
		b := make([]byte, 6)
		if _, err = rand.Read(b); err != nil {
			return nil, internal(err)
		}
		for i := range b {
			b[i] = alphabet[int(b[i])%len(alphabet)]
		}
		code = string(b)
		var existing Room
		err = s.store.Load(ctx, code, &existing)
		if errors.Is(err, store.ErrNotFound) {
			break
		}
		if err != nil {
			return nil, internal(err)
		}
	}
	t := token()
	r := Room{Code: code, GameID: q.GameId, Status: "lobby", Revision: 1, PhaseSeconds: 720, Participants: []Participant{{ID: token()[:16], Name: n, Host: true, TokenHash: hash(t)}}}
	if err = s.store.Save(ctx, code, r); err != nil {
		return nil, internal(err)
	}
	return &roomv1.SessionReply{Code: code, SessionToken: t}, nil
}
func (s *Service) Join(ctx context.Context, q *roomv1.JoinRequest) (*roomv1.SessionReply, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, err := s.load(ctx, q.Code)
	if err != nil {
		return nil, err
	}
	if _, err = session(r, q.SessionToken); err == nil {
		return &roomv1.SessionReply{Code: r.Code, SessionToken: q.SessionToken}, nil
	}
	if r.Status != "lobby" {
		return nil, status.Error(codes.FailedPrecondition, "Игра уже началась. Вернитесь с устройства, на котором присоединились")
	}
	if len(r.Participants) >= 51 {
		return nil, status.Error(codes.ResourceExhausted, "Комната заполнена")
	}
	n, err := name(q.Name)
	if err != nil {
		return nil, err
	}
	t := token()
	r.Participants = append(r.Participants, Participant{ID: token()[:16], Name: n, TokenHash: hash(t)})
	r.Revision++
	if err = s.store.Save(ctx, r.Code, r); err != nil {
		return nil, internal(err)
	}
	return &roomv1.SessionReply{Code: r.Code, SessionToken: t}, nil
}
func asActor(p *Participant) *runtimev1.Actor {
	if p == nil {
		return &runtimev1.Actor{}
	}
	return &runtimev1.Actor{ParticipantId: p.ID, Name: p.Name, CountryId: p.CountryID, Host: p.Host}
}
func (s *Service) snapshot(ctx context.Context, r *Room, p *Participant) (*roomv1.SnapshotReply, error) {
	client := s.runtimes[r.GameID]
	def, err := client.Definition(ctx, &runtimev1.DefinitionRequest{})
	if err != nil {
		return nil, err
	}
	participants := []any{}
	for _, v := range r.Participants {
		participants = append(participants, map[string]any{"id": v.ID, "name": v.Name, "countryId": v.CountryID, "role": v.Role, "host": v.Host})
	}
	v := map[string]any{"code": r.Code, "gameId": r.GameID, "status": r.Status, "revision": r.Revision, "phaseSeconds": r.PhaseSeconds, "participants": participants, "definition": json.RawMessage(def.Json), "serverTime": time.Now().UnixMilli()}
	if p != nil {
		v["me"] = p.ID
	}
	if r.Status == "playing" {
		match, err := client.Snapshot(ctx, &runtimev1.SnapshotRequest{MatchId: r.Code, Actor: asActor(p)})
		if err != nil {
			return nil, err
		}
		v["match"] = json.RawMessage(match.Json)
	}
	b, err := json.Marshal(v)
	return &roomv1.SnapshotReply{Json: b}, err
}
func (s *Service) Snapshot(ctx context.Context, q *roomv1.SnapshotRequest) (*roomv1.SnapshotReply, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, err := s.load(ctx, q.Code)
	if err != nil {
		return nil, err
	}
	var p *Participant
	if !q.PublicScreen {
		p, err = session(r, q.SessionToken)
		if err != nil {
			return nil, err
		}
	}
	return s.snapshot(ctx, r, p)
}
func (s *Service) Command(ctx context.Context, q *roomv1.CommandRequest) (*roomv1.SnapshotReply, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, err := s.load(ctx, q.Code)
	if err != nil {
		return nil, err
	}
	p, err := session(r, q.SessionToken)
	if err != nil {
		return nil, err
	}
	client := s.runtimes[r.GameID]
	if q.Kind == "select" {
		if r.Status != "lobby" {
			return nil, status.Error(codes.FailedPrecondition, "Выбор команды завершён")
		}
		var body struct {
			CountryID string `json:"countryId"`
			Role      string `json:"role"`
		}
		if json.Unmarshal(q.PayloadJson, &body) != nil {
			return nil, status.Error(codes.InvalidArgument, "Некорректные данные")
		}
		def, err := client.Definition(ctx, &runtimev1.DefinitionRequest{})
		if err != nil {
			return nil, err
		}
		var d Definition
		if err = json.Unmarshal(def.Json, &d); err != nil {
			return nil, internal(err)
		}
		country, role := false, false
		for _, c := range d.Countries {
			country = country || c.ID == body.CountryID
		}
		for _, v := range d.Roles {
			role = role || v == body.Role
		}
		if !country || !role {
			return nil, status.Error(codes.InvalidArgument, "Выберите страну и роль")
		}
		for _, other := range r.Participants {
			if other.ID != p.ID && other.CountryID == body.CountryID && other.Role == body.Role {
				return nil, status.Error(codes.AlreadyExists, "Эта роль уже занята")
			}
		}
		p.CountryID = body.CountryID
		p.Role = body.Role
	} else if q.Kind == "start" {
		if !p.Host {
			return nil, status.Error(codes.PermissionDenied, "Только ведущий запускает игру")
		}
		if r.Status == "playing" {
			return s.snapshot(ctx, r, p)
		}
		if r.Status == "lobby" {
			var body struct {
				PhaseSeconds int `json:"phaseSeconds"`
			}
			if json.Unmarshal(q.PayloadJson, &body) != nil {
				return nil, status.Error(codes.InvalidArgument, "Некорректные данные")
			}
			if body.PhaseSeconds < 10 || body.PhaseSeconds > 3600 {
				return nil, status.Error(codes.InvalidArgument, "Длительность фазы: 10–3600 секунд")
			}
			r.PhaseSeconds = body.PhaseSeconds
			r.CountryIDs = []string{}
			seen := map[string]bool{}
			for _, v := range r.Participants {
				if v.CountryID == "" && !v.Host {
					return nil, status.Error(codes.FailedPrecondition, "Все игроки должны выбрать страну и роль")
				}
				if v.CountryID != "" && !seen[v.CountryID] {
					seen[v.CountryID] = true
					r.CountryIDs = append(r.CountryIDs, v.CountryID)
				}
			}
			if len(r.CountryIDs) < 2 {
				return nil, status.Error(codes.FailedPrecondition, "Нужны хотя бы две страны с игроками")
			}
			r.Status = "starting"
			r.Revision++
			if err = s.store.Save(ctx, r.Code, r); err != nil {
				return nil, internal(err)
			}
		}
		if _, err = client.Create(ctx, &runtimev1.CreateRequest{MatchId: r.Code, CountryIds: r.CountryIDs, PhaseSeconds: int32(r.PhaseSeconds), Actor: asActor(p)}); err != nil {
			return nil, err
		}
		r.Status = "playing"
	} else {
		if r.Status != "playing" {
			return nil, status.Error(codes.FailedPrecondition, "Игра ещё не началась")
		}
		if _, err = client.Command(ctx, &runtimev1.CommandRequest{MatchId: r.Code, Actor: asActor(p), Kind: q.Kind, PayloadJson: q.PayloadJson, RequestId: q.RequestId}); err != nil {
			return nil, err
		}
		return s.snapshot(ctx, r, p)
	}
	r.Revision++
	if err = s.store.Save(ctx, r.Code, r); err != nil {
		return nil, internal(err)
	}
	return s.snapshot(ctx, r, p)
}
