// Package httprouter provides HTTP routing for the Grinch server.
package httprouter

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/woodleighschool/grinch/internal/domain/santa"
	"github.com/woodleighschool/grinch/internal/logging"
	apihttp "github.com/woodleighschool/grinch/internal/transport/http/api"
	"github.com/woodleighschool/grinch/internal/transport/http/sync"
)

// RouterConfig holds dependencies for building the HTTP router.
type RouterConfig struct {
	API         apihttp.Services
	Sync        santa.SyncService
	Log         *slog.Logger
	FrontendDir string
}

// NewRouter builds the main HTTP router.
func NewRouter(cfg RouterConfig) (chi.Router, error) {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(logging.Middleware(cfg.Log))

	authHandler, avatarHandler := apihttp.AuthHandlers(cfg.API.Auth)
	r.Mount("/auth", authHandler)
	r.Mount("/avatar", avatarHandler)

	r.Mount("/api", apihttp.Router(cfg.API))
	r.Mount("/sync", sync.Router(cfg.Sync))

	spa, err := NewFrontendHandler(FrontendConfig{
		Dir:        cfg.FrontendDir,
		EnableGzip: true,
	})
	if err != nil {
		return nil, err
	}
	if spa != nil {
		r.Mount("/", spa)
	}

	return r, nil
}
