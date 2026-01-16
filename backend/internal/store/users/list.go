package users

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	coreusers "github.com/woodleighschool/grinch/internal/core/users"
	"github.com/woodleighschool/grinch/internal/listing"
	dblisting "github.com/woodleighschool/grinch/internal/store/listing"
)

func listUsers(ctx context.Context, pool *pgxpool.Pool, query listing.Query) ([]coreusers.User, int64, error) {
	cfg := dblisting.Config{
		Table:      "users",
		SelectCols: []string{"id", "upn", "display_name"},
		Columns: map[string]string{
			"id":           "id",
			"upn":          "upn",
			"display_name": "display_name",
		},
		SearchColumns: []string{"upn", "display_name"},
		DefaultSort:   listing.Sort{Field: "display_name", Desc: false},
	}
	return dblisting.List(ctx, pool, cfg, query, scanUserListItem)
}

func scanUserListItem(rows pgx.Rows) (coreusers.User, error) {
	var u coreusers.User
	err := rows.Scan(&u.ID, &u.UPN, &u.DisplayName)
	return u, err
}
