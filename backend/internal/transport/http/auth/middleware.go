package authhttp

import (
	"net/http"

	"github.com/go-pkgz/auth/token"

	"github.com/woodleighschool/grinch/internal/transport/http/apierrors"
)

func NewAPIMiddleware(
	sessionAuth func(http.Handler) http.Handler,
) func(http.Handler) http.Handler {
	if sessionAuth == nil {
		panic("sessionAuth middleware is required")
	}

	return func(next http.Handler) http.Handler {
		return makeSessionProtected(next, sessionAuth)
	}
}

func makeSessionProtected(
	next http.Handler,
	sessionAuth func(http.Handler) http.Handler,
) http.Handler {
	return sessionAuth(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		user, err := token.GetUserInfo(request)
		if err != nil {
			writeError(writer, http.StatusUnauthorized, "unauthorized")
			return
		}
		if user.ID == "" {
			writeError(writer, http.StatusUnauthorized, "unauthorized")
			return
		}
		next.ServeHTTP(writer, request)
	}))
}

func writeError(writer http.ResponseWriter, statusCode int, message string) {
	apierrors.Write(writer, statusCode, message)
}
