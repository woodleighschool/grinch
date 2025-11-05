package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMaxConns          = int32(64)
	defaultMinConns          = int32(8)
	defaultMaxConnLifetime   = time.Hour
	defaultMaxConnIdleTime   = 15 * time.Minute
	defaultHealthCheckPeriod = time.Minute
	defaultConnJitter        = time.Minute
)

// PoolOption allows callers to customise the pgx connection pool configuration.
type PoolOption func(*pgxpool.Config)

// PoolOptions collates optional pool settings. Fields left nil will retain the defaults.
type PoolOptions struct {
	MaxConns          *int32
	MinConns          *int32
	MaxConnLifetime   *time.Duration
	MaxConnIdleTime   *time.Duration
	HealthCheckPeriod *time.Duration
}

// WithPoolOptions applies a group of connection pool options in one call.
func WithPoolOptions(opts PoolOptions) PoolOption {
	return func(cfg *pgxpool.Config) {
		if opts.MaxConns != nil && *opts.MaxConns > 0 {
			cfg.MaxConns = *opts.MaxConns
		}
		if opts.MinConns != nil && *opts.MinConns >= 0 {
			cfg.MinConns = *opts.MinConns
		}
		if opts.MaxConnLifetime != nil && *opts.MaxConnLifetime > 0 {
			cfg.MaxConnLifetime = *opts.MaxConnLifetime
		}
		if opts.MaxConnIdleTime != nil && *opts.MaxConnIdleTime > 0 {
			cfg.MaxConnIdleTime = *opts.MaxConnIdleTime
		}
		if opts.HealthCheckPeriod != nil && *opts.HealthCheckPeriod > 0 {
			cfg.HealthCheckPeriod = *opts.HealthCheckPeriod
		}
	}
}

// WithMaxConns overrides the maximum number of pooled connections.
func WithMaxConns(max int32) PoolOption {
	return func(cfg *pgxpool.Config) {
		if max > 0 {
			cfg.MaxConns = max
		}
	}
}

// WithMinConns overrides the minimum number of pooled connections to keep warm.
func WithMinConns(min int32) PoolOption {
	return func(cfg *pgxpool.Config) {
		if min >= 0 {
			cfg.MinConns = min
		}
	}
}

// WithConnLifetime overrides the maximum lifetime a connection may exist for.
func WithConnLifetime(d time.Duration) PoolOption {
	return func(cfg *pgxpool.Config) {
		if d > 0 {
			cfg.MaxConnLifetime = d
		}
	}
}

// WithConnIdleTime overrides the idle timeout for pooled connections.
func WithConnIdleTime(d time.Duration) PoolOption {
	return func(cfg *pgxpool.Config) {
		if d > 0 {
			cfg.MaxConnIdleTime = d
		}
	}
}

// WithHealthCheckPeriod overrides the pool health check cadence.
func WithHealthCheckPeriod(d time.Duration) PoolOption {
	return func(cfg *pgxpool.Config) {
		if d > 0 {
			cfg.HealthCheckPeriod = d
		}
	}
}

// Connect establishes a pgx connection pool using sensible defaults that can be customised.
func Connect(ctx context.Context, url string, opts ...PoolOption) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}

	applyPoolDefaults(cfg)
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}

	if cfg.MinConns > cfg.MaxConns {
		cfg.MinConns = cfg.MaxConns
	}
	if cfg.ConnConfig.RuntimeParams == nil {
		cfg.ConnConfig.RuntimeParams = map[string]string{}
	}
	if cfg.ConnConfig.RuntimeParams["application_name"] == "" {
		cfg.ConnConfig.RuntimeParams["application_name"] = "grinch-backend"
	}

	return pgxpool.NewWithConfig(ctx, cfg)
}

func applyPoolDefaults(cfg *pgxpool.Config) {
	cfg.MaxConns = defaultMaxConns
	cfg.MinConns = defaultMinConns
	cfg.MaxConnLifetime = defaultMaxConnLifetime
	cfg.MaxConnIdleTime = defaultMaxConnIdleTime
	cfg.HealthCheckPeriod = defaultHealthCheckPeriod
	cfg.MaxConnLifetimeJitter = defaultConnJitter
}
