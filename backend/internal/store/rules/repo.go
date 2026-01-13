// Package rules provides persistence for rule records.
package rules

import (
	"context"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/woodleighschool/grinch/internal/domain/errx"
	"github.com/woodleighschool/grinch/internal/domain/rules"
	"github.com/woodleighschool/grinch/internal/listing"
	"github.com/woodleighschool/grinch/internal/store/constraints"
	"github.com/woodleighschool/grinch/internal/store/db/sqlc"
)

// Repo persists rule records using SQLC queries.
type Repo struct {
	q    *sqlc.Queries
	pool *pgxpool.Pool
}

// New returns a Repo backed by the provided database pool.
func New(pool *pgxpool.Pool) *Repo {
	return &Repo{q: sqlc.New(pool), pool: pool}
}

// Get returns the rule with the given ID.
func (r *Repo) Get(ctx context.Context, id uuid.UUID) (rules.Rule, error) {
	row, err := r.q.GetRuleByID(ctx, id)
	if err != nil {
		return rules.Rule{}, errx.FromStore(err, nil)
	}
	return toDomainRule(row), nil
}

// GetMany returns rules for the given IDs.
func (r *Repo) GetMany(ctx context.Context, ids []uuid.UUID) ([]rules.Rule, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := r.q.ListRulesByIDs(ctx, ids)
	if err != nil {
		return nil, errx.FromStore(err, nil)
	}
	return toDomainRules(rows), nil
}

// Create inserts a new rule and returns the stored record.
func (r *Repo) Create(ctx context.Context, rule rules.Rule) (rules.Rule, error) {
	row, err := r.q.CreateRule(ctx, sqlc.CreateRuleParams{
		Name:                rule.Name,
		Description:         rule.Description,
		Identifier:          rule.Identifier,
		RuleType:            int32(rule.RuleType),
		CustomMsg:           rule.CustomMsg,
		CustomUrl:           rule.CustomURL,
		NotificationAppName: rule.NotificationAppName,
	})
	if err != nil {
		return rules.Rule{}, errx.FromStore(err, constraints.RuleFields())
	}
	return toDomainRule(row), nil
}

// Update updates an existing rule and returns the stored record.
func (r *Repo) Update(ctx context.Context, rule rules.Rule) (rules.Rule, error) {
	row, err := r.q.UpdateRuleByID(ctx, sqlc.UpdateRuleByIDParams{
		ID:                  rule.ID,
		Name:                rule.Name,
		Description:         rule.Description,
		Identifier:          rule.Identifier,
		RuleType:            int32(rule.RuleType),
		CustomMsg:           rule.CustomMsg,
		CustomUrl:           rule.CustomURL,
		NotificationAppName: rule.NotificationAppName,
	})
	if err != nil {
		return rules.Rule{}, errx.FromStore(err, constraints.RuleFields())
	}
	return toDomainRule(row), nil
}

// Delete deletes the rule with the given ID.
func (r *Repo) Delete(ctx context.Context, id uuid.UUID) error {
	return errx.FromStore(r.q.DeleteRuleByID(ctx, id), nil)
}

// List returns rules matching the query and the total result count.
func (r *Repo) List(ctx context.Context, query listing.Query) ([]rules.Rule, listing.Page, error) {
	items, total, err := listRules(ctx, r.pool, query)
	if err != nil {
		return nil, listing.Page{}, errx.FromStore(err, nil)
	}
	return items, listing.Page{Total: total}, nil
}

func toDomainRule(row sqlc.Rule) rules.Rule {
	return rules.Rule{
		ID:                  row.ID,
		Name:                row.Name,
		Description:         row.Description,
		Identifier:          row.Identifier,
		RuleType:            syncv1.RuleType(row.RuleType),
		CustomMsg:           row.CustomMsg,
		CustomURL:           row.CustomUrl,
		NotificationAppName: row.NotificationAppName,
	}
}

func toDomainRules(rows []sqlc.Rule) []rules.Rule {
	out := make([]rules.Rule, len(rows))
	for i, row := range rows {
		out[i] = toDomainRule(row)
	}
	return out
}
