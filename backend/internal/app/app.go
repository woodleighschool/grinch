// Package app wires the application components and manages process lifecycle.
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"golang.org/x/sync/errgroup"

	"github.com/woodleighschool/grinch/internal/integra/entra"
	"github.com/woodleighschool/grinch/internal/jobs"
	"github.com/woodleighschool/grinch/internal/platform/config"
	"github.com/woodleighschool/grinch/internal/platform/db"
	"github.com/woodleighschool/grinch/internal/platform/httpserver"
	"github.com/woodleighschool/grinch/internal/platform/logging"
	"github.com/woodleighschool/grinch/internal/service"
	eventsrepo "github.com/woodleighschool/grinch/internal/store/events"
	groupsrepo "github.com/woodleighschool/grinch/internal/store/groups"
	machinesrepo "github.com/woodleighschool/grinch/internal/store/machines"
	membershipsrepo "github.com/woodleighschool/grinch/internal/store/memberships"
	policiesrepo "github.com/woodleighschool/grinch/internal/store/policies"
	rulesrepo "github.com/woodleighschool/grinch/internal/store/rules"
	usersrepo "github.com/woodleighschool/grinch/internal/store/users"
	httprouter "github.com/woodleighschool/grinch/internal/transport/http"
	apihttp "github.com/woodleighschool/grinch/internal/transport/http/api"
)

// App holds the long lived application components.
type App struct {
	Config      config.Config
	Log         *slog.Logger
	Server      *httpserver.Server
	RiverClient *river.Client[pgx.Tx]
	DBPool      *pgxpool.Pool
}

const (
	defaultQueueWorkers = 5
	riverStopTimeout    = 10 * time.Second
)

// New constructs the application and initialises its dependencies.
func New(ctx context.Context) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	log := logging.NewLogger(cfg.LogLevel)

	pool, err := db.Connect(ctx, db.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		Database: cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	})
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	// Ensure we close the pool on any early return.
	cleanupPool := true
	defer func() {
		if cleanupPool {
			pool.Close()
		}
	}()

	if err = runMigrations(ctx, pool, log); err != nil {
		return nil, err
	}

	services := buildServices(pool)

	authSvc, err := apihttp.NewAuthService(apihttp.AuthConfig{
		AppName:               "Grinch",
		SecretKey:             cfg.AuthSecret,
		BaseURL:               cfg.BaseURL,
		MicrosoftClientID:     cfg.MicrosoftClientID,
		MicrosoftClientSecret: cfg.MicrosoftClientSecret,
		MicrosoftTenantID:     cfg.MicrosoftTenantID,
		AdminPassword:         cfg.AdminPassword,
		TokenDuration:         cfg.TokenDuration,
		CookieDuration:        cfg.CookieDuration,
	})
	if err != nil {
		return nil, fmt.Errorf("create auth service: %w", err)
	}

	router, err := httprouter.NewRouter(httprouter.RouterConfig{
		API: apihttp.Services{
			Auth:        authSvc,
			Users:       services.Users,
			Groups:      services.Groups,
			Memberships: services.Memberships,
			Machines:    services.Machines,
			Events:      services.Events,
			Rules:       services.Rules,
			Policies:    services.Policies,
		},
		Sync:        services.Sync,
		Log:         log,
		FrontendDir: cfg.FrontendDir,
	})
	if err != nil {
		return nil, fmt.Errorf("create router: %w", err)
	}

	server := httpserver.New(router, cfg.Port, log)

	syncer, err := buildEntraSyncer(cfg, services.Users, services.Groups, log)
	if err != nil {
		return nil, fmt.Errorf("build entra syncer: %w", err)
	}

	riverClient, err := buildRiver(pool, services, syncer, cfg, log)
	if err != nil {
		return nil, fmt.Errorf("build river: %w", err)
	}

	cleanupPool = false

	return &App{
		Config:      cfg,
		Log:         log,
		Server:      server,
		RiverClient: riverClient,
		DBPool:      pool,
	}, nil
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool, log *slog.Logger) error {
	if err := db.Migrate(ctx, pool); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	driver := riverpgxv5.New(pool)

	migrator, err := rivermigrate.New(driver, &rivermigrate.Config{
		Logger: log,
	})
	if err != nil {
		return fmt.Errorf("river migrator: %w", err)
	}
	if _, err = migrator.Migrate(ctx, rivermigrate.DirectionUp, nil); err != nil {
		return fmt.Errorf("river migrate: %w", err)
	}

	return nil
}

func buildServices(pool *pgxpool.Pool) service.Services {
	stores := defaultStoreFactory(pool)
	return service.NewServices(stores)
}

func defaultStoreFactory(pool *pgxpool.Pool) service.Stores {
	return service.Stores{
		Users:       usersrepo.New(pool),
		Groups:      groupsrepo.New(pool),
		Memberships: membershipsrepo.New(pool),
		Machines:    machinesrepo.New(pool),
		Events:      eventsrepo.New(pool),
		Rules:       rulesrepo.New(pool),
		Policies:    policiesrepo.New(pool),
	}
}

func buildEntraSyncer(
	cfg config.Config,
	usersSvc entra.UserWriter,
	groupsSvc entra.GroupWriter,
	log *slog.Logger,
) (*entra.Syncer, error) {
	client, err := entra.NewClient(entra.Config{
		TenantID:     cfg.EntraTenantID,
		ClientID:     cfg.EntraClientID,
		ClientSecret: cfg.EntraClientSecret,
	})
	if err != nil {
		return nil, fmt.Errorf("create entra client: %w", err)
	}

	return entra.NewSyncer(client, usersSvc, groupsSvc, log), nil
}

func buildRiver(
	pool *pgxpool.Pool,
	services service.Services,
	syncer *entra.Syncer,
	cfg config.Config,
	log *slog.Logger,
) (*river.Client[pgx.Tx], error) {
	driver := riverpgxv5.New(pool)

	workers := river.NewWorkers()
	river.AddWorker(workers, &jobs.EntraSyncWorker{
		Syncer: syncer,
	})
	river.AddWorker(workers, &jobs.PolicyReconcileWorker{
		Policies: services.Policies,
	})
	river.AddWorker(workers, &jobs.PruneEventsWorker{
		Events:    services.Events,
		Retention: cfg.EventRetentionPeriod,
	})

	retention := cfg.EventRetentionPeriod

	periodicJobs := jobs.PeriodicJobs(jobs.RecurringConfig{
		EntraInterval:   cfg.EntraSyncInterval,
		PruneInterval:   cfg.EventPruneInterval,
		RetentionPeriod: retention,
	})

	client, err := river.NewClient(driver, &river.Config{
		Logger:       log,
		PeriodicJobs: periodicJobs,
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {
				MaxWorkers: defaultQueueWorkers,
			},
		},
		Workers: workers,
	})
	if err != nil {
		return nil, fmt.Errorf("river client: %w", err)
	}

	return client, nil
}

// Run starts background workers and runs the HTTP server until the context is canceled.
func (app *App) Run(ctx context.Context) error {
	ctx = logging.WithContext(ctx, app.Log)
	app.Log.InfoContext(ctx, "starting services", "addr", fmt.Sprintf(":%d", app.Config.Port))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	g, groupCtx := errgroup.WithContext(ctx)

	if app.RiverClient == nil {
		return errors.New("river client not configured")
	}

	g.Go(func() error {
		if err := app.RiverClient.Start(groupCtx); err != nil {
			return fmt.Errorf("start river: %w", err)
		}

		<-groupCtx.Done()

		stopCtx, stopCancel := context.WithTimeout(context.WithoutCancel(groupCtx), riverStopTimeout)
		defer stopCancel()

		if err := app.RiverClient.Stop(stopCtx); err != nil && !errors.Is(err, context.Canceled) {
			app.Log.WarnContext(stopCtx, "river shutdown", "error", err)
			return fmt.Errorf("stop river: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		return app.Server.Run(groupCtx)
	})

	err := g.Wait()

	if app.DBPool != nil {
		app.DBPool.Close()
	}

	return err
}
