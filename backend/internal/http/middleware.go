package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/woodleighschool/grinch/internal/auth"
)

type contextKey string

const sessionContextKey contextKey = "session"

// AdminAuth enforces a valid session cookie and injects it into the request context.
func AdminAuth(sessions *auth.SessionManager, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if sessions == nil {
				writeError(w, http.StatusUnauthorized, "session manager missing")
				return
			}
			sess, err := sessions.Read(r)
			if err != nil {
				logger.Warn("unauthorized", "err", err)
				writeError(w, http.StatusUnauthorized, "auth required")
				return
			}
			ctx := context.WithValue(r.Context(), sessionContextKey, sess)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// SessionFromContext extracts the authenticated session from the request context.
func SessionFromContext(ctx context.Context) (auth.Session, bool) {
	val := ctx.Value(sessionContextKey)
	if val == nil {
		return auth.Session{}, false
	}
	sess, ok := val.(auth.Session)
	return sess, ok
}
