package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	catalogv1 "funtime.local/contracts/catalog/v1"
	roomv1 "funtime.local/contracts/room/v1"
	"google.golang.org/grpc"
)

type CatalogClient interface {
	ListGames(context.Context, *catalogv1.ListGamesRequest, ...grpc.CallOption) (*catalogv1.ListGamesResponse, error)
}

// HTTP DTOs are deliberately separate from the internal Protobuf wire format.
type gameDTO struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Status      string `json:"status"`
	Accent      string `json:"accent"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func New(client CatalogClient, rooms ...roomv1.RoomServiceClient) http.Handler {
	mux := http.NewServeMux()
	if len(rooms) > 0 {
		registerRooms(mux, rooms[0])
	}
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()
		if _, err := client.ListGames(ctx, &catalogv1.ListGamesRequest{}); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /api/v1/games", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		response, err := client.ListGames(ctx, &catalogv1.ListGamesRequest{})
		if err != nil {
			slog.WarnContext(ctx, "catalog request failed", "error", err)
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": map[string]string{
				"code": "CATALOG_UNAVAILABLE", "message": "Каталог временно недоступен. Попробуйте ещё раз.",
			}})
			return
		}
		games := make([]gameDTO, 0, len(response.GetGames()))
		for _, game := range response.GetGames() {
			games = append(games, gameDTO{ID: game.GetId(), Title: game.GetTitle(), Description: game.GetDescription(), Category: game.GetCategory(), Status: game.GetStatus(), Accent: game.GetAccent()})
		}
		writeJSON(w, http.StatusOK, struct {
			Games []gameDTO `json:"games"`
		}{games})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": map[string]string{"code": "NOT_FOUND", "message": "Маршрут не найден."}})
	})
	return mux
}
