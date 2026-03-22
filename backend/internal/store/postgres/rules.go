package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/store/db"
)

func (store *Store) ListRules(
	ctx context.Context,
	options domain.RuleListOptions,
) ([]domain.RuleSummary, int32, error) {
	orderBy, err := orderBy(options.Sort, options.Order, map[string]string{
		"id":          "r.id",
		"name":        "r.name",
		"description": "r.description",
		"rule_type":   "r.rule_type",
		"identifier":  "r.identifier",
		"enabled":     "r.enabled",
		"created_at":  "r.created_at",
		"updated_at":  "r.updated_at",
	}, []string{"r.name ASC", "r.id ASC"})
	if err != nil {
		return nil, 0, err
	}

	whereClauses := []string{
		`($1 = '' OR
  r.name ILIKE $1 OR
  r.description ILIKE $1 OR
  r.identifier ILIKE $1 OR
  r.rule_type ILIKE $1)`,
	}
	args := []any{searchPattern(options.Search)}
	if len(options.IDs) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("r.id = ANY($%d)", len(args)+1))
		args = append(args, options.IDs)
	}
	if len(options.Enabled) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("r.enabled = ANY($%d)", len(args)+1))
		args = append(args, options.Enabled)
	}
	if len(options.RuleTypes) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("r.rule_type = ANY($%d)", len(args)+1))
		args = append(args, toStrings(options.RuleTypes))
	}
	limitParam := len(args) + 1
	offsetParam := limitParam + 1

	query := fmt.Sprintf(`
SELECT
  r.id,
  r.name,
  r.description,
  r.rule_type,
  r.identifier,
  r.enabled,
  r.created_at,
  r.updated_at,
  COUNT(*) OVER()::INT4 AS total
FROM rules AS r
WHERE %s
ORDER BY %s
LIMIT NULLIF($%d::INT, 0)
OFFSET $%d
`, strings.Join(whereClauses, " AND "), orderBy, limitParam, offsetParam)

	args = append(args, options.Limit, options.Offset)

	rows, err := store.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return collectRows(rows, func(rows pgx.Rows) (domain.RuleSummary, int32, error) {
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
			&item.Enabled,
			&item.CreatedAt,
			&item.UpdatedAt,
			&total,
		); scanErr != nil {
			return domain.RuleSummary{}, 0, scanErr
		}

		parsedRuleType, parseErr := domain.ParseRuleType(ruleType)
		if parseErr != nil {
			return domain.RuleSummary{}, 0, parseErr
		}
		item.RuleType = parsedRuleType

		return item, total, nil
	})
}

func (store *Store) GetRule(ctx context.Context, id uuid.UUID) (domain.Rule, error) {
	queries := store.Queries()
	row, err := queries.GetRule(ctx, id)
	if err != nil {
		return domain.Rule{}, err
	}

	targets, err := store.listRuleTargets(ctx, queries, id)
	if err != nil {
		return domain.Rule{}, err
	}

	return mapRuleFields(row, targets)
}

func (store *Store) CreateRule(ctx context.Context, input domain.RuleWriteInput) (domain.Rule, error) {
	ruleID, err := uuid.NewV7()
	if err != nil {
		return domain.Rule{}, fmt.Errorf("create rule id: %w", err)
	}

	var row db.Rule
	var targets domain.RuleTargets
	runErr := store.RunInTx(ctx, func(queries *db.Queries) error {
		createdRule, createErr := queries.CreateRule(ctx, db.CreateRuleParams{
			ID:            ruleID,
			Name:          input.Name,
			Description:   input.Description,
			RuleType:      string(input.RuleType),
			Identifier:    input.Identifier,
			CustomMessage: input.CustomMessage,
			CustomURL:     input.CustomURL,
			Enabled:       input.Enabled,
		})
		if createErr != nil {
			return createErr
		}
		row = createdRule

		if replaceErr := store.replaceRuleTargets(ctx, queries, ruleID, input.Targets); replaceErr != nil {
			return replaceErr
		}
		resolvedTargets, targetErr := store.listRuleTargets(ctx, queries, ruleID)
		if targetErr != nil {
			return targetErr
		}
		targets = resolvedTargets
		return nil
	})
	if runErr != nil {
		return domain.Rule{}, runErr
	}

	return mapRuleFields(row, targets)
}

func (store *Store) UpdateRule(ctx context.Context, id uuid.UUID, input domain.RuleWriteInput) (domain.Rule, error) {
	var row db.Rule
	var targets domain.RuleTargets
	runErr := store.RunInTx(ctx, func(queries *db.Queries) error {
		updatedRule, updateErr := queries.UpdateRule(ctx, db.UpdateRuleParams{
			ID:            id,
			Name:          input.Name,
			Description:   input.Description,
			RuleType:      string(input.RuleType),
			Identifier:    input.Identifier,
			CustomMessage: input.CustomMessage,
			CustomURL:     input.CustomURL,
			Enabled:       input.Enabled,
		})
		if updateErr != nil {
			return updateErr
		}
		row = updatedRule

		if replaceErr := store.replaceRuleTargets(ctx, queries, id, input.Targets); replaceErr != nil {
			return replaceErr
		}
		resolvedTargets, targetErr := store.listRuleTargets(ctx, queries, id)
		if targetErr != nil {
			return targetErr
		}
		targets = resolvedTargets
		return nil
	})
	if runErr != nil {
		return domain.Rule{}, runErr
	}

	return mapRuleFields(row, targets)
}

func (store *Store) DeleteRule(ctx context.Context, id uuid.UUID) error {
	_, err := store.Queries().DeleteRule(ctx, id)
	return err
}

func (store *Store) ListResolvedMachineRules(
	ctx context.Context,
	machineID uuid.UUID,
) ([]domain.MachineResolvedRule, error) {
	rows, err := store.Queries().ListResolvedRulesForMachine(ctx, machineID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.MachineResolvedRule, 0, len(rows))
	for _, row := range rows {
		ruleType, ruleTypeErr := domain.ParseRuleType(row.RuleType)
		if ruleTypeErr != nil {
			return nil, ruleTypeErr
		}
		policy, policyErr := domain.ParseRulePolicy(row.Policy.String)
		if policyErr != nil {
			return nil, policyErr
		}

		result = append(result, domain.MachineResolvedRule{
			MachineRuleTarget: domain.MachineRuleTarget{
				RuleType:      ruleType,
				Identifier:    row.Identifier,
				Policy:        policy,
				CustomMessage: row.CustomMessage,
				CustomURL:     row.CustomURL,
				CELExpression: row.CelExpression,
			},
			RuleID: row.ID,
			Name:   row.Name,
		})
	}

	return result, nil
}

func mapRuleFields(row db.Rule, targets domain.RuleTargets) (domain.Rule, error) {
	ruleType, err := domain.ParseRuleType(row.RuleType)
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
		Enabled:       row.Enabled,
		Targets:       targets,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}, nil
}

func (store *Store) listRuleTargets(
	ctx context.Context,
	queries *db.Queries,
	ruleID uuid.UUID,
) (domain.RuleTargets, error) {
	rows, err := queries.ListRuleTargetsByRule(ctx, ruleID)
	if err != nil {
		return domain.RuleTargets{}, err
	}

	targets := domain.RuleTargets{
		Include: make([]domain.IncludeRuleTarget, 0, len(rows)),
		Exclude: make([]domain.ExcludedGroup, 0, len(rows)),
	}
	for _, row := range rows {
		mapErr := appendRuleTargetRow(&targets, row)
		if mapErr != nil {
			return domain.RuleTargets{}, mapErr
		}
	}

	return targets, nil
}

func (store *Store) replaceRuleTargets(
	ctx context.Context,
	queries *db.Queries,
	ruleID uuid.UUID,
	targets domain.RuleTargetsWriteInput,
) error {
	if err := queries.DeleteRuleTargetsByRule(ctx, ruleID); err != nil {
		return err
	}

	includePriority := int32(0)
	for _, target := range targets.Include {
		targetID, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("create rule target id: %w", err)
		}

		var priority pgtype.Int4
		var policy pgtype.Text
		includePriority++
		priority = pgtype.Int4{Int32: includePriority, Valid: true}
		policy = pgtype.Text{String: string(target.Policy), Valid: true}

		createErr := queries.CreateRuleTarget(ctx, db.CreateRuleTargetParams{
			ID:            targetID,
			RuleID:        ruleID,
			SubjectKind:   string(target.SubjectKind),
			SubjectID:     target.SubjectID,
			Assignment:    string(domain.RuleTargetAssignmentInclude),
			Priority:      priority,
			Policy:        policy,
			CelExpression: target.CELExpression,
		})
		if createErr != nil {
			return createErr
		}
	}

	for _, group := range targets.Exclude {
		targetID, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("create rule target id: %w", err)
		}

		createErr := queries.CreateRuleTarget(ctx, db.CreateRuleTargetParams{
			ID:            targetID,
			RuleID:        ruleID,
			SubjectKind:   string(domain.RuleTargetSubjectKindGroup),
			SubjectID:     &group.GroupID,
			Assignment:    string(domain.RuleTargetAssignmentExclude),
			Priority:      pgtype.Int4{},
			Policy:        pgtype.Text{},
			CelExpression: "",
		})
		if createErr != nil {
			return createErr
		}
	}

	return nil
}

func appendRuleTargetRow(targets *domain.RuleTargets, row db.ListRuleTargetsByRuleRow) error {
	subjectKind, err := domain.ParseRuleTargetSubjectKind(row.SubjectKind)
	if err != nil {
		return err
	}
	assignment, err := domain.ParseRuleTargetAssignment(row.Assignment)
	if err != nil {
		return err
	}

	switch assignment {
	case domain.RuleTargetAssignmentInclude:
		if !row.Policy.Valid {
			return errors.New("include rule target missing policy")
		}
		policy, policyErr := domain.ParseRulePolicy(row.Policy.String)
		if policyErr != nil {
			return policyErr
		}
		targets.Include = append(targets.Include, domain.IncludeRuleTarget{
			SubjectKind:   subjectKind,
			SubjectID:     row.SubjectID,
			SubjectName:   row.SubjectName,
			Policy:        policy,
			CELExpression: row.CelExpression,
		})
	case domain.RuleTargetAssignmentExclude:
		if subjectKind != domain.RuleTargetSubjectKindGroup {
			return errors.New("exclude rule target must be group")
		}
		if row.SubjectID == nil {
			return errors.New("exclude rule target missing group id")
		}
		targets.Exclude = append(targets.Exclude, domain.ExcludedGroup{
			GroupID:   *row.SubjectID,
			GroupName: row.SubjectName,
		})
	default:
		return fmt.Errorf("unsupported rule target assignment %q", assignment)
	}

	return nil
}
