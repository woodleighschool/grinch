package admin

import (
	"log/slog"

	"github.com/go-chi/chi/v5"

	"github.com/woodleighschool/grinch/internal/config"
	"github.com/woodleighschool/grinch/internal/store"
)

type Handler struct {
	Store  *store.Store
	Logger *slog.Logger
	Config config.Config
}

func RegisterRoutes(r chi.Router, cfg config.Config, store *store.Store, logger *slog.Logger) {
	h := Handler{Store: store, Logger: logger, Config: cfg}
	r.Route("/users", h.usersRoutes)
	r.Route("/groups", h.groupsRoutes)
	r.Route("/rules", h.rulesRoutes)
	r.Route("/apps", h.appsRoutes)
	r.Route("/scopes", h.scopesRoutes)
	r.Route("/machines", h.machinesRoutes)
	r.Route("/events", h.eventsRoutes)
	r.Route("/settings", h.settingsRoutes)
}
