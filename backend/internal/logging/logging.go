// Package logging provides structured logging and request logging middleware.
package logging

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/httplog/v3"
)

// Logger is the application logger type.
type Logger = *slog.Logger

type contextKey struct{}

// NewLogger builds a JSON logger configured for the provided level.
func NewLogger(level string) Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: lvl,
	}))
}

// FromContext returns the logger stored in ctx or slog.Default when absent.
func FromContext(ctx context.Context) Logger {
	if ctx == nil {
		return slog.Default()
	}
	if log, ok := ctx.Value(contextKey{}).(Logger); ok {
		return log
	}
	return slog.Default()
}

// WithContext returns a copy of ctx with log stored for downstream retrieval.
func WithContext(ctx context.Context, log Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, log)
}

// Middleware returns an HTTP middleware that logs requests and recovers panics.
func Middleware(log Logger) func(http.Handler) http.Handler {
	requestLogger := httplog.RequestLogger(log, &httplog.Options{
		Level:         slog.LevelInfo,
		RecoverPanics: true,
	})

	return func(next http.Handler) http.Handler {
		return requestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := WithContext(r.Context(), log)
			next.ServeHTTP(w, r.WithContext(ctx))
		}))
	}
}
