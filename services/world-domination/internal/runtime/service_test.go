package runtime

import (
	"context"
	"encoding/json"
	"errors"
	runtimev1 "funtime.local/contracts/runtime/v1"
	"funtime.local/platform/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
	"time"
)

type failingStore struct {
	store.Store
	fail bool
}

func (s *failingStore) Save(ctx context.Context, k string, v any) error {
	if s.fail {
		return errors.New("disk full")
	}
	return s.Store.Save(ctx, k, v)
}
func TestPersistenceIdempotencyAndFailedWrite(t *testing.T) {
	ctx := context.Background()
	disk, err := store.Open(ctx, "", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer disk.Close()
	fs := &failingStore{Store: disk}
	s, err := New(ctx, fs)
	if err != nil {
		t.Fatal(err)
	}
	host := &runtimev1.Actor{ParticipantId: "host", Host: true}
	create := &runtimev1.CreateRequest{MatchId: "ABC234", CountryIds: []string{"norway", "germany"}, PhaseSeconds: 720, Actor: host}
	if _, err = s.Create(ctx, create); err != nil {
		t.Fatal(err)
	}
	cmd := &runtimev1.CommandRequest{MatchId: "ABC234", Actor: host, Kind: "advance", PayloadJson: []byte(`{"expectedPhase":"1:council"}`), RequestId: "request-1234"}
	fs.fail = true
	if _, err = s.Command(ctx, cmd); status.Code(err) != codes.Unavailable {
		t.Fatal("save failure ignored")
	}
	if s.matches["ABC234"].Phase != "council" {
		t.Fatal("uncommitted mutation leaked")
	}
	fs.fail = false
	if _, err = s.Command(ctx, cmd); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Command(ctx, cmd); err != nil {
		t.Fatal("duplicate must return committed state", err)
	}
	if s.matches["ABC234"].Phase != "headquarters" {
		t.Fatal("duplicate advanced twice")
	}
	restored, err := New(ctx, disk)
	if err != nil {
		t.Fatal(err)
	}
	if restored.matches["ABC234"].Phase != "headquarters" {
		t.Fatal("restart lost state")
	}
	if _, err = restored.Command(ctx, cmd); err != nil {
		t.Fatal("idempotency lost on restart", err)
	}
	if _, err = restored.Create(ctx, create); err != nil {
		t.Fatal(err)
	}
	if restored.matches["ABC234"].Phase != "headquarters" {
		t.Fatal("retry create reset game")
	}
}
func TestTimerExpiresOnceAndDoesNotSkipRoundsAfterDowntime(t *testing.T) {
	ctx := context.Background()
	disk, _ := store.Open(ctx, "", t.TempDir())
	defer disk.Close()
	s, _ := New(ctx, disk)
	_, err := s.Create(ctx, &runtimev1.CreateRequest{MatchId: "ABC234", CountryIds: []string{"norway", "germany"}, PhaseSeconds: 720, Actor: &runtimev1.Actor{Host: true}})
	if err != nil {
		t.Fatal(err)
	}
	m := s.matches["ABC234"]
	m.Deadline = time.Now().Add(-24 * time.Hour).UnixMilli()
	s.Tick(ctx)
	s.Tick(ctx)
	if s.matches[m.ID].Phase != "headquarters" || s.matches[m.ID].Round != 1 {
		t.Fatal("timer skipped phases")
	}
	reply, err := s.Snapshot(ctx, &runtimev1.SnapshotRequest{MatchId: m.ID})
	if err != nil {
		t.Fatal(err)
	}
	var v map[string]any
	if json.Unmarshal(reply.Json, &v) != nil {
		t.Fatal("invalid snapshot")
	}
}
