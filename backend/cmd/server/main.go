package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/woodleighschool/grinch/backend/internal/api"
	"github.com/woodleighschool/grinch/backend/internal/auth"
	"github.com/woodleighschool/grinch/backend/internal/config"
	"github.com/woodleighschool/grinch/backend/internal/db"
	"github.com/woodleighschool/grinch/backend/internal/entra"
	"github.com/woodleighschool/grinch/backend/internal/events"
	"github.com/woodleighschool/grinch/backend/internal/santa"
	"github.com/woodleighschool/grinch/backend/internal/store"
	"github.com/woodleighschool/grinch/backend/internal/sync"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		slog.Error("command execution failed", "error", err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	var flagLogLevel string

	cmd := &cobra.Command{
		Use:           "grinch",
		Short:         "Grinch Santa sync server",
		Long:          "Grinch coordinates Santa event ingestion and synchronisation services.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			baseLogger := slog.Default().With("command", cmd.Name())
			baseLogger.Debug("loading configuration from environment")
			cfg, err := config.Load()
			if err != nil {
				baseLogger.Error("failed to load configuration", "error", err)
				return fmt.Errorf("load config: %w", err)
			}

			if flagLogLevel != "" {
				baseLogger.Debug("overriding log level from flag", "log_level", flagLogLevel)
				cfg.LogLevel = flagLogLevel
			}

			logger := setupLogging(cfg)
			logger.Debug("runtime configuration resolved",
				"server_address", cfg.ServerAddress,
				"frontend_dist", cfg.FrontendDistDir,
				"database_host", cfg.DatabaseHost,
				"database_name", cfg.DatabaseName,
				"sync_interval", cfg.SyncInterval,
				"allowed_origins", cfg.AllowedOrigins,
				"metrics_enabled", cfg.EnableMetrics,
			)
			logger.Info("starting server",
				"version", version,
				"commit", commit,
				"log_level", cfg.LogLevel,
				"address", cfg.ServerAddress,
				"frontend_dist", cfg.FrontendDistDir,
			)

			ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			if err := runServer(ctx, cfg, logger); err != nil {
				return err
			}

			logger.Info("shutdown complete")
			return nil
		},
	}

	cmd.SetContext(context.Background())
	cmd.Flags().StringVar(&flagLogLevel, "log-level", "", "logging level: debug, info, warn, error")
	cmd.AddCommand(newVersionCmd())

	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show build information",
		Run: func(cmd *cobra.Command, _ []string) {
			logger := slog.Default().With(
				"command", cmd.Name(),
			)
			logger.Info("grinch version information",
				"version", version,
				"commit", commit,
				"built", date,
			)
		},
	}
}

func runServer(ctx context.Context, cfg *config.Config, logger *slog.Logger) error {
	poolOpts := db.PoolOptions{
		MaxConns:          cfg.DatabaseMaxConns,
		MinConns:          cfg.DatabaseMinConns,
		MaxConnLifetime:   cfg.DatabaseMaxConnLifetime,
		MaxConnIdleTime:   cfg.DatabaseMaxConnIdleTime,
		HealthCheckPeriod: cfg.DatabaseHealthCheckEvery,
	}

	logger.Debug("configuring database connection",
		"host", cfg.DatabaseHost,
		"port", cfg.DatabasePort,
		"name", cfg.DatabaseName,
	)
	pool, err := db.Connect(ctx, cfg.GetDatabaseURL(), db.WithPoolOptions(poolOpts))
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}
	defer pool.Close()
	logger.Debug("database connection established")

	if err := db.Migrate(ctx, pool); err != nil {
		return fmt.Errorf("migrate db: %w", err)
	}
	logger.Info("database migrations completed")

	store := store.New(pool, cfg)
	logger.Debug("store initialised")

	// Initialise admin user if configured
	if cfg.InitialAdminPassword != "" {
		logger.Debug("ensuring configured initial admin user")
		if err := store.EnsureInitialAdminUser(ctx, cfg.InitialAdminPassword); err != nil {
			logger.Warn("failed to initialise admin user", "error", err)
		} else {
			logger.Info("initial admin user ensured", "principal", "admin")
		}
	}

	sessionManager := auth.NewSessionManager([]byte(cfg.CookieSecret), cfg.CookieName)
	sessionManager.SetUserStore(store)
	logger.Debug("session manager configured",
		"cookie_name", cfg.CookieName,
		"secure", true,
	)
	broadcaster := events.NewBroadcaster()
	santaSvc := santa.NewService(store)
	logger.Debug("santa service initialised")

	server, err := api.NewServer(ctx, cfg, store, santaSvc, sessionManager, broadcaster, logger)
	if err != nil {
		return fmt.Errorf("init server: %w", err)
	}
	logger.Debug("api server initialised")

	entraSvc, err := entra.NewService(cfg, store, logger)
	if err != nil {
		return fmt.Errorf("init entra service: %w", err)
	}
	logger.Debug("entra service initialised")

	entraSvc.Start(ctx, cfg.SyncInterval)
	logger.Info("entra sync service started", "interval", cfg.SyncInterval)

	// Start periodic sync service
	periodicSync := sync.NewPeriodicSyncService(store, entraSvc, logger)
	go periodicSync.Start(ctx)
	logger.Info("periodic sync service started")

	rootHandler := server.Routes()
	if cfg.FrontendDistDir != "" {
		rootHandler = mountStatic(cfg.FrontendDistDir, rootHandler)
	}

	srv := &http.Server{
		Addr:              cfg.ServerAddress,
		Handler:           rootHandler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		logger.Info("shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("server shutdown error", "error", err)
			return
		}

		logger.Info("http server shut down gracefully")
	}()

	logger.Info("server listening", "address", cfg.ServerAddress)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("listen: %w", err)
	}
	logger.Debug("http server listener exited")

	return nil
}

func setupLogging(cfg *config.Config) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.GetLogLevel(),
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	logger.Info("log level configured", "level", cfg.LogLevel)
	return logger
}

func mountStatic(distDir string, apiHandler http.Handler) http.Handler {
	fs := http.FileServer(http.Dir(distDir))
	indexPath := filepath.Join(distDir, "index.html")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") || strings.HasPrefix(r.URL.Path, "/santa") || r.URL.Path == "/healthz" {
			apiHandler.ServeHTTP(w, r)
			return
		}
		candidate := filepath.Join(distDir, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, indexPath)
	})
}
