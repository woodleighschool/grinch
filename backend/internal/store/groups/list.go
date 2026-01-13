package groups

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/woodleighschool/grinch/internal/domain/groups"
	"github.com/woodleighschool/grinch/internal/listing"
	dblisting "github.com/woodleighschool/grinch/internal/store/listing"
)

func listGroups(ctx context.Context, pool *pgxpool.Pool, query listing.Query) ([]groups.Group, int64, error) {
	cfg := dblisting.Config{
		Table:      "groups",
		SelectCols: []string{"id", "display_name", "description", "member_count"},
		Columns: map[string]string{
			"id":           "id",
			"display_name": "display_name",
			"description":  "description",
			"member_count": "member_count",
		},
		SearchColumns: []string{"display_name", "description"},
		DefaultSort:   listing.Sort{Field: "display_name", Desc: false},
	}
	return dblisting.List(ctx, pool, cfg, query, scanGroupListItem)
}

func scanGroupListItem(rows pgx.Rows) (groups.Group, error) {
	var g groups.Group
	err := rows.Scan(&g.ID, &g.DisplayName, &g.Description, &g.MemberCount)
	return g, err
}
