package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/store/db"
)

var (
	userListSortColumns = map[string]string{ //nolint:gochecknoglobals // package-level lookup table, not mutable state
		"id":           "u.id",
		"display_name": "u.display_name",
		"upn":          "u.upn",
		"source":       "u.source",
		"created_at":   "u.created_at",
		"updated_at":   "u.updated_at",
	}

	userListDefaultOrder = []string{ //nolint:gochecknoglobals // package-level lookup table, not mutable state
		"u.display_name ASC",
		"u.id ASC",
	}
)

func (s *Store) ListUsers( //nolint:dupl // structurally similar to other List* functions by design
	ctx context.Context,
	opts domain.ListOptions,
) ([]domain.User, int32, error) {
	orderBy, err := orderBy(opts.Sort, opts.Order, userListSortColumns, userListDefaultOrder)
	if err != nil {
		return nil, 0, err
	}

	where := []string{
		"($1 = '' OR u.display_name ILIKE $1 OR u.upn ILIKE $1 OR u.source::text ILIKE $1)",
	}
	args := []any{searchPattern(opts.Search)}

	if len(opts.IDs) > 0 {
		where = append(where, fmt.Sprintf("u.id = ANY($%d)", len(args)+1))
		args = append(args, opts.IDs)
	}

	limitArg := len(args) + 1
	offsetArg := limitArg + 1

	query := fmt.Sprintf(`
SELECT
  u.id,
  u.upn,
  u.display_name,
  u.source,
  u.created_at,
  u.updated_at,
  COUNT(*) OVER()::INT4 AS total
FROM users AS u
WHERE %s
ORDER BY %s
LIMIT NULLIF($%d::INT, 0)
OFFSET $%d
`, strings.Join(where, " AND "), orderBy, limitArg, offsetArg)

	args = append(args, opts.Limit, opts.Offset)

	rows, err := s.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}

	return collectRows(rows, scanUserRow)
}

func (s *Store) GetUser(ctx context.Context, id uuid.UUID) (domain.User, error) {
	row, err := s.Queries().GetUser(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	user, err := mapUser(row)
	if err != nil {
		return domain.User{}, err
	}

	memberships, _, err := s.ListMemberships(ctx, domain.MembershipListOptions{
		ListOptions: domain.ListOptions{},
		UserID:      &id,
	})
	if err != nil {
		return domain.User{}, err
	}

	user.GroupIDs = make([]uuid.UUID, 0, len(memberships))
	for _, membership := range memberships {
		user.GroupIDs = append(user.GroupIDs, membership.Group.ID)
	}

	return user, nil
}

func scanUserRow(rows pgx.Rows) (domain.User, int32, error) {
	var (
		row   db.User
		total int32
	)

	if err := rows.Scan(
		&row.ID,
		&row.Upn,
		&row.DisplayName,
		&row.Source,
		&row.CreatedAt,
		&row.UpdatedAt,
		&total,
	); err != nil {
		return domain.User{}, 0, err
	}

	user, err := mapUser(row)
	if err != nil {
		return domain.User{}, 0, err
	}

	return user, total, nil
}

func mapUser(row db.User) (domain.User, error) {
	source, err := domain.ParsePrincipalSource(string(row.Source))
	if err != nil {
		return domain.User{}, fmt.Errorf("parse user source: %w", err)
	}

	return domain.User{
		ID:          row.ID,
		UPN:         row.Upn,
		DisplayName: row.DisplayName,
		Source:      source,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}
