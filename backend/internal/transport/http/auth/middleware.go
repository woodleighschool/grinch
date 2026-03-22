package authhttp

import (
	"net/http"

	"github.com/go-pkgz/auth/v2/token"

	apihttp "github.com/woodleighschool/grinch/internal/transport/http/api"
)

func APIMiddleware(
	sessionAuth func(http.Handler) http.Handler,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return sessionAuth(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			user, err := token.GetUserInfo(request)
			if err != nil || user.ID == "" {
				apihttp.WriteProblem(writer, http.StatusUnauthorized, apihttp.ProblemSpec{
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
