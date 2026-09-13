package httpapi

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	catalogv1 "funtime.local/contracts/catalog/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type testCatalog struct {
	catalogv1.UnimplementedCatalogServiceServer
	fail bool
}

func (s *testCatalog) ListGames(context.Context, *catalogv1.ListGamesRequest) (*catalogv1.ListGamesResponse, error) {
	if s.fail {
		return nil, status.Error(codes.Unavailable, "internal host secret.example:9001")
	}
	return &catalogv1.ListGamesResponse{Games: []*catalogv1.Game{{Id: "world-domination", Title: "Мировое господство", Description: "Скоро", Category: "Игра для компании", Status: "coming_soon", Accent: "lime"}}}, nil
}

func grpcClient(t *testing.T, fail bool) catalogv1.CatalogServiceClient {
	t.Helper()
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	catalogv1.RegisterCatalogServiceServer(server, &testCatalog{fail: fail})
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)
	connection, err := grpc.NewClient("passthrough:///catalog", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	return catalogv1.NewCatalogServiceClient(connection)
}

func TestCatalogGRPCToHTTP(t *testing.T) {
	router := New(grpcClient(t, false))
	result := httptest.NewRecorder()
	router.ServeHTTP(result, httptest.NewRequest(http.MethodGet, "/api/v1/games", nil))
	if result.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", result.Code, result.Body)
	}
	var body struct {
		Games []gameDTO `json:"games"`
	}
	if err := json.Unmarshal(result.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Games) != 1 || body.Games[0].ID != "world-domination" || body.Games[0].Status != "coming_soon" {
		t.Fatalf("unexpected HTTP catalog: %+v", body)
	}
	if result.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Fatal("missing JSON content type")
	}
}

func TestUnavailableCatalogDoesNotLeakInternalError(t *testing.T) {
	router := New(grpcClient(t, true))
	result := httptest.NewRecorder()
	router.ServeHTTP(result, httptest.NewRequest(http.MethodGet, "/api/v1/games", nil))
	if result.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", result.Code)
	}
	var body struct {
		Error struct {
			Code    string
			Message string
		}
	}
	if err := json.Unmarshal(result.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Code != "CATALOG_UNAVAILABLE" || body.Error.Message != "Каталог временно недоступен. Попробуйте ещё раз." {
		t.Fatalf("unexpected public error: %+v", body)
	}
	for _, test := range []struct {
		path string
		want int
	}{{"/healthz", 200}, {"/readyz", 503}, {"/api/unknown", 404}} {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
		if recorder.Code != test.want {
			t.Errorf("%s: got %d want %d", test.path, recorder.Code, test.want)
		}
	}
}
