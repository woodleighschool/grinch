package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/store/db"
)

const (
	groupMutationStatusDeleted  = "deleted"
	groupMutationStatusNotFound = "not_found"
	groupMutationStatusReadOnly = "read_only"
)

var (
	errIncompleteGroupMutationRow = errors.New("group mutation returned incomplete row")

	groupListSortColumns = map[string]string{ //nolint:gochecknoglobals // package-level lookup table, not mutable state
		"id":           "g.id",
		"name":         "g.name",
		"description":  "g.description",
		"source":       "g.source",
		"member_count": "member_count",
		"created_at":   "g.created_at",
		"updated_at":   "g.updated_at",
	}

	groupListDefaultOrder = []string{ //nolint:gochecknoglobals // package-level lookup table, not mutable state
		"g.name ASC",
		"g.id ASC",
	}
)

func (s *Store) ListGroups( //nolint:dupl // structurally similar to other List* functions by design
	ctx context.Context,
	opts domain.ListOptions,
) ([]domain.Group, int32, error) {
	orderBy, err := orderBy(
		opts.Sort,
		opts.Order,
		groupListSortColumns,
		groupListDefaultOrder,
	)
	if err != nil {
		return nil, 0, err
	}

	where := []string{
		"($1 = '' OR g.name ILIKE $1 OR g.description ILIKE $1 OR g.source::text ILIKE $1)",
	}
	args := []any{searchPattern(opts.Search)}

	if len(opts.IDs) > 0 {
		where = append(where, fmt.Sprintf("g.id = ANY($%d)", len(args)+1))
		args = append(args, opts.IDs)
	}

	limitArg := len(args) + 1
	offsetArg := limitArg + 1

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
  SELECT
    group_id,
    COUNT(*)::INT4 AS member_count
  FROM group_memberships
  GROUP BY group_id
) AS member_counts
  ON member_counts.group_id = g.id
WHERE %s
ORDER BY %s
LIMIT NULLIF($%d::INT, 0)
OFFSET $%d
`, strings.Join(where, " AND "), orderBy, limitArg, offsetArg)

	args = append(args, opts.Limit, opts.Offset)

	rows, err := s.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list groups: %w", err)
	}

	return collectRows(rows, scanGroupRow)
}

func (s *Store) GetGroup(ctx context.Context, id uuid.UUID) (domain.Group, error) {
	row, err := s.Queries().GetGroup(ctx, id)
	if err != nil {
		return domain.Group{}, err
	}

	return mapGroupFields(
		row.ID,
		row.Name,
		row.Description,
		string(row.Source),
		row.MemberCount,
		row.CreatedAt,
		row.UpdatedAt,
	)
}

func (s *Store) CreateLocalGroup(
	ctx context.Context,
	name string,
	description string,
) (domain.Group, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return domain.Group{}, fmt.Errorf("create group id: %w", err)
	}

	row, err := s.Queries().UpsertGroup(ctx, db.UpsertGroupParams{
		ID:          id,
		Name:        name,
		Description: description,
		Source:      db.PrincipalSource(domain.PrincipalSourceLocal),
	})
	if err != nil {
		return domain.Group{}, err
	}

	return mapGroupFields(
		row.ID,
		row.Name,
		row.Description,
		string(row.Source),
		row.MemberCount,
		row.CreatedAt,
		row.UpdatedAt,
	)
}

func (s *Store) UpdateGroup(
	ctx context.Context,
	id uuid.UUID,
	name string,
	description string,
) (domain.Group, error) {
	row, err := s.Queries().UpdateGroup(ctx, db.UpdateGroupParams{
		ID:          id,
		Name:        name,
		Description: description,
	})
	if err != nil {
		return domain.Group{}, err
	}

	switch row.Status {
	case groupMutationStatusNotFound:
		return domain.Group{}, pgx.ErrNoRows
	case groupMutationStatusReadOnly:
		return domain.Group{}, domain.ErrGroupReadOnly
	}

	group, err := mapUpdatedGroupRow(row)
	if err != nil {
		return domain.Group{}, fmt.Errorf("map updated group: %w", err)
	}

	return group, nil
}

func (s *Store) DeleteGroup(ctx context.Context, id uuid.UUID) error {
	status, err := s.Queries().DeleteGroup(ctx, id)
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

func scanGroupRow(rows pgx.Rows) (domain.Group, int32, error) {
	var (
		row         db.Group
		memberCount int32
		total       int32
	)

	if err := rows.Scan(
		&row.ID,
		&row.Name,
		&row.Description,
		&row.Source,
		&memberCount,
		&row.CreatedAt,
		&row.UpdatedAt,
		&total,
	); err != nil {
		return domain.Group{}, 0, err
	}

	group, err := mapGroup(row, memberCount)
	if err != nil {
		return domain.Group{}, 0, err
	}

	return group, total, nil
}

func mapGroup(row db.Group, memberCount int32) (domain.Group, error) {
	return mapGroupFields(
		row.ID,
		row.Name,
		row.Description,
		string(row.Source),
		memberCount,
		row.CreatedAt,
		row.UpdatedAt,
	)
}

func mapUpdatedGroupRow(row db.UpdateGroupRow) (domain.Group, error) {
	if row.ID == nil || !row.Name.Valid || !row.Source.Valid || row.CreatedAt == nil || row.UpdatedAt == nil {
		return domain.Group{}, errIncompleteGroupMutationRow
	}

	return mapGroupFields(
		*row.ID,
		row.Name.String,
		row.Description.String,
		string(row.Source.PrincipalSource),
		row.MemberCount,
		*row.CreatedAt,
		*row.UpdatedAt,
	)
}

func mapGroupFields(
	id uuid.UUID,
	name string,
	description string,
	sourceText string,
	memberCount int32,
	createdAt time.Time,
	updatedAt time.Time,
) (domain.Group, error) {
	source, err := domain.ParsePrincipalSource(sourceText)
	if err != nil {
		return domain.Group{}, fmt.Errorf("parse group source: %w", err)
	}

	return domain.Group{
		ID:          id,
		Name:        name,
		Description: description,
		Source:      source,
		MemberCount: memberCount,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}
