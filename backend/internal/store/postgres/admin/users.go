package admin

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/store/db"
	pgutil "github.com/woodleighschool/grinch/internal/store/postgres/shared"
)

func (store *Store) ListUsers(
	ctx context.Context,
	options domain.UserListOptions,
) ([]domain.User, int32, error) {
	orderBy, err := pgutil.OrderBy(options.Sort, options.Order, map[string]string{
		"id":           "u.id",
		"display_name": "u.display_name",
		"upn":          "u.upn",
		"source":       "u.source",
		"created_at":   "u.created_at",
		"updated_at":   "u.updated_at",
	}, []string{"u.display_name ASC", "u.id ASC"})
	if err != nil {
		return nil, 0, err
	}

	whereClauses := []string{
		"($1 = '' OR u.display_name ILIKE $1 OR u.upn ILIKE $1 OR u.source ILIKE $1)",
	}
	args := []any{pgutil.SearchPattern(options.Search)}
	if len(options.IDs) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("u.id = ANY($%d)", len(args)+1))
		args = append(args, options.IDs)
	}
	limitParam := len(args) + 1
	offsetParam := limitParam + 1

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
`, strings.Join(whereClauses, " AND "), orderBy, limitParam, offsetParam)

	args = append(args, options.Limit, options.Offset)

	rows, err := store.store.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return pgutil.CollectRows(rows, func(rows pgx.Rows) (domain.User, int32, error) {
		var (
			row   db.User
			total int32
		)

		if scanErr := rows.Scan(
			&row.ID,
			&row.Upn,
			&row.DisplayName,
			&row.Source,
			&row.CreatedAt,
			&row.UpdatedAt,
			&total,
		); scanErr != nil {
			return domain.User{}, 0, scanErr
		}

		mapped, mapErr := mapUser(row)
		if mapErr != nil {
			return domain.User{}, 0, mapErr
		}

		return mapped, total, nil
	})
}

func (store *Store) GetUser(ctx context.Context, id uuid.UUID) (domain.User, error) {
	return pgutil.GetUser(ctx, store.store.Queries(), id)
}

func mapUser(row db.User) (domain.User, error) {
	source, err := domain.ParsePrincipalSource(row.Source)
	if err != nil {
		return domain.User{}, err
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
