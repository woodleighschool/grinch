// Package db provides PostgreSQL connection configuration and pool initialisation.
package db

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config defines the inputs required to establish a PostgreSQL connection pool.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string

	MaxConns    int32
	MinConns    int32
	MaxLifetime time.Duration
}

// DSN returns a PostgreSQL connection string derived from the configuration.
func (c Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=%s",
		c.User,
		c.Password,
		net.JoinHostPort(c.Host, strconv.Itoa(c.Port)),
		c.Database,
		c.SSLMode,
	)
}

// Connect creates and validates a PostgreSQL connection pool.
// TODO: check for readiness before trying to connect.
func Connect(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	if cfg.MaxConns == 0 {
		cfg.MaxConns = 25
	}
	if cfg.MinConns == 0 {
		cfg.MinConns = 5
	}
	if cfg.MaxLifetime == 0 {
		cfg.MaxLifetime = time.Hour
	}

	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnLifetime = cfg.MaxLifetime

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	return pool, nil
}
