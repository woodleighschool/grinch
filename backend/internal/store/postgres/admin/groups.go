package admin

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/store/db"
	pgutil "github.com/woodleighschool/grinch/internal/store/postgres/shared"
)

func (store *Store) ListGroups(
	ctx context.Context,
	options domain.GroupListOptions,
) ([]domain.Group, int32, error) {
	orderBy, err := pgutil.OrderBy(options.Sort, options.Order, map[string]string{
		"id":           "g.id",
		"name":         "g.name",
		"created_at":   "g.created_at",
		"updated_at":   "g.updated_at",
		"description":  "g.description",
		"source":       "g.source",
		"member_count": "member_count",
	}, []string{"g.name ASC", "g.id ASC"})
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
SELECT
  g.id,
  g.name,
  g.description,
  g.source,
  COALESCE(member_counts.member_count, 0)::INT4 AS member_count,
  g.created_at,
  g.updated_at,
  COUNT(*) OVER()::INT4 AS total
FROM groups AS g
LEFT JOIN (
  SELECT group_id, COUNT(*)::INT4 AS member_count
  FROM group_memberships
  GROUP BY group_id
) AS member_counts
  ON member_counts.group_id = g.id
WHERE ($1 = '' OR g.name ILIKE $1 OR g.description ILIKE $1 OR g.source ILIKE $1)
ORDER BY %s
LIMIT NULLIF($2::INT, 0)
OFFSET $3
`, orderBy)

	rows, err := store.store.Pool().
		Query(ctx, query, pgutil.SearchPattern(options.Search), options.Limit, options.Offset)
	if err != nil {
		return nil, 0, err
	}

	return pgutil.CollectRows(rows, func(rows pgx.Rows) (domain.Group, int32, error) {
		var (
			row         db.Group
			memberCount int32
			total       int32
		)

		if scanErr := rows.Scan(
			&row.ID,
			&row.Name,
			&row.Description,
			&row.Source,
			&memberCount,
			&row.CreatedAt,
			&row.UpdatedAt,
			&total,
		); scanErr != nil {
			return domain.Group{}, 0, scanErr
		}

		mapped, mapErr := mapGroup(row, memberCount)
		if mapErr != nil {
			return domain.Group{}, 0, mapErr
		}

		return mapped, total, nil
	})
}

func (store *Store) GetGroup(ctx context.Context, id uuid.UUID) (domain.Group, error) {
	return pgutil.GetGroup(ctx, store.queries, id)
}

func (store *Store) CreateLocalGroup(ctx context.Context, name string, description string) (domain.Group, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return domain.Group{}, fmt.Errorf("create group id: %w", err)
	}

	row, err := store.queries.UpsertGroup(ctx, db.UpsertGroupParams{
		ID:          id,
		Name:        name,
		Description: description,
		Source:      string(domain.PrincipalSourceLocal),
	})
	if err != nil {
		return domain.Group{}, err
	}

	return store.GetGroup(ctx, row.ID)
}

func (store *Store) UpdateGroup(
	ctx context.Context,
	id uuid.UUID,
	name string,
	description string,
) (domain.Group, error) {
	row, err := store.queries.UpdateGroup(ctx, db.UpdateGroupParams{
		ID:          id,
		Name:        name,
		Description: description,
	})
	if err != nil {
		return domain.Group{}, err
	}
	if row.Status == groupMutationStatusNotFound {
		return domain.Group{}, pgx.ErrNoRows
	}
	if row.Status == groupMutationStatusReadOnly {
		return domain.Group{}, domain.ErrGroupReadOnly
	}
	if row.ID == nil {
		return domain.Group{}, errors.New("update group returned incomplete row")
	}

	return store.GetGroup(ctx, *row.ID)
}

func (store *Store) DeleteGroup(ctx context.Context, id uuid.UUID) error {
	status, err := store.queries.DeleteGroup(ctx, id)
	if err != nil {
		return err
	}

	switch status {
	case groupMutationStatusDeleted:
		return nil
	case groupMutationStatusReadOnly:
		return domain.ErrGroupReadOnly
	default:
		return pgx.ErrNoRows
	}
}

func mapGroup(row db.Group, memberCount int32) (domain.Group, error) {
	source, err := pgutil.ToSource(row.Source)
	if err != nil {
		return domain.Group{}, err
	}

	return domain.Group{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		Source:      source,
		MemberCount: memberCount,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}
