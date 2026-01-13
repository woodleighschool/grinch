// Package groups provides persistence operations for groups and memberships.
package groups

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/woodleighschool/grinch/internal/domain/errx"
	"github.com/woodleighschool/grinch/internal/domain/groups"
	"github.com/woodleighschool/grinch/internal/listing"
	"github.com/woodleighschool/grinch/internal/store/db/sqlc"
)

// Repo provides persistence operations for groups.
type Repo struct {
	q    *sqlc.Queries
	pool *pgxpool.Pool
}

// New constructs a Repo backed by the provided database pool.
func New(pool *pgxpool.Pool) *Repo {
	return &Repo{q: sqlc.New(pool), pool: pool}
}

// Get returns the group with the given ID.
func (r *Repo) Get(ctx context.Context, id uuid.UUID) (groups.Group, error) {
	row, err := r.q.GetGroupByID(ctx, id)
	if err != nil {
		return groups.Group{}, errx.FromStore(err, nil)
	}
	return toDomainGroup(row), nil
}

// Upsert inserts or updates the given group.
func (r *Repo) Upsert(ctx context.Context, g groups.Group) error {
	err := r.q.UpsertGroupByID(ctx, sqlc.UpsertGroupByIDParams{
		ID:          g.ID,
		DisplayName: g.DisplayName,
		Description: g.Description,
		MemberCount: g.MemberCount,
	})
	return errx.FromStore(err, nil)
}

// ReplaceMemberships replaces all memberships for the given group in a single transaction.
func (r *Repo) ReplaceMemberships(ctx context.Context, groupID uuid.UUID, userIDs []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return errx.FromStore(err, nil)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := r.q.WithTx(tx)

	if err = qtx.DeleteMembershipsByGroupID(ctx, groupID); err != nil {
		return errx.FromStore(err, nil)
	}

	for _, userID := range userIDs {
		if err = qtx.CreateGroupMembership(ctx, sqlc.CreateGroupMembershipParams{
			GroupID: groupID,
			UserID:  userID,
		}); err != nil {
			return errx.FromStore(err, nil)
		}
	}

	return errx.FromStore(tx.Commit(ctx), nil)
}

// List returns groups matching the listing query.
func (r *Repo) List(ctx context.Context, query listing.Query) ([]groups.Group, listing.Page, error) {
	items, total, err := listGroups(ctx, r.pool, query)
	if err != nil {
		return nil, listing.Page{}, errx.FromStore(err, nil)
	}
	return items, listing.Page{Total: total}, nil
}

func toDomainGroup(row sqlc.Group) groups.Group {
	return groups.Group{
		ID:          row.ID,
		DisplayName: row.DisplayName,
		Description: row.Description,
		MemberCount: row.MemberCount,
	}
}
