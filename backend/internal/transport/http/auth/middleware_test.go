package authhttp_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-pkgz/auth/v2/token"

	authhttp "github.com/woodleighschool/grinch/internal/transport/http/auth"
)

func testSessionAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		user := token.User{ID: "session_user", Name: "Session User"}
		next.ServeHTTP(writer, token.SetUserInfo(request, user))
	})
}

func TestAPIMiddleware_AuthenticatesSessionRequests(t *testing.T) {
	middleware := authhttp.APIMiddleware(testSessionAuth)

	called := false
	handler := middleware(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		called = true
		writer.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/machines", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("Code = %d, want 200", response.Code)
	}
	if !called {
		t.Fatalf("called = false, want true")
	}
}

func TestAPIMiddleware_ReturnsUnauthorizedWithoutUserInfo(t *testing.T) {
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
		t.Fatalf("Code = %d, want 401", response.Code)
	}
	if response.Body.Len() != 0 {
		t.Fatalf("Body length = %d, want 0", response.Body.Len())
	}
}
