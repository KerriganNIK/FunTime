package room

import (
	"context"
	"encoding/json"
	roomv1 "funtime.local/contracts/room/v1"
	runtimev1 "funtime.local/contracts/runtime/v1"
	"funtime.local/platform/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	"sync"
	"testing"
)

type runtimeStub struct {
	creates int
	public  bool
}

func (s *runtimeStub) Definition(context.Context, *runtimev1.DefinitionRequest, ...grpc.CallOption) (*runtimev1.JsonReply, error) {
	return &runtimev1.JsonReply{Json: []byte(`{"id":"world-domination","countries":[{"id":"norway","name":"Норвегия"},{"id":"germany","name":"Германия"}],"roles":["president","economy"]}`)}, nil
}
func (s *runtimeStub) Create(context.Context, *runtimev1.CreateRequest, ...grpc.CallOption) (*runtimev1.JsonReply, error) {
	s.creates++
	return &runtimev1.JsonReply{Json: []byte(`{}`)}, nil
}
func (s *runtimeStub) Snapshot(_ context.Context, q *runtimev1.SnapshotRequest, _ ...grpc.CallOption) (*runtimev1.JsonReply, error) {
	s.public = !q.Actor.Host && q.Actor.CountryId == ""
	return &runtimev1.JsonReply{Json: []byte(`{}`)}, nil
}
func (s *runtimeStub) Command(context.Context, *runtimev1.CommandRequest, ...grpc.CallOption) (*runtimev1.JsonReply, error) {
	return &runtimev1.JsonReply{Json: []byte(`{}`)}, nil
}
func TestRoomSessionsRolesAndStart(t *testing.T) {
	ctx := context.Background()
	disk, err := store.Open(ctx, "", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer disk.Close()
	runtime := &runtimeStub{}
	s := New(disk, map[string]runtimev1.GameRuntimeClient{"world-domination": runtime})
	host, err := s.Create(ctx, &roomv1.CreateRequest{Name: "Ведущий", GameId: "world-domination"})
	if err != nil {
		t.Fatal(err)
	}
	a, err := s.Join(ctx, &roomv1.JoinRequest{Code: host.Code, Name: "A"})
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.Join(ctx, &roomv1.JoinRequest{Code: host.Code, Name: "B"})
	if err != nil {
		t.Fatal(err)
	}
	command := func(session *roomv1.SessionReply, kind string, payload string) error {
		_, err := s.Command(ctx, &roomv1.CommandRequest{Code: session.Code, SessionToken: session.SessionToken, Kind: kind, PayloadJson: []byte(payload), RequestId: "request-1234"})
		return err
	}
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, player := range []*roomv1.SessionReply{a, b} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- command(player, "select", `{"countryId":"norway","role":"president"}`)
		}()
	}
	wg.Wait()
	close(results)
	success, conflict := 0, 0
	for err := range results {
		if err == nil {
			success++
		} else if status.Code(err) == codes.AlreadyExists {
			conflict++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatal("role race assigned two presidents")
	}
	if err = command(a, "select", `{"countryId":"norway","role":"economy"}`); err != nil {
		t.Fatal(err)
	}
	if err = command(b, "select", `{"countryId":"germany","role":"president"}`); err != nil {
		t.Fatal(err)
	}
	if status.Code(command(a, "start", `{"phaseSeconds":720}`)) != codes.PermissionDenied {
		t.Fatal("non-host started game")
	}
	if err = command(host, "start", `{"phaseSeconds":720}`); err != nil {
		t.Fatal(err)
	}
	if err = command(host, "start", `{"phaseSeconds":720}`); err != nil || runtime.creates != 1 {
		t.Fatal("duplicate start")
	}
	view, err := s.Snapshot(ctx, &roomv1.SnapshotRequest{Code: host.Code, SessionToken: host.SessionToken, PublicScreen: true})
	if err != nil {
		t.Fatal(err)
	}
	if !runtime.public {
		t.Fatal("host cookie leaked into public projection")
	}
	if strings.Contains(string(view.Json), "tokenHash") || strings.Contains(string(view.Json), host.SessionToken) || strings.Contains(string(view.Json), `"me"`) {
		t.Fatal("session leaked")
	}
	if _, err = s.Snapshot(ctx, &roomv1.SnapshotRequest{Code: host.Code, SessionToken: "forged"}); status.Code(err) != codes.Unauthenticated {
		t.Fatal("forged session accepted")
	}
	restored := New(disk, s.runtimes)
	if _, err = restored.Snapshot(ctx, &roomv1.SnapshotRequest{Code: host.Code, SessionToken: a.SessionToken}); err != nil {
		t.Fatal("session lost after restart")
	}
	joined, err := restored.Join(ctx, &roomv1.JoinRequest{Code: host.Code, SessionToken: a.SessionToken})
	if err != nil || joined.SessionToken != a.SessionToken {
		t.Fatal("resume failed")
	}
	r, err := restored.load(ctx, host.Code)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(r)
	if strings.Contains(string(raw), host.SessionToken) {
		t.Fatal("raw bearer token persisted")
	}
}
