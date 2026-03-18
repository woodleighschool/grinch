package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	graphsync "github.com/woodleighschool/go-entrasync"

	appauth "github.com/woodleighschool/grinch/internal/app/auth"
	appentrasync "github.com/woodleighschool/grinch/internal/app/entrasync"
	appevents "github.com/woodleighschool/grinch/internal/app/events"
	appgroupmemberships "github.com/woodleighschool/grinch/internal/app/groupmemberships"
	apprules "github.com/woodleighschool/grinch/internal/app/rules"
	appsanta "github.com/woodleighschool/grinch/internal/app/santa"
	"github.com/woodleighschool/grinch/internal/config"
	"github.com/woodleighschool/grinch/internal/platform/logging"
	"github.com/woodleighschool/grinch/internal/store/postgres"
	adminpostgres "github.com/woodleighschool/grinch/internal/store/postgres/admin"
	entrasyncpostgres "github.com/woodleighschool/grinch/internal/store/postgres/entrasync"
	groupmembershipspostgres "github.com/woodleighschool/grinch/internal/store/postgres/groupmemberships"
	rulespostgres "github.com/woodleighschool/grinch/internal/store/postgres/rules"
	santapostgres "github.com/woodleighschool/grinch/internal/store/postgres/santa"
	authhttp "github.com/woodleighschool/grinch/internal/transport/http/auth"
	httpapi "github.com/woodleighschool/grinch/internal/transport/http/httpapi"
	httprouter "github.com/woodleighschool/grinch/internal/transport/http/router"
	"github.com/woodleighschool/grinch/internal/transport/http/synchttp"
)

const (
	shutdownTimeout   = 10 * time.Second
	readHeaderTimeout = 5 * time.Second
	idleTimeout       = 2 * time.Minute
	retentionInterval = 1 * time.Hour
	frontendDistDir   = "/frontend"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "grinch exited with error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := logging.New(cfg.Logging.Level)
	if err != nil {
		return fmt.Errorf("configure logger: %w", err)
	}
	slog.SetDefault(logger)

	store, err := postgres.New(context.Background(), cfg.Database)
	if err != nil {
		return fmt.Errorf("create store: %w", err)
	}
	defer store.Close()

	eventService, server, err := buildServer(logger, cfg, store)
	if err != nil {
		return err
	}

	stopContext, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if syncErr := maybeStartEntraSync(stopContext, logger, cfg, store); syncErr != nil {
		return syncErr
	}
	maybeStartEventRetention(stopContext, eventService)

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("starting server", "port", cfg.HTTP.Port)
		if serveErr := server.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			serverErr <- serveErr
			return
		}
		serverErr <- nil
	}()

	select {
	case serveErr := <-serverErr:
		if serveErr != nil {
			return fmt.Errorf("serve: %w", serveErr)
		}
		return nil
	case <-stopContext.Done():
		logger.Info("shutdown signal received")
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if shutdownErr := server.Shutdown(shutdownContext); shutdownErr != nil {
		return fmt.Errorf("shutdown: %w", shutdownErr)
	}

	if serveErr := <-serverErr; serveErr != nil {
		return fmt.Errorf("serve after shutdown: %w", serveErr)
	}

	logger.Info("server stopped")
	return nil
}

func buildServer(
	logger *slog.Logger,
	cfg config.Config,
	store *postgres.Store,
) (*appevents.Service, *http.Server, error) {
	adminStore := adminpostgres.New(store)
	ruleService := apprules.New(rulespostgres.New(store))
	groupMembershipService := appgroupmemberships.New(groupmembershipspostgres.New(store))
	santaStore := santapostgres.New(store)
	syncService := appsanta.New(
		logger,
		santaStore,
		cfg.Events,
		ruleService,
	)
	eventService := appevents.New(logger, santaStore, cfg.Events.RetentionDays)
	syncHandler := synchttp.New(
		syncService,
		cfg.Sync.SharedSecret,
	)
	authService, err := appauth.New(appauth.Config{
		RootURL:            cfg.HTTP.BaseURL,
		EntraTenantID:      cfg.Auth.EntraTenantID,
		EntraClientID:      cfg.Auth.EntraClientID,
		EntraClientSecret:  cfg.Auth.EntraClientSecret,
		JWTSecret:          cfg.Auth.JWTSecret,
		LocalAdminPassword: cfg.Auth.LocalAdminPass,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("configure auth: %w", err)
	}

	apiAuthMiddleware := authhttp.NewAPIMiddleware(
		authService.SessionAuthMiddleware(),
	)
	apiHandler := httpapi.New(adminStore, ruleService, groupMembershipService)

	server := &http.Server{
		Handler: httprouter.New(
			logger,
			store.Ping,
			syncHandler.RegisterRoutes,
			authService.RegisterRoutes,
			newAPIRouteRegistrar(apiAuthMiddleware, apiHandler),
			frontendDistDir,
		),
		ReadHeaderTimeout: readHeaderTimeout,
		IdleTimeout:       idleTimeout,
		Addr:              cfg.HTTP.Addr(),
	}

	return eventService, server, nil
}

func maybeStartEntraSync(
	ctx context.Context,
	logger *slog.Logger,
	cfg config.Config,
	store *postgres.Store,
) error {
	if !cfg.Entra.Enabled {
		return nil
	}

	graphClient, err := graphsync.NewClient(graphsync.Config{
		TenantID:     cfg.Auth.EntraTenantID,
		ClientID:     cfg.Auth.EntraClientID,
		ClientSecret: cfg.Auth.EntraClientSecret,
	})
	if err != nil {
		return fmt.Errorf("configure entra graph client: %w", err)
	}

	entraSyncService := appentrasync.New(
		logger,
		graphClient,
		entrasyncpostgres.New(store),
		cfg.Entra.Interval,
	)

	go entraSyncService.Run(ctx)
	return nil
}

func newAPIRouteRegistrar(middleware func(http.Handler) http.Handler, handler *httpapi.Server) func(chi.Router) {
	return func(router chi.Router) {
		router.Use(middleware)
		handler.RegisterRoutes(router)
	}
}

func maybeStartEventRetention(ctx context.Context, eventService *appevents.Service) {
	go eventService.RunRetention(ctx, retentionInterval)
}
