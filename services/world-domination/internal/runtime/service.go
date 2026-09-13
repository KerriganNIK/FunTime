package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"slices"
	"sync"
	"time"

	runtimev1 "funtime.local/contracts/runtime/v1"
	"funtime.local/platform/store"
	"funtime.local/worlddomination/internal/game"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	runtimev1.UnimplementedGameRuntimeServer
	mu      sync.Mutex
	store   store.Store
	matches map[string]*game.Match
}

func New(ctx context.Context, s store.Store) (*Service, error) {
	v := &Service{store: s, matches: map[string]*game.Match{}}
	keys, err := s.Keys(ctx)
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		var m game.Match
		if err = s.Load(ctx, key, &m); err != nil {
			return nil, err
		}
		v.matches[key] = &m
	}
	return v, nil
}
func actor(a *runtimev1.Actor) game.Actor {
	return game.Actor{ID: a.GetParticipantId(), Name: a.GetName(), CountryID: a.GetCountryId(), Host: a.GetHost()}
}
func reply(v any) (*runtimev1.JsonReply, error) {
	b, err := json.Marshal(v)
	return &runtimev1.JsonReply{Json: b}, err
}
func (s *Service) Definition(context.Context, *runtimev1.DefinitionRequest) (*runtimev1.JsonReply, error) {
	return reply(map[string]any{"id": "world-domination", "countries": game.Definitions, "roles": []string{"president", "economy", "defense", "foreign", "statistics"}})
}
func (s *Service) Create(ctx context.Context, r *runtimev1.CreateRequest) (*runtimev1.JsonReply, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !r.GetActor().GetHost() {
		return nil, status.Error(codes.PermissionDenied, "Только ведущий запускает игру")
	}
	if m := s.matches[r.MatchId]; m != nil {
		return reply(m.View(actor(r.Actor), time.Now()))
	}
	m, err := game.New(r.MatchId, r.CountryIds, int(r.PhaseSeconds), time.Now())
	if err != nil {
		return nil, err
	}
	if err = s.store.Save(ctx, m.ID, m); err != nil {
		return nil, internal(err)
	}
	s.matches[m.ID] = m
	return reply(m.View(actor(r.Actor), time.Now()))
}
func (s *Service) Snapshot(ctx context.Context, r *runtimev1.SnapshotRequest) (*runtimev1.JsonReply, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.matches[r.MatchId]
	if m == nil {
		return nil, status.Error(codes.NotFound, "Игра не найдена")
	}
	return reply(m.View(actor(r.Actor), time.Now()))
}
func (s *Service) Command(ctx context.Context, r *runtimev1.CommandRequest) (*runtimev1.JsonReply, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.matches[r.MatchId]
	if m == nil {
		return nil, status.Error(codes.NotFound, "Игра не найдена")
	}
	if len(r.RequestId) < 8 || len(r.RequestId) > 100 {
		return nil, status.Error(codes.InvalidArgument, "Нужен идентификатор запроса")
	}
	key := r.GetActor().GetParticipantId() + ":" + r.RequestId
	if slices.Contains(m.Requests, key) {
		return reply(m.View(actor(r.Actor), time.Now()))
	}
	// Enforce the deadline even between timer ticks. A late order cannot sneak into the previous phase.
	if m.Phase != "finished" && !m.Paused && time.Now().UnixMilli() >= m.Deadline {
		expired, err := store.Clone(m)
		if err != nil {
			return nil, internal(err)
		}
		expired.Advance(time.Now())
		expired.Revision++
		if err = s.store.Save(ctx, expired.ID, expired); err != nil {
			return nil, internal(err)
		}
		s.matches[expired.ID] = expired
		return nil, status.Error(codes.Aborted, "Фаза уже завершилась. Обновите состояние игры")
	}
	copy, err := store.Clone(m)
	if err != nil {
		return nil, internal(err)
	}
	if err = copy.Command(actor(r.Actor), r.Kind, r.PayloadJson, time.Now()); err != nil {
		return nil, err
	}
	copy.Revision++
	copy.Requests = append(copy.Requests, key)
	if len(copy.Requests) > 4096 {
		copy.Requests = copy.Requests[len(copy.Requests)-4096:]
	}
	if err = s.store.Save(ctx, copy.ID, copy); err != nil {
		return nil, internal(err)
	}
	s.matches[copy.ID] = copy
	return reply(copy.View(actor(r.Actor), time.Now()))
}
func (s *Service) Tick(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for id, m := range s.matches {
		if m.Phase == "finished" || m.Paused || now.UnixMilli() < m.Deadline {
			continue
		}
		copy, err := store.Clone(m)
		if err == nil {
			copy.Advance(now)
			copy.Revision++
			err = s.store.Save(ctx, id, copy)
		}
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				slog.Error("phase save failed", "match", id, "error", err)
			}
			continue
		}
		s.matches[id] = copy
	}
}
func internal(err error) error {
	slog.Error("runtime persistence failure", "error", err)
	return status.Error(codes.Unavailable, "Не удалось сохранить игру. Повторите запрос")
}
