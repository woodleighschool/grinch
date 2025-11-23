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

	if cfg.AdminListenAddr == cfg.SantaListenAddr {
		logger.Error("admin and santa listen addresses must differ", "addr", cfg.AdminListenAddr)
		os.Exit(1)
	}

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
		if err := scheduler.Add(cfg.SyncCron, "entra-users", 15*time.Minute, syncer.NewUserJob(db, graphClient, logger)); err != nil {
			logger.Warn("schedule users", "err", err)
		}
		if err := scheduler.Add(cfg.SyncCron, "entra-groups", 15*time.Minute, syncer.NewGroupJob(db, graphClient, logger)); err != nil {
			logger.Warn("schedule groups", "err", err)
		}
	}
	if err := scheduler.Add("@every 10m", "rule-compiler", 5*time.Minute, syncer.NewRuleCompilerJob(db, compiler, logger)); err != nil {
		logger.Warn("schedule compiler", "err", err)
	}
	scheduler.Start()
	defer scheduler.Stop()

	adminRouter := httpapi.NewAdminRouter(cfg, httpapi.AdminDeps{
		Store:         db,
		Logger:        logger,
		Sessions:      sessions,
		SantaCompiler: compiler,
		OIDCProvider:  oidcProvider,
		BuildInfo:     buildInfo,
	})
	santaRouter := httpapi.NewSantaRouter(httpapi.SantaDeps{
		Store:     db,
		Logger:    logger,
		Compiler:  compiler,
		BuildInfo: buildInfo,
	})

	adminServer := newHTTPServer(cfg.AdminListenAddr, adminRouter)
	santaServer := newHTTPServer(cfg.SantaListenAddr, santaRouter)

	adminErrCh := startServer(logger, "admin", adminServer)
	santaErrCh := startServer(logger, "santa", santaServer)

	var serveErr error

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-adminErrCh:
		if err != nil {
			logger.Error("admin server error", "err", err)
			serveErr = err
		}
	case err := <-santaErrCh:
		if err != nil {
			logger.Error("santa server error", "err", err)
			serveErr = err
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := shutdownServer(shutdownCtx, logger, "admin", adminServer); err != nil {
		serveErr = err
	}
	if err := shutdownServer(shutdownCtx, logger, "santa", santaServer); err != nil {
		serveErr = err
	}

	// Ensure server goroutines exit before leaving main.
	<-adminErrCh
	<-santaErrCh

	if serveErr != nil {
		os.Exit(1)
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

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func startServer(logger *slog.Logger, name string, server *http.Server) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		logger.Info("listening", "server", name, "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()
	return errCh
}

func shutdownServer(ctx context.Context, logger *slog.Logger, name string, server *http.Server) error {
	if server == nil {
		return nil
	}
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "server", name, "err", err)
		return err
	}
	logger.Info("server stopped", "server", name)
	return nil
}
