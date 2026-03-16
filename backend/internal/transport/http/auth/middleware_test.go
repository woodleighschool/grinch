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

func TestNewAPIMiddleware_AuthenticatesSessionRequests(t *testing.T) {
	t.Parallel()

	middleware := authhttp.NewAPIMiddleware(sessionAuthStub)

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

func TestNewAPIMiddleware_ReturnsUnauthorizedWithoutUserInfo(t *testing.T) {
	t.Parallel()

	middleware := authhttp.NewAPIMiddleware(func(next http.Handler) http.Handler {
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
	assertAuthErrorBody(t, response, "unauthorized")
}

func assertAuthErrorBody(t *testing.T, response *httptest.ResponseRecorder, want string) {
	t.Helper()

	if got := response.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	var body map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if body["error"] != want {
		t.Fatalf("error body = %q, want %q", body["error"], want)
	}
}
