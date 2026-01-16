// Package memberships provides persistence for group memberships.
package memberships

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	coreerrors "github.com/woodleighschool/grinch/internal/core/errors"
	corememberships "github.com/woodleighschool/grinch/internal/core/memberships"
	"github.com/woodleighschool/grinch/internal/store/db/pgconv"
	"github.com/woodleighschool/grinch/internal/store/db/sqlc"
)

// Repo provides persistence operations for memberships.
type Repo struct {
	q *sqlc.Queries
}

// New constructs a Repo backed by PostgreSQL.
func New(pool *pgxpool.Pool) *Repo {
	return &Repo{q: sqlc.New(pool)}
}

// ListByUser lists memberships for a user and returns the total count.
func (r *Repo) ListByUser(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
) ([]corememberships.Membership, int64, error) {
	limitPg, offsetPg := pgconv.LimitOffset(limit, offset)
	rows, err := r.q.ListMembershipsByUser(ctx, sqlc.ListMembershipsByUserParams{
		UserID: userID,
		Limit:  limitPg,
		Offset: offsetPg,
	})
	if err != nil {
		return nil, 0, coreerrors.FromStore(err, nil)
	}

	total, err := r.q.CountMembershipsByUser(ctx, userID)
	if err != nil {
		return nil, 0, coreerrors.FromStore(err, nil)
	}

	return mapMemberships(rows), total, nil
}

// ListByGroup lists memberships for a group and returns the total count.
func (r *Repo) ListByGroup(
	ctx context.Context,
	groupID uuid.UUID,
	limit, offset int,
) ([]corememberships.Membership, int64, error) {
	limitPg, offsetPg := pgconv.LimitOffset(limit, offset)
	rows, err := r.q.ListMembershipsByGroup(ctx, sqlc.ListMembershipsByGroupParams{
		GroupID: groupID,
		Limit:   limitPg,
		Offset:  offsetPg,
	})
	if err != nil {
		return nil, 0, coreerrors.FromStore(err, nil)
	}

	total, err := r.q.CountMembershipsByGroup(ctx, groupID)
	if err != nil {
		return nil, 0, coreerrors.FromStore(err, nil)
	}

	return mapMemberships(rows), total, nil
}

// GroupIDsForUser returns the group IDs that the user is a member of.
func (r *Repo) GroupIDsForUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	ids, err := r.q.ListGroupIDsByUserID(ctx, userID)
	if err != nil {
		return nil, coreerrors.FromStore(err, nil)
	}
	return ids, nil
}

func mapMemberships(rows []sqlc.GroupMembership) []corememberships.Membership {
	out := make([]corememberships.Membership, len(rows))
	for i, row := range rows {
		out[i] = corememberships.Membership{
			ID:      row.ID,
			GroupID: row.GroupID,
			UserID:  row.UserID,
		}
	}
	return out
}
