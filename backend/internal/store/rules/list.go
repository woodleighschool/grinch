package rules

import (
	"context"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	corerules "github.com/woodleighschool/grinch/internal/core/rules"
	"github.com/woodleighschool/grinch/internal/listing"
	dblisting "github.com/woodleighschool/grinch/internal/store/listing"
)

func listRules(ctx context.Context, pool *pgxpool.Pool, query listing.Query) ([]corerules.Rule, int64, error) {
	cfg := dblisting.Config{
		Table: "rules",
		SelectCols: []string{
			"id", "name", "description", "identifier", "rule_type",
			"custom_msg", "custom_url", "notification_app_name",
		},
		Columns: map[string]string{
			"id":          "id",
			"name":        "name",
			"description": "description",
			"identifier":  "identifier",
			"rule_type":   "rule_type",
		},
		SearchColumns: []string{"name", "description", "identifier"},
		DefaultSort:   listing.Sort{Field: "name", Desc: false},
	}
	return dblisting.List(ctx, pool, cfg, query, scanRuleListItem)
}

func scanRuleListItem(rows pgx.Rows) (corerules.Rule, error) {
	var (
		r        corerules.Rule
		ruleType int32
	)
	err := rows.Scan(
		&r.ID, &r.Name, &r.Description, &r.Identifier, &ruleType,
		&r.CustomMsg, &r.CustomURL, &r.NotificationAppName,
	)
	r.RuleType = syncv1.RuleType(ruleType)
	return r, err
}
