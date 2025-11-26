package store

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/woodleighschool/grinch/internal/store/sqlc"
)

//go:embed migrate/*.sql
var migrations embed.FS

// ErrNilPool indicates the Store was used after being closed.
var ErrNilPool = errors.New("store: nil pool")

// Options tweak the Postgres pool behaviour.
type Options struct {
	URL             string
	MaxConnections  int32
	MinConnections  int32
	MaxConnLifetime time.Duration
}

// Store wraps sqlc queries and connection pool helpers.
type Store struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// Open initialises the pgx connection pool and sqlc helpers.
func Open(ctx context.Context, opts Options) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(opts.URL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	if opts.MaxConnections > 0 {
		cfg.MaxConns = opts.MaxConnections
	}
	if opts.MinConnections > 0 {
		cfg.MinConns = opts.MinConnections
	}
	if opts.MaxConnLifetime > 0 {
		cfg.MaxConnLifetime = opts.MaxConnLifetime
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	return &Store{pool: pool, queries: sqlc.New(pool)}, nil
}

// Close shuts down the underlying pool.
func (s *Store) Close() {
	if s == nil || s.pool == nil {
		return
	}
	s.pool.Close()
}

// Ping verifies connectivity to Postgres.
func (s *Store) Ping(ctx context.Context) error {
	if s.pool == nil {
		return ErrNilPool
	}
	return s.pool.Ping(ctx)
}

// WithTx runs fn inside a transaction and commits on success.
func (s *Store) WithTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if tx != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				_ = rollbackErr
			}
		}
	}()
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	tx = nil
	return nil
}

// Queries exposes the raw sqlc handle for callers that need fine control.
func (s *Store) Queries() *sqlc.Queries {
	return s.queries
}

// Migrate applies embedded SQL migration files in alphabetical order.
func (s *Store) Migrate(ctx context.Context) error {
	entries, err := migrations.ReadDir("migrate")
	if err != nil {
		return fmt.Errorf("list migrations: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		stmt, err := migrations.ReadFile("migrate/" + entry.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}
		if _, err := s.pool.Exec(ctx, string(stmt)); err != nil {
			return fmt.Errorf("apply migration %s: %w", entry.Name(), err)
		}
	}
	return nil
}
