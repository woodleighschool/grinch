// Package config loads application configuration from environment variables.
package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
)

// Config holds application configuration.
type Config struct {
	Port        int    `env:"PORT" envDefault:"8080" validate:"gte=1,lte=65535"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
	BaseURL     string `env:"BASE_URL" envRequired:"true" validate:"required,url"`
	FrontendDir string `env:"FRONTEND_DIR"`

	DBHost     string `env:"DB_HOST" envRequired:"true" validate:"required"`
	DBPort     int    `env:"DB_PORT" envDefault:"5432" validate:"gte=1,lte=65535"`
	DBUser     string `env:"DB_USER" envRequired:"true" validate:"required"`
	DBPassword string `env:"DB_PASSWORD" envRequired:"true" validate:"required"`
	DBName     string `env:"DB_NAME" envRequired:"true" validate:"required"`
	DBSSLMode  string `env:"DB_SSLMODE" envDefault:"disable" validate:"oneof=disable require verify-ca verify-full"`

	AuthSecret        string        `env:"AUTH_SECRET" envRequired:"true" validate:"required,min=32"`
	TokenDuration     time.Duration `env:"TOKEN_DURATION" envDefault:"1h" validate:"gt=0"`
	CookieDuration    time.Duration `env:"COOKIE_DURATION" envDefault:"24h" validate:"gt=0"`
	AdminPassword     string        `env:"ADMIN_PASSWORD"`
	MicrosoftClientID string        `env:"MICROSOFT_CLIENT_ID"`
	MicrosoftSecret   string        `env:"MICROSOFT_CLIENT_SECRET"`

	EntraTenantID     string        `env:"ENTRA_TENANT_ID" envRequired:"true" validate:"required"`
	EntraClientID     string        `env:"ENTRA_CLIENT_ID" envRequired:"true" validate:"required"`
	EntraClientSecret string        `env:"ENTRA_CLIENT_SECRET" envRequired:"true" validate:"required"`
	EntraSyncInterval time.Duration `env:"ENTRA_SYNC_INTERVAL" envDefault:"15m" validate:"gt=0"`
}

// Load reads environment variables into a Config.
func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("validate config: %w", err)
	}
	return cfg, nil
}

// Validate runs basic validation on required and safety-critical fields.
func (c Config) Validate() error {
	v := validator.New(validator.WithRequiredStructEnabled())
	if err := v.Struct(c); err != nil {
		return err
	}
	if err := validateAuthProviders(c); err != nil {
		return err
	}
	return nil
}

func validateAuthProviders(c Config) error {
	hasAdmin := c.AdminPassword != ""
	hasMicrosoft := c.MicrosoftClientID != "" && c.MicrosoftSecret != ""
	if hasAdmin || hasMicrosoft {
		return nil
	}
	return fmt.Errorf("admin password or microsoft credentials must be set")
}
