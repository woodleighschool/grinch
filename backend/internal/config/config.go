package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	AdminListenAddr      string        `env:"ADMIN_LISTEN_ADDR" envDefault:":8080"`
	SantaListenAddr      string        `env:"SANTA_LISTEN_ADDR" envDefault:":8081"`
	DatabaseHost         string        `env:"DATABASE_HOST,required"`
	DatabasePort         string        `env:"DATABASE_PORT" envDefault:"5432"`
	DatabaseName         string        `env:"DATABASE_NAME,required"`
	DatabaseUser         string        `env:"DATABASE_USER,required"`
	DatabasePassword     string        `env:"DATABASE_PASSWORD,required"`
	DatabaseSSLMode      string        `env:"DATABASE_SSLMODE" envDefault:"disable"`
	MaxConnLifetime      time.Duration `env:"DB_MAX_CONN_LIFETIME" envDefault:"30m"`
	MaxConnections       int32         `env:"DB_MAX_CONNECTIONS" envDefault:"10"`
	MinConnections       int32         `env:"DB_MIN_CONNECTIONS" envDefault:"2"`
	AdminIssuer          string        `env:"ADMIN_OIDC_ISSUER,required"`
	AdminClientID        string        `env:"ADMIN_OIDC_CLIENT_ID,required"`
	AdminClientSecret    string        `env:"ADMIN_OIDC_CLIENT_SECRET,required"`
	SessionSecret        string        `env:"SESSION_SECRET,required"`
	SessionCookieName    string        `env:"SESSION_COOKIE_NAME" envDefault:"grinch_session"`
	InitialAdminPassword string        `env:"INITIAL_ADMIN_PASSWORD"`
	SyncCron             string        `env:"SYNC_CRON" envDefault:"@every 5m"`
	SiteBaseURL          string        `env:"SITE_BASE_URL,required"`
	GraphTenantID        string        `env:"GRAPH_TENANT_ID"`
	GraphClientID        string        `env:"GRAPH_CLIENT_ID"`
	GraphClientSecret    string        `env:"GRAPH_CLIENT_SECRET"`
	LogLevel             string        `env:"LOG_LEVEL" envDefault:"info"`
	FrontendDistDir      string        `env:"FRONTEND_DIST_DIR"`
}

func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

func (c Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DatabaseUser, c.DatabasePassword, c.DatabaseHost, c.DatabasePort, c.DatabaseName, c.DatabaseSSLMode)
}
