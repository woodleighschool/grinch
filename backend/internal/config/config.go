package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds runtime configuration for the server.
type Config struct {
	ServerAddress            string
	FrontendDistDir          string
	DatabaseURL              string
	DatabaseMaxConns         *int32
	DatabaseMinConns         *int32
	DatabaseMaxConnLifetime  *time.Duration
	DatabaseMaxConnIdleTime  *time.Duration
	DatabaseHealthCheckEvery *time.Duration
	CookieName               string
	CookieSecret             string
	AzureTenantID            string
	AzureClientID            string
	AzureClientSecret        string
	SyncInterval             time.Duration
	SSEBufferSize            int
	AllowedOrigins           []string
	EnableMetrics            bool
	LogLevel                 string
	SAMLMetadataURL          string
	SAMLEntityID             string
	SAMLACSURL               string
	SAMLNameIDFormat         string
	SAMLObjectIDAttribute    string
	SAMLUPNAttribute         string
	SAMLEmailAttribute       string
	SAMLDisplayNameAttribute string
	// Initial admin user configuration
	InitialAdminEmail       string
	InitialAdminPrincipal   string
	InitialAdminDisplayName string
	InitialAdminPassword    string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{
		ServerAddress:            getEnv("SERVER_ADDRESS", ":8080"),
		FrontendDistDir:          getEnv("FRONTEND_DIST", "/frontend"),
		DatabaseURL:              os.Getenv("DATABASE_URL"),
		CookieName:               getEnv("SESSION_COOKIE_NAME", "grinch_session"),
		CookieSecret:             os.Getenv("SESSION_COOKIE_SECRET"),
		AzureTenantID:            os.Getenv("AZURE_TENANT_ID"),
		AzureClientID:            os.Getenv("AZURE_CLIENT_ID"),
		AzureClientSecret:        os.Getenv("AZURE_CLIENT_SECRET"),
		SSEBufferSize:            getEnvInt("EVENT_SSE_BUFFER", 64),
		EnableMetrics:            getEnvBool("ENABLE_METRICS", true),
		LogLevel:                 getEnv("LOG_LEVEL", "info"),
		SAMLMetadataURL:          os.Getenv("SAML_METADATA_URL"),
		SAMLEntityID:             os.Getenv("SAML_SP_ENTITY_ID"),
		SAMLACSURL:               os.Getenv("SAML_ACS_URL"),
		SAMLNameIDFormat:         getEnv("SAML_NAME_ID_FORMAT", ""),
		SAMLObjectIDAttribute:    getEnv("SAML_OBJECT_ID_ATTRIBUTE", "http://schemas.microsoft.com/identity/claims/objectidentifier"),
		SAMLUPNAttribute:         getEnv("SAML_UPN_ATTRIBUTE", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/upn"),
		SAMLEmailAttribute:       getEnv("SAML_EMAIL_ATTRIBUTE", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"),
		SAMLDisplayNameAttribute: getEnv("SAML_DISPLAY_NAME_ATTRIBUTE", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"),
		InitialAdminEmail:        os.Getenv("INITIAL_ADMIN_EMAIL"),
		InitialAdminPrincipal:    os.Getenv("INITIAL_ADMIN_PRINCIPAL_NAME"),
		InitialAdminDisplayName:  os.Getenv("INITIAL_ADMIN_DISPLAY_NAME"),
		InitialAdminPassword:     os.Getenv("INITIAL_ADMIN_PASSWORD"),
	}

	if raw := os.Getenv("SYNC_INTERVAL"); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid SYNC_INTERVAL: %w", err)
		}
		cfg.SyncInterval = d
	} else {
		cfg.SyncInterval = 15 * time.Minute
	}

	if origins := os.Getenv("ALLOWED_ORIGINS"); origins != "" {
		cfg.AllowedOrigins = splitAndTrim(origins)
	}

	var err error
	if cfg.DatabaseMaxConns, err = getEnvInt32Ptr("DB_MAX_CONNS"); err != nil {
		return nil, err
	}
	if cfg.DatabaseMinConns, err = getEnvInt32Ptr("DB_MIN_CONNS"); err != nil {
		return nil, err
	}
	if cfg.DatabaseMaxConnLifetime, err = getEnvDurationPtr("DB_MAX_CONN_LIFETIME"); err != nil {
		return nil, err
	}
	if cfg.DatabaseMaxConnIdleTime, err = getEnvDurationPtr("DB_MAX_CONN_IDLE_TIME"); err != nil {
		return nil, err
	}
	if cfg.DatabaseHealthCheckEvery, err = getEnvDurationPtr("DB_HEALTH_CHECK_PERIOD"); err != nil {
		return nil, err
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL not set")
	}
	if cfg.CookieSecret == "" {
		return nil, fmt.Errorf("SESSION_COOKIE_SECRET not set")
	}
	if cfg.AzureTenantID == "" || cfg.AzureClientID == "" || cfg.AzureClientSecret == "" {
		return nil, fmt.Errorf("azure entra variables missing")
	}

	return cfg, nil
}

// GetLogLevel parses the configured log level and falls back to info if invalid.
func (c *Config) GetLogLevel() slog.Level {
	switch strings.ToLower(c.LogLevel) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "info":
		return slog.LevelInfo
	default:
		return slog.LevelInfo
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		switch v {
		case "1", "true", "TRUE", "True", "yes", "YES":
			return true
		case "0", "false", "FALSE", "False", "no", "NO":
			return false
		}
	}
	return fallback
}

func splitAndTrim(value string) []string {
	raw := strings.Split(value, ",")
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		trimmed := strings.TrimSpace(v)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func getEnvInt32Ptr(key string) (*int32, error) {
	if v := os.Getenv(key); v != "" {
		parsed, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid %s: %w", key, err)
		}
		if parsed <= 0 {
			return nil, fmt.Errorf("%s must be positive", key)
		}
		val := int32(parsed)
		return &val, nil
	}
	return nil, nil
}

func getEnvDurationPtr(key string) (*time.Duration, error) {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("invalid %s: %w", key, err)
		}
		if d <= 0 {
			return nil, fmt.Errorf("%s must be positive", key)
		}
		return &d, nil
	}
	return nil, nil
}
