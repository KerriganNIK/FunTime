package httpapi

import (
	"errors"
	roomv1 "funtime.local/contracts/room/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRoomCookieAndOrigin(t *testing.T) {
	t.Setenv("COOKIE_SECURE", "true")
	w := httptest.NewRecorder()
	setSession(w, &roomv1.SessionReply{Code: "ABC234", SessionToken: "secret"})
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].HttpOnly || !cookies[0].Secure || cookies[0].SameSite != http.SameSiteStrictMode || cookies[0].Path != "/api/v1/rooms/ABC234" {
		t.Fatal("unsafe session cookie")
	}
	r := httptest.NewRequest("POST", "http://localhost/api/v1/rooms", nil)
	r.Header.Set("Origin", "https://evil.example")
	if sameOrigin(r) {
		t.Fatal("cross origin accepted")
	}
	r.Header.Set("Origin", "http://localhost")
	if !sameOrigin(r) {
		t.Fatal("same origin denied")
	}
	r.Header.Set("Sec-Fetch-Site", "cross-site")
	if sameOrigin(r) {
		t.Fatal("cross site accepted")
	}
}
func TestRoomErrorsAndOversizedBodies(t *testing.T) {
	for _, tc := range []struct {
		err  error
		code int
	}{{status.Error(codes.Aborted, "План изменился"), 409}, {status.Error(codes.Unauthenticated, "Войдите"), 401}, {errors.New("postgres password secret"), 503}} {
		w := httptest.NewRecorder()
		roomError(w, tc.err)
		if w.Code != tc.code || strings.Contains(w.Body.String(), "secret") {
			t.Fatal("wrong public error")
		}
	}
	for _, body := range []string{`{} {}`, `{"unknown":1}`, `{"name":"` + strings.Repeat("x", 65536) + `"}`} {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/", strings.NewReader(body))
		var target struct {
			Name string `json:"name"`
		}
		if readBody(w, r, &target) {
			t.Fatal("invalid body accepted")
		}
	}
}
