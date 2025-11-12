package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	httpapi "github.com/woodleighschool/grinch/internal/http"

	"github.com/woodleighschool/grinch/internal/auth"
	"github.com/woodleighschool/grinch/internal/config"
	"github.com/woodleighschool/grinch/internal/graph"
	"github.com/woodleighschool/grinch/internal/rules"
	"github.com/woodleighschool/grinch/internal/store"
	"github.com/woodleighschool/grinch/internal/syncer"
)

var (
	buildVersion = "dev"
	gitCommit    = "unknown"
	buildDate    = "unknown"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := newLogger(cfg.LogLevel)
	buildInfo := httpapi.BuildInfo{
		Version:   buildVersion,
		GitCommit: gitCommit,
		BuildDate: buildDate,
	}
	logger.Info("starting grinch",
		"version", buildInfo.Version,
		"commit", buildInfo.GitCommit,
		"build_date", buildInfo.BuildDate,
	)

	db, err := store.Open(ctx, store.Options{
		URL:             cfg.DatabaseURL(),
		MaxConnections:  cfg.MaxConnections,
		MinConnections:  cfg.MinConnections,
		MaxConnLifetime: cfg.MaxConnLifetime,
	})
	if err != nil {
		logger.Error("connect db", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Migrate(ctx); err != nil {
		logger.Error("run migrations", "err", err)
		os.Exit(1)
	}

	oidcProvider, err := auth.NewOIDCProvider(ctx, cfg.AdminIssuer, cfg.AdminClientID, cfg.AdminClientSecret, cfg.SiteBaseURL)
	if err != nil {
		logger.Error("oidc provider", "err", err)
		os.Exit(1)
	}

	sessions, err := auth.NewSessionManager(cfg.SessionCookieName, cfg.SessionSecret, strings.HasPrefix(cfg.SiteBaseURL, "https"))
	if err != nil {
		logger.Error("session manager", "err", err)
		os.Exit(1)
	}

	compiler := rules.NewCompiler()

	scheduler := syncer.NewScheduler(logger)
	graphClient, err := graph.NewClient(ctx, cfg.GraphTenantID, cfg.GraphClientID, cfg.GraphClientSecret)
	if err != nil {
		logger.Warn("graph client", "err", err)
	}
	if graphClient != nil && graphClient.Enabled() {
		if err := scheduler.Add(cfg.SyncCron, "entra-users", syncer.NewUserJob(db, graphClient, logger)); err != nil {
			logger.Warn("schedule users", "err", err)
		}
		if err := scheduler.Add(cfg.SyncCron, "entra-groups", syncer.NewGroupJob(db, graphClient, logger)); err != nil {
			logger.Warn("schedule groups", "err", err)
		}
	}
	if err := scheduler.Add("@every 10m", "rule-compiler", syncer.NewRuleCompilerJob(db, compiler, logger)); err != nil {
		logger.Warn("schedule compiler", "err", err)
	}
	scheduler.Start()
	defer scheduler.Stop()

	deps := httpapi.Deps{
		Store:         db,
		Logger:        logger,
		Sessions:      sessions,
		SantaCompiler: compiler,
		OIDCProvider:  oidcProvider,
		BuildInfo:     buildInfo,
	}
	router := httpapi.NewRouter(cfg, deps)

	server := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown", "err", err)
	}
}

func newLogger(level string) *slog.Logger {
	lvl := new(slog.LevelVar)
	switch strings.ToLower(level) {
	case "debug":
		lvl.Set(slog.LevelDebug)
	case "warn":
		lvl.Set(slog.LevelWarn)
	case "error":
		lvl.Set(slog.LevelError)
	default:
		lvl.Set(slog.LevelInfo)
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	return slog.New(handler)
}
