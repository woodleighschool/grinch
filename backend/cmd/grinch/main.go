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

	appentrasync "github.com/woodleighschool/grinch/internal/app/entrasync"
	appevents "github.com/woodleighschool/grinch/internal/app/events"
	appfileaccessevents "github.com/woodleighschool/grinch/internal/app/fileaccessevents"
	appgroups "github.com/woodleighschool/grinch/internal/app/groups"
	appmemberships "github.com/woodleighschool/grinch/internal/app/memberships"
	apprules "github.com/woodleighschool/grinch/internal/app/rules"
	appsanta "github.com/woodleighschool/grinch/internal/app/santa"
	"github.com/woodleighschool/grinch/internal/config"
	"github.com/woodleighschool/grinch/internal/platform/logging"
	"github.com/woodleighschool/grinch/internal/store/postgres"
	apihttp "github.com/woodleighschool/grinch/internal/transport/http/api"
	authhttp "github.com/woodleighschool/grinch/internal/transport/http/auth"
	httprouter "github.com/woodleighschool/grinch/internal/transport/http/router"
	synchttp "github.com/woodleighschool/grinch/internal/transport/http/sync"
)

const (
	frontendDistDir   = "/frontend"
	idleTimeout       = 2 * time.Minute
	readHeaderTimeout = 5 * time.Second
	retentionInterval = 1 * time.Hour
	shutdownTimeout   = 10 * time.Second
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "grinch exited with error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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

	server, err := buildServer(ctx, logger, cfg, store)
	if err != nil {
		return err
	}

	if err = startEntraSync(ctx, logger, cfg, store); err != nil {
		return err
	}

	return serve(ctx, logger, server, cfg.HTTP.Port)
}

func buildServer(
	ctx context.Context,
	logger *slog.Logger,
	cfg config.Config,
	store *postgres.Store,
) (*http.Server, error) {
	groupService := appgroups.New(store)
	ruleService := apprules.New(store)
	membershipService := appmemberships.New(store)
	fileAccessEventService := appfileaccessevents.New(store)

	syncService := appsanta.New(
		logger,
		store,
		cfg.Events.DecisionAllowlist,
		ruleService,
	)

	eventService := appevents.New(logger, store, cfg.Events.RetentionDays)

	authService, err := authhttp.New(authhttp.Config{
		RootURL:            cfg.HTTP.BaseURL,
		EntraTenantID:      cfg.Auth.EntraTenantID,
		EntraClientID:      cfg.Auth.EntraClientID,
		EntraClientSecret:  cfg.Auth.EntraClientSecret,
		JWTSecret:          cfg.Auth.JWTSecret,
		LocalAdminPassword: cfg.Auth.LocalAdminPass,
	})
	if err != nil {
		return nil, fmt.Errorf("configure auth: %w", err)
	}

	syncHandler := synchttp.New(
		syncService,
	)

	apiHandler := apihttp.New(
		store,
		groupService,
		fileAccessEventService,
		ruleService,
		membershipService,
	)

	go eventService.RunRetention(ctx, retentionInterval)

	return &http.Server{
		Addr: cfg.HTTP.Addr(),
		Handler: httprouter.New(
			logger,
			store.Ping,
			syncHandler.RegisterRoutes,
			authService.RegisterRoutes,
			func(router chi.Router) {
				router.Use(authhttp.APIMiddleware(authService.SessionAuthMiddleware()))
				apiHandler.RegisterRoutes(router)
			},
			frontendDistDir,
		),
		ReadHeaderTimeout: readHeaderTimeout,
		IdleTimeout:       idleTimeout,
	}, nil
}

func startEntraSync(
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

	service := appentrasync.New(
		logger,
		graphClient,
		store,
		cfg.Entra.Interval,
	)

	go service.Run(ctx)

	return nil
}

func serve(
	ctx context.Context,
	logger *slog.Logger,
	server *http.Server,
	port int,
) error {
	serverErr := make(chan error, 1)

	go func() {
		logger.Info("starting server", "port", port)

		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}

		serverErr <- nil
	}()

	select {
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("serve: %w", err)
		}
		return nil

	case <-ctx.Done():
		logger.InfoContext(ctx, "shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	if err := <-serverErr; err != nil {
		return fmt.Errorf("serve after shutdown: %w", err)
	}

	logger.InfoContext(ctx, "server stopped")
	return nil
}
