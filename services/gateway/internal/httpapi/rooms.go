package httpapi

import (
	"context"
	"encoding/json"
	roomv1 "funtime.local/contracts/room/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

var roomCode = regexp.MustCompile(`^[A-Z2-9]{6}$`)

func roomError(w http.ResponseWriter, err error) {
	s := status.Convert(err)
	code := http.StatusServiceUnavailable
	message := "Игровой сервис временно недоступен"
	switch s.Code() {
	case codes.InvalidArgument:
		code = 400
	case codes.Unauthenticated:
		code = 401
	case codes.PermissionDenied:
		code = 403
	case codes.NotFound:
		code = 404
	case codes.Aborted, codes.AlreadyExists, codes.FailedPrecondition:
		code = 409
	case codes.ResourceExhausted:
		code = 429
	}
	if code < 500 {
		message = s.Message()
	}
	writeJSON(w, code, map[string]any{"error": map[string]string{"code": s.Code().String(), "message": message}})
}
func readBody(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 65536)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		writeJSON(w, 400, map[string]any{"error": map[string]string{"message": "Некорректный запрос"}})
		return false
	}
	if d.Decode(&struct{}{}) != io.EOF {
		writeJSON(w, 400, map[string]any{"error": map[string]string{"message": "Лишние данные в запросе"}})
		return false
	}
	return true
}
func cookieToken(r *http.Request, code string) string {
	c, err := r.Cookie("ft_" + code)
	if err != nil {
		return ""
	}
	return c.Value
}
func setSession(w http.ResponseWriter, s *roomv1.SessionReply) {
	http.SetCookie(w, &http.Cookie{Name: "ft_" + s.Code, Value: s.SessionToken, Path: "/api/v1/rooms/" + s.Code, HttpOnly: true, Secure: os.Getenv("COOKIE_SECURE") == "true", SameSite: http.SameSiteStrictMode, MaxAge: 7 * 24 * 3600})
}
func sameOrigin(r *http.Request) bool {
	if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
		return false
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	expected := os.Getenv("PUBLIC_ORIGIN")
	if expected != "" {
		return origin == expected
	}
	return u.Host == r.Host
}
func registerRooms(mux *http.ServeMux, client roomv1.RoomServiceClient) {
	handle := func(pattern string, fn func(http.ResponseWriter, *http.Request, context.Context, string)) {
		mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" && !sameOrigin(r) {
				writeJSON(w, 403, map[string]any{"error": map[string]string{"message": "Запрос с другого сайта запрещён"}})
				return
			}
			code := strings.ToUpper(r.PathValue("code"))
			if code != "" && !roomCode.MatchString(code) {
				writeJSON(w, 404, map[string]any{"error": map[string]string{"message": "Комната не найдена"}})
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			fn(w, r, ctx, code)
		})
	}
	handle("POST /api/v1/rooms", func(w http.ResponseWriter, r *http.Request, ctx context.Context, code string) {
		var b struct {
			Name   string `json:"name"`
			GameID string `json:"gameId"`
		}
		if !readBody(w, r, &b) {
			return
		}
		s, err := client.Create(ctx, &roomv1.CreateRequest{Name: b.Name, GameId: b.GameID})
		if err != nil {
			roomError(w, err)
			return
		}
		setSession(w, s)
		writeJSON(w, 201, map[string]string{"code": s.Code})
	})
	handle("POST /api/v1/rooms/{code}/join", func(w http.ResponseWriter, r *http.Request, ctx context.Context, code string) {
		var b struct {
			Name string `json:"name"`
		}
		if !readBody(w, r, &b) {
			return
		}
		s, err := client.Join(ctx, &roomv1.JoinRequest{Code: code, Name: b.Name, SessionToken: cookieToken(r, code)})
		if err != nil {
			roomError(w, err)
			return
		}
		setSession(w, s)
		writeJSON(w, 200, map[string]string{"code": s.Code})
	})
	snapshot := func(public bool) func(http.ResponseWriter, *http.Request, context.Context, string) {
		return func(w http.ResponseWriter, r *http.Request, ctx context.Context, code string) {
			s, err := client.Snapshot(ctx, &roomv1.SnapshotRequest{Code: code, SessionToken: cookieToken(r, code), PublicScreen: public})
			if err != nil {
				roomError(w, err)
				return
			}
			writeJSON(w, 200, json.RawMessage(s.Json))
		}
	}
	handle("GET /api/v1/rooms/{code}", snapshot(false))
	handle("GET /api/v1/rooms/{code}/screen", snapshot(true))
	handle("POST /api/v1/rooms/{code}/commands", func(w http.ResponseWriter, r *http.Request, ctx context.Context, code string) {
		var b struct {
			Kind      string          `json:"kind"`
			Payload   json.RawMessage `json:"payload"`
			RequestID string          `json:"requestId"`
		}
		if !readBody(w, r, &b) {
			return
		}
		s, err := client.Command(ctx, &roomv1.CommandRequest{Code: code, SessionToken: cookieToken(r, code), Kind: b.Kind, PayloadJson: b.Payload, RequestId: b.RequestID})
		if err != nil {
			roomError(w, err)
			return
		}
		writeJSON(w, 200, json.RawMessage(s.Json))
	})
}
