package authhttp_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-pkgz/auth/v2/token"

	authhttp "github.com/woodleighschool/grinch/internal/transport/http/auth"
)

func sessionAuthStub(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		user := token.User{ID: "session_user", Name: "Session User"}
		next.ServeHTTP(writer, token.SetUserInfo(request, user))
	})
}

func TestAPIMiddleware_AuthenticatesSessionRequests(t *testing.T) {
	t.Parallel()

	middleware := authhttp.APIMiddleware(sessionAuthStub)

	called := false
	handler := middleware(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		called = true
		writer.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/machines", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if !called {
		t.Fatalf("expected downstream handler to be called")
	}
}

func TestAPIMiddleware_ReturnsUnauthorizedWithoutUserInfo(t *testing.T) {
	t.Parallel()

	middleware := authhttp.APIMiddleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			next.ServeHTTP(writer, request)
		})
	})

	handler := middleware(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}
	assertAuthErrorBody(t, response)
}

func assertAuthErrorBody(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()

	if got := response.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	var body struct {
		Type   string `json:"type"`
		Title  string `json:"title"`
		Status int    `json:"status"`
		Detail string `json:"detail"`
		Code   string `json:"code"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if body.Type != "urn:grinch:problem:unauthorized" {
		t.Fatalf("Type = %q, want unauthorized problem type", body.Type)
	}
	if body.Title != "Unauthorized" {
		t.Fatalf("Title = %q, want Unauthorized", body.Title)
	}
	if body.Status != http.StatusUnauthorized {
		t.Fatalf("Status = %d, want 401", body.Status)
	}
	if body.Detail != "Authentication is required." {
		t.Fatalf("Detail = %q, want authentication detail", body.Detail)
	}
	if body.Code != "unauthorized" {
		t.Fatalf("Code = %q, want unauthorized", body.Code)
	}
}
