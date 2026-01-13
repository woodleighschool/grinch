package policies

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/woodleighschool/grinch/internal/domain/policies"
	"github.com/woodleighschool/grinch/internal/listing"
	dblisting "github.com/woodleighschool/grinch/internal/store/listing"
)

func listPolicies(ctx context.Context, pool *pgxpool.Pool, query listing.Query) ([]policies.ListItem, int64, error) {
	cfg := dblisting.Config{
		Table:      "policies",
		SelectCols: []string{"id", "name", "description", "enabled", "priority"},
		Columns: map[string]string{
			"id":          "id",
			"name":        "name",
			"description": "description",
			"enabled":     "enabled",
			"priority":    "priority",
		},
		SearchColumns: []string{"name", "description"},
		DefaultSort:   listing.Sort{Field: "priority", Desc: true},
	}
	return dblisting.List(ctx, pool, cfg, query, scanPolicyListItem)
}

func scanPolicyListItem(rows pgx.Rows) (policies.ListItem, error) {
	var item policies.ListItem
	err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Enabled, &item.Priority)
	return item, err
}
