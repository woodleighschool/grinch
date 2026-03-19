package authhttp

import (
	"net/http"

	"github.com/go-pkgz/auth/v2/token"

	httpapi "github.com/woodleighschool/grinch/internal/transport/http/httpapi"
)

func NewAPIMiddleware(
	sessionAuth func(http.Handler) http.Handler,
) func(http.Handler) http.Handler {
	if sessionAuth == nil {
		panic("sessionAuth middleware is required")
	}

	return func(next http.Handler) http.Handler {
		return sessionAuth(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			user, err := token.GetUserInfo(request)
			if err != nil || user.ID == "" {
				httpapi.WriteProblem(writer, http.StatusUnauthorized, httpapi.ProblemSpec{
					Type:   "urn:grinch:problem:unauthorized",
					Title:  "Unauthorized",
					Code:   "unauthorized",
					Detail: "Authentication is required.",
				})
				return
			}
			next.ServeHTTP(writer, request)
		}))
	}
}
