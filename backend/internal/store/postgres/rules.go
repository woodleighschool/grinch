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

var (
	ruleListSortColumns = map[string]string{ //nolint:gochecknoglobals // package-level lookup table, not mutable state
		"id":          "r.id",
		"name":        "r.name",
		"description": "r.description",
		"rule_type":   "r.rule_type",
		"identifier":  "r.identifier",
		"enabled":     "r.enabled",
		"created_at":  "r.created_at",
		"updated_at":  "r.updated_at",
	}

	ruleListDefaultOrder = []string{ //nolint:gochecknoglobals // package-level lookup table, not mutable state
		"r.name ASC",
		"r.id ASC",
	}
)

func (s *Store) ListRules(
	ctx context.Context,
	opts domain.RuleListOptions,
) ([]domain.RuleSummary, int32, error) {
	orderBy, err := orderBy(opts.Sort, opts.Order, ruleListSortColumns, ruleListDefaultOrder)
	if err != nil {
		return nil, 0, err
	}

	where := []string{
		`($1 = '' OR
  r.name ILIKE $1 OR
  r.description ILIKE $1 OR
  r.identifier ILIKE $1 OR
  r.rule_type::text ILIKE $1)`,
	}
	args := []any{searchPattern(opts.Search)}

	if len(opts.IDs) > 0 {
		where = append(where, fmt.Sprintf("r.id = ANY($%d)", len(args)+1))
		args = append(args, opts.IDs)
	}
	if len(opts.Enabled) > 0 {
		where = append(where, fmt.Sprintf("r.enabled = ANY($%d)", len(args)+1))
		args = append(args, opts.Enabled)
	}
	if len(opts.RuleTypes) > 0 {
		where = append(where, fmt.Sprintf("r.rule_type = ANY($%d)", len(args)+1))
		args = append(args, toStrings(opts.RuleTypes))
	}

	limitArg := len(args) + 1
	offsetArg := limitArg + 1

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
`, strings.Join(where, " AND "), orderBy, limitArg, offsetArg)

	args = append(args, opts.Limit, opts.Offset)

	rows, err := s.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list rules: %w", err)
	}

	return collectRows(rows, scanRuleSummaryRow)
}

func (s *Store) GetRule(ctx context.Context, id uuid.UUID) (domain.Rule, error) {
	queries := s.Queries()

	row, err := queries.GetRule(ctx, id)
	if err != nil {
		return domain.Rule{}, err
	}

	targets, err := s.listRuleTargets(ctx, queries, id)
	if err != nil {
		return domain.Rule{}, err
	}

	return mapRule(row, targets)
}

func (s *Store) CreateRule(ctx context.Context, input domain.RuleWriteInput) (domain.Rule, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return domain.Rule{}, fmt.Errorf("create rule id: %w", err)
	}

	return s.writeRule(ctx, id, input, func(q *db.Queries) (db.Rule, error) {
		return q.CreateRule(ctx, db.CreateRuleParams{
			ID:            id,
			Name:          input.Name,
			Description:   input.Description,
			RuleType:      db.RuleType(input.RuleType),
			Identifier:    input.Identifier,
			CustomMessage: input.CustomMessage,
			CustomURL:     input.CustomURL,
			Enabled:       input.Enabled,
		})
	})
}

func (s *Store) UpdateRule(ctx context.Context, id uuid.UUID, input domain.RuleWriteInput) (domain.Rule, error) {
	return s.writeRule(ctx, id, input, func(q *db.Queries) (db.Rule, error) {
		return q.UpdateRule(ctx, db.UpdateRuleParams{
			ID:            id,
			Name:          input.Name,
			Description:   input.Description,
			RuleType:      db.RuleType(input.RuleType),
			Identifier:    input.Identifier,
			CustomMessage: input.CustomMessage,
			CustomURL:     input.CustomURL,
			Enabled:       input.Enabled,
		})
	})
}

func (s *Store) DeleteRule(ctx context.Context, id uuid.UUID) error {
	_, err := s.Queries().DeleteRule(ctx, id)
	return err
}

func (s *Store) ListResolvedMachineRules(
	ctx context.Context,
	machineID uuid.UUID,
) ([]domain.MachineResolvedRule, error) {
	rows, err := s.Queries().ListResolvedRulesForMachine(ctx, machineID)
	if err != nil {
		return nil, err
	}

	rules := make([]domain.MachineResolvedRule, 0, len(rows))
	var rule domain.MachineResolvedRule
	for _, row := range rows {
		rule, err = mapResolvedMachineRule(row)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

func (s *Store) writeRule(
	ctx context.Context,
	ruleID uuid.UUID,
	input domain.RuleWriteInput,
	write func(*db.Queries) (db.Rule, error),
) (domain.Rule, error) {
	var (
		row     db.Rule
		targets domain.RuleTargets
	)

	if err := s.RunInTx(ctx, func(q *db.Queries) error {
		var err error

		row, err = write(q)
		if err != nil {
			return err
		}

		if err = s.replaceRuleTargets(ctx, q, ruleID, input.Targets); err != nil {
			return err
		}

		targets, err = s.listRuleTargets(ctx, q, ruleID)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return domain.Rule{}, err
	}

	return mapRule(row, targets)
}

func scanRuleSummaryRow(rows pgx.Rows) (domain.RuleSummary, int32, error) {
	var (
		row   db.Rule
		total int32
	)

	if err := rows.Scan(
		&row.ID,
		&row.Name,
		&row.Description,
		&row.RuleType,
		&row.Identifier,
		&row.Enabled,
		&row.CreatedAt,
		&row.UpdatedAt,
		&total,
	); err != nil {
		return domain.RuleSummary{}, 0, err
	}

	rule, err := mapRuleSummary(row)
	if err != nil {
		return domain.RuleSummary{}, 0, err
	}

	return rule, total, nil
}

func mapRuleSummary(row db.Rule) (domain.RuleSummary, error) {
	ruleType, err := domain.ParseRuleType(string(row.RuleType))
	if err != nil {
		return domain.RuleSummary{}, fmt.Errorf("parse rule type: %w", err)
	}

	return domain.RuleSummary{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		RuleType:    ruleType,
		Identifier:  row.Identifier,
		Enabled:     row.Enabled,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

func mapRule(row db.Rule, targets domain.RuleTargets) (domain.Rule, error) {
	ruleType, err := domain.ParseRuleType(string(row.RuleType))
	if err != nil {
		return domain.Rule{}, fmt.Errorf("parse rule type: %w", err)
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

func mapResolvedMachineRule(row db.ListResolvedRulesForMachineRow) (domain.MachineResolvedRule, error) {
	ruleType, err := domain.ParseRuleType(string(row.RuleType))
	if err != nil {
		return domain.MachineResolvedRule{}, fmt.Errorf("parse rule type: %w", err)
	}

	policy, err := domain.ParseRulePolicy(string(row.Policy.RulePolicy))
	if err != nil {
		return domain.MachineResolvedRule{}, fmt.Errorf("parse rule policy: %w", err)
	}

	return domain.MachineResolvedRule{
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
	}, nil
}

func (s *Store) listRuleTargets(
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
		if err = appendRuleTarget(&targets, row); err != nil {
			return domain.RuleTargets{}, fmt.Errorf("append rule target: %w", err)
		}
	}

	return targets, nil
}

func (s *Store) replaceRuleTargets(
	ctx context.Context,
	queries *db.Queries,
	ruleID uuid.UUID,
	targets domain.RuleTargetsWriteInput,
) error {
	if err := queries.DeleteRuleTargetsByRule(ctx, ruleID); err != nil {
		return err
	}

	var includePriority int32
	for _, target := range targets.Include {
		targetID, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("create include target id: %w", err)
		}

		includePriority++

		if err = queries.CreateRuleTarget(ctx, db.CreateRuleTargetParams{
			ID:          targetID,
			RuleID:      ruleID,
			SubjectKind: db.RuleTargetSubjectKind(target.SubjectKind),
			SubjectID:   target.SubjectID,
			Assignment:  db.RuleTargetAssignment(domain.RuleTargetAssignmentInclude),
			Priority:    pgtype.Int4{Int32: includePriority, Valid: true},
			Policy: db.NullRulePolicy{
				RulePolicy: db.RulePolicy(target.Policy),
				Valid:      true,
			},
			CelExpression: target.CELExpression,
		}); err != nil {
			return fmt.Errorf("create include rule target: %w", err)
		}
	}

	for _, group := range targets.Exclude {
		targetID, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("create exclude target id: %w", err)
		}

		if err = queries.CreateRuleTarget(ctx, db.CreateRuleTargetParams{
			ID:            targetID,
			RuleID:        ruleID,
			SubjectKind:   db.RuleTargetSubjectKind(domain.RuleTargetSubjectKindGroup),
			SubjectID:     &group.GroupID,
			Assignment:    db.RuleTargetAssignment(domain.RuleTargetAssignmentExclude),
			Priority:      pgtype.Int4{},
			Policy:        db.NullRulePolicy{},
			CelExpression: "",
		}); err != nil {
			return fmt.Errorf("create exclude rule target: %w", err)
		}
	}

	return nil
}

func appendRuleTarget(targets *domain.RuleTargets, row db.ListRuleTargetsByRuleRow) error {
	subjectKind, err := domain.ParseRuleTargetSubjectKind(string(row.SubjectKind))
	if err != nil {
		return err
	}

	assignment, err := domain.ParseRuleTargetAssignment(string(row.Assignment))
	if err != nil {
		return err
	}

	switch assignment {
	case domain.RuleTargetAssignmentInclude:
		if !row.Policy.Valid {
			return errors.New("include rule target missing policy")
		}

		var policy domain.RulePolicy
		policy, err = domain.ParseRulePolicy(string(row.Policy.RulePolicy))
		if err != nil {
			return err
		}

		targets.Include = append(targets.Include, domain.IncludeRuleTarget{
			SubjectKind:   subjectKind,
			SubjectID:     row.SubjectID,
			SubjectName:   row.SubjectName,
			Policy:        policy,
			CELExpression: row.CelExpression,
		})

	case domain.RuleTargetAssignmentExclude:
		targets.Exclude = append(targets.Exclude, domain.ExcludedGroup{
			GroupID:   *row.SubjectID,
			GroupName: row.SubjectName,
		})

	default:
		return fmt.Errorf("unsupported rule target assignment %q", assignment)
	}

	return nil
}
