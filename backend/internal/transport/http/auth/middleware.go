package authhttp

import (
	"net/http"

	"github.com/go-pkgz/auth/v2/token"
)

func APIMiddleware(
	sessionAuth func(http.Handler) http.Handler,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return sessionAuth(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			user, err := token.GetUserInfo(request)
			if err != nil || user.ID == "" {
				writeUnauthorized(writer)
				return
			}
			next.ServeHTTP(writer, request)
		}))
	}
}

func writeUnauthorized(w http.ResponseWriter) {
	const body = `{"type":"urn:grinch:problem:unauthorized","title":"Unauthorized","status":401,"code":"unauthorized","detail":"Authentication is required."}`
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(body))
}
