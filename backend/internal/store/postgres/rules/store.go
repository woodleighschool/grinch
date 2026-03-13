package rules

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	apprules "github.com/woodleighschool/grinch/internal/app/rules"
	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/store/db"
	pgutil "github.com/woodleighschool/grinch/internal/store/postgres/shared"
)

func (store *Store) ListRules(
	ctx context.Context,
	options domain.RuleListOptions,
) ([]domain.RuleSummary, int32, error) {
	orderBy, err := pgutil.OrderBy(options.Sort, options.Order, ruleSortColumns(), []string{"r.name ASC", "r.id ASC"})
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
SELECT
  r.id,
  r.name,
  r.description,
  r.rule_type,
  r.identifier,
  r.created_at,
  r.updated_at,
  COUNT(*) OVER()::INT4 AS total
FROM rules AS r
WHERE ($1 = '' OR
  r.name ILIKE $1 OR
  r.description ILIKE $1 OR
  r.identifier ILIKE $1 OR
  r.rule_type ILIKE $1)
ORDER BY %s
LIMIT NULLIF($2::INT, 0)
OFFSET $3
`, orderBy)

	rows, err := store.store.Pool().
		Query(ctx, query, pgutil.SearchPattern(options.Search), options.Limit, options.Offset)
	if err != nil {
		return nil, 0, err
	}

	return pgutil.CollectRows(rows, func(rows pgx.Rows) (domain.RuleSummary, int32, error) {
		var (
			ruleType string
			item     domain.RuleSummary
			total    int32
		)

		if scanErr := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&ruleType,
			&item.Identifier,
			&item.CreatedAt,
			&item.UpdatedAt,
			&total,
		); scanErr != nil {
			return domain.RuleSummary{}, 0, scanErr
		}

		parsedRuleType, parseErr := pgutil.ToRuleType(ruleType)
		if parseErr != nil {
			return domain.RuleSummary{}, 0, parseErr
		}
		item.RuleType = parsedRuleType

		return item, total, nil
	})
}

func ruleSortColumns() map[string]string {
	return map[string]string{
		"id":          "r.id",
		"name":        "r.name",
		"description": "r.description",
		"rule_type":   "r.rule_type",
		"identifier":  "r.identifier",
		"created_at":  "r.created_at",
		"updated_at":  "r.updated_at",
	}
}

func (store *Store) GetRule(ctx context.Context, id uuid.UUID) (domain.Rule, error) {
	row, err := store.queriesSet().GetRule(ctx, id)
	if err != nil {
		return domain.Rule{}, err
	}

	return mapRuleFields(
		ruleFieldsRow{
			ID:            row.ID,
			Name:          row.Name,
			Description:   row.Description,
			RuleType:      row.RuleType,
			Identifier:    row.Identifier,
			CustomMessage: row.CustomMessage,
			CustomURL:     row.CustomUrl,
			CreatedAt:     row.CreatedAt,
			UpdatedAt:     row.UpdatedAt,
		},
	)
}

func (store *Store) CreateRule(ctx context.Context, input apprules.RuleCreateInput) (domain.Rule, error) {
	ruleID, err := uuid.NewV7()
	if err != nil {
		return domain.Rule{}, fmt.Errorf("create rule id: %w", err)
	}

	row, err := store.queriesSet().CreateRule(ctx, db.CreateRuleParams{
		ID:            ruleID,
		Name:          input.Name,
		Description:   input.Description,
		RuleType:      string(input.RuleType),
		Identifier:    input.Identifier,
		CustomMessage: input.CustomMessage,
		CustomUrl:     input.CustomURL,
	})
	if err != nil {
		return domain.Rule{}, err
	}

	return mapRuleFields(
		ruleFieldsRow{
			ID:            row.ID,
			Name:          row.Name,
			Description:   row.Description,
			RuleType:      row.RuleType,
			Identifier:    row.Identifier,
			CustomMessage: row.CustomMessage,
			CustomURL:     row.CustomUrl,
			CreatedAt:     row.CreatedAt,
			UpdatedAt:     row.UpdatedAt,
		},
	)
}

func (store *Store) PatchRule(ctx context.Context, id uuid.UUID, input apprules.RulePatchInput) (domain.Rule, error) {
	current, err := store.GetRule(ctx, id)
	if err != nil {
		return domain.Rule{}, err
	}

	name := current.Name
	if input.Name != nil {
		name = *input.Name
	}
	description := current.Description
	if input.Description != nil {
		description = *input.Description
	}
	ruleType := current.RuleType
	if input.RuleType != nil {
		ruleType = *input.RuleType
	}
	identifier := current.Identifier
	if input.Identifier != nil {
		identifier = *input.Identifier
	}
	customMessage := current.CustomMessage
	if input.CustomMessage != nil {
		customMessage = *input.CustomMessage
	}
	customURL := current.CustomURL
	if input.CustomURL != nil {
		customURL = *input.CustomURL
	}

	row, err := store.queriesSet().UpdateRule(ctx, db.UpdateRuleParams{
		ID:            id,
		Name:          name,
		Description:   description,
		RuleType:      string(ruleType),
		Identifier:    identifier,
		CustomMessage: customMessage,
		CustomUrl:     customURL,
	})
	if err != nil {
		return domain.Rule{}, err
	}

	return mapRuleFields(
		ruleFieldsRow{
			ID:            row.ID,
			Name:          row.Name,
			Description:   row.Description,
			RuleType:      row.RuleType,
			Identifier:    row.Identifier,
			CustomMessage: row.CustomMessage,
			CustomURL:     row.CustomUrl,
			CreatedAt:     row.CreatedAt,
			UpdatedAt:     row.UpdatedAt,
		},
	)
}

func (store *Store) DeleteRule(ctx context.Context, id uuid.UUID) error {
	_, err := store.queriesSet().DeleteRule(ctx, id)
	return err
}

func (store *Store) ListResolvedMachineRules(
	ctx context.Context,
	machineID uuid.UUID,
) ([]domain.MachineResolvedRule, error) {
	rows, err := store.queriesSet().ListResolvedRulesForMachine(ctx, machineID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.MachineResolvedRule, 0, len(rows))
	for _, row := range rows {
		ruleType, ruleTypeErr := pgutil.ToRuleType(row.RuleType)
		if ruleTypeErr != nil {
			return nil, ruleTypeErr
		}
		policy, policyErr := pgutil.ToRulePolicy(row.Policy.String)
		if policyErr != nil {
			return nil, policyErr
		}

		result = append(result, domain.MachineResolvedRule{
			MachineRuleTarget: domain.MachineRuleTarget{
				RuleType:      ruleType,
				Identifier:    row.Identifier,
				IdentifierKey: row.IdentifierKey.String,
				Policy:        policy,
				CustomMessage: row.CustomMessage,
				CustomURL:     row.CustomUrl,
				CELExpression: row.CelExpression,
			},
			RuleID: row.ID,
		})
	}

	return result, nil
}

type ruleFieldsRow struct {
	ID            uuid.UUID
	Name          string
	Description   string
	RuleType      string
	Identifier    string
	CustomMessage string
	CustomURL     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func mapRuleFields(row ruleFieldsRow) (domain.Rule, error) {
	ruleType, err := pgutil.ToRuleType(row.RuleType)
	if err != nil {
		return domain.Rule{}, err
	}

	return domain.Rule{
		ID:            row.ID,
		Name:          row.Name,
		Description:   row.Description,
		RuleType:      ruleType,
		Identifier:    row.Identifier,
		CustomMessage: row.CustomMessage,
		CustomURL:     row.CustomURL,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}, nil
}
