package router

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/woodleighschool/grinch/internal/platform/httpmiddleware"
)

const readinessTimeout = 2 * time.Second

type statusResponse struct {
	Status string `json:"status"`
}

func New(
	logger *slog.Logger,
	readinessCheck func(context.Context) error,
	registerSyncRoutes func(chi.Router),
	registerAuthRoutes func(chi.Router),
	registerAPIRoutes func(chi.Router),
	frontendDir string,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(httpmiddleware.RequestLogger(logger))
	r.Use(middleware.Recoverer)

	r.Get("/healthz", healthHandler(logger))
	r.Get("/readyz", readinessHandler(logger, readinessCheck))

	r.Route("/sync", func(r chi.Router) {
		registerSyncRoutes(r)
		r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
	})
	r.Route("/auth", registerAuthRoutes)
	r.Route("/api/v1", registerAPIRoutes)

	mountFrontend(r, frontendDir)

	return r
}

func healthHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, logger, http.StatusOK, statusResponse{Status: "ok"})
	}
}

func readinessHandler(logger *slog.Logger, readinessCheck func(context.Context) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if readinessCheck == nil {
			writeJSON(w, logger, http.StatusOK, statusResponse{Status: "ready"})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), readinessTimeout)
		defer cancel()

		if err := readinessCheck(ctx); err != nil {
			logger.Warn("readiness check failed", "error", err)
			writeJSON(w, logger, http.StatusServiceUnavailable, statusResponse{Status: "not_ready"})
			return
		}

		writeJSON(w, logger, http.StatusOK, statusResponse{Status: "ready"})
	}
}

func writeJSON(w http.ResponseWriter, logger *slog.Logger, statusCode int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.Error("encode JSON response", "error", err)
	}
}
