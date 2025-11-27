package httpapi

import (
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/woodleighschool/grinch/internal/auth"
	"github.com/woodleighschool/grinch/internal/config"
	"github.com/woodleighschool/grinch/internal/http/admin"
	authhttp "github.com/woodleighschool/grinch/internal/http/auth"
	"github.com/woodleighschool/grinch/internal/http/santa"
	"github.com/woodleighschool/grinch/internal/rules"
	"github.com/woodleighschool/grinch/internal/store"
)

// AdminDeps contains dependencies for the internal admin API + UI.
type AdminDeps struct {
	Store         *store.Store
	Logger        *slog.Logger
	Sessions      *auth.SessionManager
	SantaCompiler *rules.Compiler
	OIDCProvider  *auth.OIDCProvider
	BuildInfo     BuildInfo
}

// SantaDeps contains dependencies for the Santa sync API.
type SantaDeps struct {
	Store     *store.Store
	Logger    *slog.Logger
	Compiler  *rules.Compiler
	BuildInfo BuildInfo
}

// BuildInfo is mirrored back to HTTP clients for display.
type BuildInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildDate string `json:"build_date"`
}

// NewAdminRouter wires the admin API, auth routes, and optional static UI.
func NewAdminRouter(cfg config.Config, deps AdminDeps) http.Handler {
	r := baseRouter()

	r.Get("/api/v1/status", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":  "ok",
			"version": deps.BuildInfo,
		})
	})

	api := chi.NewRouter()
	api.Use(AdminAuth(deps.Sessions, deps.Logger))
	admin.RegisterRoutes(api, cfg, deps.Store, deps.Logger, deps.SantaCompiler)
	r.Mount("/api", api)

	authRoutes := chi.NewRouter()
	authhttp.RegisterRoutes(authRoutes, cfg, deps.OIDCProvider, deps.Sessions, deps.Logger)
	r.Mount("/api/auth", authRoutes)

	handler := http.Handler(r)
	if cfg.FrontendDistDir != "" {
		handler = mountStatic(cfg.FrontendDistDir, handler)
	}
	return handler
}

// NewSantaRouter exposes the limited Santa sync endpoints.
func NewSantaRouter(deps SantaDeps) http.Handler {
	r := baseRouter()

	santaRouter := chi.NewRouter()
	santa.RegisterRoutes(santaRouter, santa.Dependencies{
		Store:    deps.Store,
		Logger:   deps.Logger,
		Compiler: deps.Compiler,
	})
	r.Mount("/santa", santaRouter)

	return r
}

// baseRouter applies the common middleware stack used by both routers.
func baseRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	return r
}

// mountStatic serves the compiled frontend unless an API/Santa path was requested.
func mountStatic(distDir string, apiHandler http.Handler) http.Handler {
	fileServer := http.FileServer(http.Dir(distDir))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/santa/") {
			apiHandler.ServeHTTP(w, r)
			return
		}

		reqPath := path.Clean("/" + r.URL.Path)
		fullPath := filepath.Join(distDir, strings.TrimPrefix(reqPath, "/"))
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			r.URL.Path = "/"
		}

		fileServer.ServeHTTP(w, r)
	})
}
