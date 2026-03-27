package config_test

import (
	"strings"
	"testing"

	"github.com/woodleighschool/grinch/internal/config"
)

func setBaseEnv(t *testing.T) {
	t.Helper()

	t.Setenv("GRINCH_PORT", "18080")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("DATABASE_HOST", "db")
	t.Setenv("DATABASE_PORT", "5432")
	t.Setenv("DATABASE_USER", "postgres")
	t.Setenv("DATABASE_PASSWORD", "postgres")
	t.Setenv("DATABASE_NAME", "grinch")
	t.Setenv("DATABASE_SSLMODE", "disable")
}

func TestLoadFromEnv_RequiresEntraCredentialsWhenSyncEnabled(t *testing.T) {
	setBaseEnv(t)

	t.Setenv("GRINCH_BASE_URL", "https://grinch.example.com")
	t.Setenv("LOCAL_ADMIN_PASSWORD", "admin")
	t.Setenv("JWT_SECRET", "jwt-secret")
	t.Setenv("ENTRA_SYNC_ENABLED", "true")
	t.Setenv("ENTRA_TENANT_ID", "")
	t.Setenv("ENTRA_CLIENT_ID", "")
	t.Setenv("ENTRA_CLIENT_SECRET", "")

	_, err := config.LoadFromEnv()
	if err == nil {
		t.Fatalf("LoadFromEnv() expected error, got nil")
	}

	if !strings.Contains(
		err.Error(),
		"missing required env vars for ENTRA_SYNC_ENABLED=true: ENTRA_TENANT_ID, ENTRA_CLIENT_ID, ENTRA_CLIENT_SECRET",
	) {
		t.Fatalf("error = %v, want missing Entra sync env vars", err)
	}
}

func TestLoadFromEnv_RequiresAuthProvider(t *testing.T) {
	setBaseEnv(t)

	t.Setenv("GRINCH_BASE_URL", "https://grinch.example.com")
	t.Setenv("ENTRA_SYNC_ENABLED", "false")

	_, err := config.LoadFromEnv()
	if err == nil {
		t.Fatalf("LoadFromEnv() expected error, got nil")
	}

	if !strings.Contains(
		err.Error(),
		"set one auth provider: LOCAL_ADMIN_PASSWORD or ENTRA_TENANT_ID, ENTRA_CLIENT_ID, ENTRA_CLIENT_SECRET",
	) {
		t.Fatalf("error = %v, want auth provider error", err)
	}
}

func TestLoadFromEnv_RequiresJWTSecretWhenAuthEnabled(t *testing.T) {
	setBaseEnv(t)

	t.Setenv("GRINCH_BASE_URL", "https://grinch.example.com")
	t.Setenv("ENTRA_SYNC_ENABLED", "false")
	t.Setenv("LOCAL_ADMIN_PASSWORD", "admin")

	_, err := config.LoadFromEnv()
	if err == nil {
		t.Fatalf("LoadFromEnv() expected error, got nil")
	}

	if !strings.Contains(err.Error(), "missing required env vars: JWT_SECRET") {
		t.Fatalf("error = %v, want missing JWT_SECRET", err)
	}
}

func TestLoadFromEnv_RequiresBaseURLWhenAuthEnabled(t *testing.T) {
	setBaseEnv(t)

	t.Setenv("ENTRA_SYNC_ENABLED", "false")
	t.Setenv("LOCAL_ADMIN_PASSWORD", "admin")
	t.Setenv("JWT_SECRET", "jwt-secret")

	_, err := config.LoadFromEnv()
	if err == nil {
		t.Fatalf("LoadFromEnv() expected error, got nil")
	}

	if !strings.Contains(err.Error(), "missing required env vars: GRINCH_BASE_URL") {
		t.Fatalf("error = %v, want missing GRINCH_BASE_URL", err)
	}
}
