// Package app wires the application components and manages process lifecycle.
package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/woodleighschool/grinch/internal/config"
	"github.com/woodleighschool/grinch/internal/domain/events"
	"github.com/woodleighschool/grinch/internal/domain/groups"
	"github.com/woodleighschool/grinch/internal/domain/machines"
	"github.com/woodleighschool/grinch/internal/domain/memberships"
	"github.com/woodleighschool/grinch/internal/domain/policies"
	"github.com/woodleighschool/grinch/internal/domain/rules"
	"github.com/woodleighschool/grinch/internal/domain/santa"
	"github.com/woodleighschool/grinch/internal/domain/users"
	"github.com/woodleighschool/grinch/internal/integra/entra"
	"github.com/woodleighschool/grinch/internal/logging"
	"github.com/woodleighschool/grinch/internal/store/db"
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
	Server      *httprouter.Server
	EntraRunner *entra.Runner
}

// New constructs the application and initialises its dependencies.
func New(ctx context.Context) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	log := logging.NewLogger(cfg.LogLevel)
	slog.SetDefault(log)

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

	if err = db.Migrate(ctx, pool); err != nil {
		return nil, fmt.Errorf("migrate database: %w", err)
	}

	services := buildServices(pool)

	authSvc, err := apihttp.NewAuthService(apihttp.AuthConfig{
		AppName:               "Grinch",
		SecretKey:             cfg.AuthSecret,
		BaseURL:               cfg.BaseURL,
		MicrosoftClientID:     cfg.MicrosoftClientID,
		MicrosoftClientSecret: cfg.MicrosoftSecret,
		AdminPassword:         cfg.AdminPassword,
		TokenDuration:         cfg.TokenDuration,
		CookieDuration:        cfg.CookieDuration,
	})
	if err != nil {
		return nil, fmt.Errorf("create auth service: %w", err)
	}

	santaSvc := santa.NewSyncService(
		services.machines,
		services.policies,
		services.rules,
		services.events,
	)

	router, err := httprouter.NewRouter(httprouter.RouterConfig{
		API: apihttp.Services{
			Auth:        authSvc,
			Users:       services.users,
			Groups:      services.groups,
			Memberships: services.memberships,
			Machines:    services.machines,
			Events:      services.events,
			Rules:       services.rules,
			Policies:    services.policies,
		},
		Sync:        santaSvc,
		Log:         log,
		FrontendDir: cfg.FrontendDir,
	})
	if err != nil {
		return nil, fmt.Errorf("create router: %w", err)
	}

	server := httprouter.NewServer(router, cfg.Port, log)

	entraRunner, err := buildEntraRunner(cfg, services.users, services.groups, log)
	if err != nil {
		return nil, err
	}

	return &App{
		Config:      cfg,
		Log:         log,
		Server:      server,
		EntraRunner: entraRunner,
	}, nil
}

type appServices struct {
	users       users.Service
	groups      groups.Service
	memberships memberships.Service
	machines    machines.Service
	events      events.Service
	rules       rules.Service
	policies    policies.Service
}

func buildServices(pool *pgxpool.Pool) appServices {
	usersRepo := usersrepo.New(pool)
	groupsRepo := groupsrepo.New(pool)
	membershipsRepo := membershipsrepo.New(pool)
	machinesRepo := machinesrepo.New(pool)
	eventsRepo := eventsrepo.New(pool)
	rulesRepo := rulesrepo.New(pool)
	policiesRepo := policiesrepo.New(pool)

	usersSvc := users.NewService(usersRepo)
	membershipsSvc := memberships.NewService(membershipsRepo)
	eventsSvc := events.NewService(eventsRepo)
	machinesSvc := machines.NewService(machinesRepo, usersSvc)

	// The policies service needs a reconciler that depends on the policies service.
	policiesSvc := policies.NewService(policiesRepo, membershipsSvc, nil)
	rulesSvc := rules.NewService(rulesRepo, policiesSvc)

	reconciler := policies.NewReconciler(machinesSvc, policiesSvc)
	policiesSvc = policies.NewService(policiesRepo, membershipsSvc, reconciler)
	groupsSvc := groups.NewService(groupsRepo, reconciler)

	return appServices{
		users:       usersSvc,
		groups:      groupsSvc,
		memberships: membershipsSvc,
		machines:    machinesSvc,
		events:      eventsSvc,
		rules:       rulesSvc,
		policies:    policiesSvc,
	}
}

func buildEntraRunner(
	cfg config.Config,
	usersSvc users.Service,
	groupsSvc groups.Service,
	log *slog.Logger,
) (*entra.Runner, error) {
	client, err := entra.NewClient(entra.Config{
		TenantID:     cfg.EntraTenantID,
		ClientID:     cfg.EntraClientID,
		ClientSecret: cfg.EntraClientSecret,
	})
	if err != nil {
		return nil, fmt.Errorf("create entra client: %w", err)
	}

	syncer := entra.NewSyncer(client, usersSvc, groupsSvc, log)
	return entra.NewRunner(syncer, cfg.EntraSyncInterval, log), nil
}

// Run starts background workers and runs the HTTP server until the context is canceled.
func (app *App) Run(ctx context.Context) error {
	app.Log.InfoContext(ctx, "starting server", "port", app.Config.Port)

	go app.EntraRunner.Start(ctx)

	return app.Server.Run(ctx)
}
