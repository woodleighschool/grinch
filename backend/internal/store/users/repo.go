// Package users provides user persistence backed by PostgreSQL.
package users

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	coreerrors "github.com/woodleighschool/grinch/internal/core/errors"
	coreusers "github.com/woodleighschool/grinch/internal/core/users"
	"github.com/woodleighschool/grinch/internal/listing"
	"github.com/woodleighschool/grinch/internal/store/db/sqlc"
)

// Repo provides persistence operations for users.
type Repo struct {
	q    *sqlc.Queries
	pool *pgxpool.Pool
}

// New constructs a user repository.
func New(pool *pgxpool.Pool) *Repo {
	return &Repo{q: sqlc.New(pool), pool: pool}
}

// Get returns the user with the given ID.
func (r *Repo) Get(ctx context.Context, id uuid.UUID) (coreusers.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return coreusers.User{}, coreerrors.FromStore(err, nil)
	}
	return toDomainUser(row), nil
}

// GetByUPN returns the user ID for the given UPN.
func (r *Repo) GetByUPN(ctx context.Context, upn string) (uuid.UUID, error) {
	id, err := r.q.GetUserIDByUPN(ctx, upn)
	if err != nil {
		if coreerrors.IsCode(coreerrors.FromStore(err, nil), coreerrors.CodeNotFound) {
			return uuid.Nil, nil
		}
		return uuid.Nil, coreerrors.FromStore(err, nil)
	}
	return id, nil
}

// Upsert inserts or updates a user by ID.
func (r *Repo) Upsert(ctx context.Context, user coreusers.User) error {
	err := r.q.UpsertUser(ctx, sqlc.UpsertUserParams{
		ID:          user.ID,
		Upn:         user.UPN,
		DisplayName: user.DisplayName,
	})
	return coreerrors.FromStore(err, nil)
}

// List returns users matching the listing query.
func (r *Repo) List(ctx context.Context, query listing.Query) ([]coreusers.User, listing.Page, error) {
	items, total, err := listUsers(ctx, r.pool, query)
	if err != nil {
		return nil, listing.Page{}, coreerrors.FromStore(err, nil)
	}
	return items, listing.Page{Total: total}, nil
}

func toDomainUser(row sqlc.User) coreusers.User {
	return coreusers.User{
		ID:          row.ID,
		UPN:         row.Upn,
		DisplayName: row.DisplayName,
	}
}
