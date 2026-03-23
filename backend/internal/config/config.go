package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"

	"github.com/woodleighschool/grinch/internal/domain"
)

type Config struct {
	HTTP     HTTPConfig
	Logging  LoggingConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Sync     SyncConfig
	Entra    EntraSyncConfig
	Events   EventsConfig
}

type HTTPConfig struct {
	Port    int    `env:"GRINCH_PORT"     envDefault:"8080"`
	BaseURL string `env:"GRINCH_BASE_URL"`
}

func (c HTTPConfig) Addr() string {
	return fmt.Sprintf(":%d", c.Port)
}

type LoggingConfig struct {
	Level string `env:"LOG_LEVEL" envDefault:"info"`
}

type DatabaseConfig struct {
	Host     string `env:"DATABASE_HOST"`
	Port     int    `env:"DATABASE_PORT"     envDefault:"5432"`
	User     string `env:"DATABASE_USER"`
	Password string `env:"DATABASE_PASSWORD"`
	Name     string `env:"DATABASE_NAME"`
	SSLMode  string `env:"DATABASE_SSLMODE"  envDefault:"disable"`
}

type AuthConfig struct {
	EntraTenantID     string `env:"ENTRA_TENANT_ID"`
	EntraClientID     string `env:"ENTRA_CLIENT_ID"`
	EntraClientSecret string `env:"ENTRA_CLIENT_SECRET"`
	JWTSecret         string `env:"JWT_SECRET"`
	LocalAdminPass    string `env:"LOCAL_ADMIN_PASSWORD"`
}

type SyncConfig struct {
	SharedSecret string `env:"SYNC_SHARED_SECRET"`
}

type EntraSyncConfig struct {
	Enabled  bool          `env:"ENTRA_SYNC_ENABLED"  envDefault:"false"`
	Interval time.Duration `env:"ENTRA_SYNC_INTERVAL" envDefault:"1h"`
}

type EventsConfig struct {
	RetentionDays     int                        `env:"EVENT_RETENTION_DAYS" envDefault:"90"`
	DecisionAllowlist []domain.ExecutionDecision `env:"-"`
}

type envVar struct {
	name  string
	value string
}

func LoadFromEnv() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse env: %w", err)
	}

	allowlist, err := parseDecisionAllowlist(os.Getenv("EVENT_DECISION_ALLOWLIST"))
	if err != nil {
		return Config{}, err
	}
	cfg.Events.DecisionAllowlist = allowlist

	if err = validateConfig(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func validateConfig(cfg Config) error {
	var problems []string

	problems = append(problems, validateHTTP(cfg.HTTP)...)
	problems = append(problems, validateLogging(cfg.Logging)...)
	problems = append(problems, validateDatabase(cfg.Database)...)
	problems = append(problems, validateAuth(cfg.HTTP, cfg.Auth)...)
	problems = append(problems, validateEntraSync(cfg.Auth, cfg.Entra)...)
	problems = append(problems, validateEvents(cfg.Events)...)

	if len(problems) == 0 {
		return nil
	}

	return fmt.Errorf("invalid config: %s", strings.Join(problems, "; "))
}

func validateHTTP(cfg HTTPConfig) []string {
	var problems []string

	if cfg.Port < 1 || cfg.Port > 65535 {
		problems = append(problems, "GRINCH_PORT must be between 1 and 65535")
	}
	if cfg.BaseURL != "" && !isValidBaseURL(cfg.BaseURL) {
		problems = append(problems, "GRINCH_BASE_URL must be a valid http or https URL")
	}

	return problems
}

func validateLogging(cfg LoggingConfig) []string {
	switch cfg.Level {
	case "debug", "info", "warn", "error":
		return nil
	default:
		return []string{"LOG_LEVEL must be one of: debug, info, warn, error"}
	}
}

func validateDatabase(cfg DatabaseConfig) []string {
	var problems []string

	missing := missingEnvVars(
		envValue("DATABASE_HOST", cfg.Host),
		envValue("DATABASE_USER", cfg.User),
		envValue("DATABASE_PASSWORD", cfg.Password),
		envValue("DATABASE_NAME", cfg.Name),
		envValue("DATABASE_SSLMODE", cfg.SSLMode),
	)
	if len(missing) > 0 {
		problems = append(problems, "missing required env vars: "+strings.Join(missing, ", "))
	}
	if cfg.Port <= 0 {
		problems = append(problems, "DATABASE_PORT must be greater than 0")
	}

	return problems
}

func validateAuth(httpCfg HTTPConfig, authCfg AuthConfig) []string {
	var problems []string

	hasLocalAdmin := authCfg.LocalAdminPass != ""
	hasCompleteEntra := hasCompleteEntraCredentials(authCfg)

	if !hasLocalAdmin && !hasCompleteEntra {
		problems = append(
			problems,
			"set one auth provider: LOCAL_ADMIN_PASSWORD or ENTRA_TENANT_ID, ENTRA_CLIENT_ID, ENTRA_CLIENT_SECRET",
		)
	}

	if hasAnyEntraCredentials(authCfg) {
		missing := missingEntraEnvVars(authCfg)
		if len(missing) > 0 {
			problems = append(
				problems,
				"missing required env vars for Entra auth: "+strings.Join(missing, ", "),
			)
		}
	}

	if hasLocalAdmin || hasCompleteEntra {
		if authCfg.JWTSecret == "" {
			problems = append(problems, "missing required env vars: JWT_SECRET")
		}
		if httpCfg.BaseURL == "" {
			problems = append(problems, "missing required env vars: GRINCH_BASE_URL")
		}
	}

	return problems
}

func validateEntraSync(authCfg AuthConfig, syncCfg EntraSyncConfig) []string {
	if !syncCfg.Enabled {
		return nil
	}

	var problems []string

	missing := missingEntraEnvVars(authCfg)
	if len(missing) > 0 {
		problems = append(
			problems,
			"missing required env vars for ENTRA_SYNC_ENABLED=true: "+strings.Join(missing, ", "),
		)
	}
	if syncCfg.Interval <= 0 {
		problems = append(problems, "ENTRA_SYNC_INTERVAL must be greater than 0")
	}

	return problems
}

func validateEvents(cfg EventsConfig) []string {
	if cfg.RetentionDays <= 0 {
		return []string{"EVENT_RETENTION_DAYS must be greater than 0"}
	}

	return nil
}

func envValue(name, value string) envVar {
	return envVar{name: name, value: strings.TrimSpace(value)}
}

func missingEnvVars(vars ...envVar) []string {
	missing := make([]string, 0, len(vars))

	for _, v := range vars {
		if v.value == "" {
			missing = append(missing, v.name)
		}
	}

	return missing
}

func hasAnyEntraCredentials(cfg AuthConfig) bool {
	return cfg.EntraTenantID != "" || cfg.EntraClientID != "" || cfg.EntraClientSecret != ""
}

func hasCompleteEntraCredentials(cfg AuthConfig) bool {
	return cfg.EntraTenantID != "" && cfg.EntraClientID != "" && cfg.EntraClientSecret != ""
}

func missingEntraEnvVars(cfg AuthConfig) []string {
	return missingEnvVars(
		envValue("ENTRA_TENANT_ID", cfg.EntraTenantID),
		envValue("ENTRA_CLIENT_ID", cfg.EntraClientID),
		envValue("ENTRA_CLIENT_SECRET", cfg.EntraClientSecret),
	)
}

func isValidBaseURL(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return false
	}

	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func parseDecisionAllowlist(raw string) ([]domain.ExecutionDecision, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	fields := strings.Fields(strings.NewReplacer(",", " ", ";", " ").Replace(raw))
	if len(fields) == 0 {
		return nil, nil
	}

	decisions := make([]domain.ExecutionDecision, 0, len(fields))
	for _, field := range fields {
		decision, err := domain.ParseExecutionDecision(field)
		if err != nil {
			return nil, fmt.Errorf("parse EVENT_DECISION_ALLOWLIST: %w", err)
		}
		decisions = append(decisions, decision)
	}

	return decisions, nil
}
