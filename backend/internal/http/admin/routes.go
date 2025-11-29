package admin

import (
	"log/slog"

	"github.com/go-chi/chi/v5"

	"github.com/woodleighschool/grinch/internal/config"
	"github.com/woodleighschool/grinch/internal/rules"
	"github.com/woodleighschool/grinch/internal/store"
)

// Handler centralises admin HTTP handlers and shared dependencies.
type Handler struct {
	Store    *store.Store
	Logger   *slog.Logger
	Config   config.Config
	Compiler *rules.Compiler
}

// RegisterRoutes attaches all admin endpoints under /v1.
func RegisterRoutes(r chi.Router, cfg config.Config, store *store.Store, logger *slog.Logger, compiler *rules.Compiler) {
	h := Handler{Store: store, Logger: logger, Config: cfg, Compiler: compiler}
	r.Route("/v1", func(r chi.Router) {
		r.Route("/applications", h.applicationsRoutes)
		r.Route("/devices", h.devicesRoutes)
		r.Route("/users", h.usersRoutes)
		r.Route("/groups", h.groupsRoutes)
		r.Route("/events", h.eventsRoutes)
		r.Route("/files", h.filesRoutes)
		r.Route("/settings", h.settingsRoutes)
	})
}
