package rules

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	apprules "github.com/woodleighschool/grinch/internal/app/rules"
	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/store/db"
	"github.com/woodleighschool/grinch/internal/store/postgres"
	pgutil "github.com/woodleighschool/grinch/internal/store/postgres/shared"
)

type Store struct {
	store   *postgres.Store
	queries *db.Queries
}

func New(store *postgres.Store) *Store {
	return &Store{
		store:   store,
		queries: store.Queries(),
	}
}

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
  r.enabled,
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
			&item.Enabled,
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
		"enabled":     "r.enabled",
		"created_at":  "r.created_at",
		"updated_at":  "r.updated_at",
	}
}

func (store *Store) GetRule(ctx context.Context, id uuid.UUID) (domain.Rule, error) {
	queries := store.queries
	row, err := queries.GetRule(ctx, id)
	if err != nil {
		return domain.Rule{}, err
	}

	targets, err := store.listRuleTargets(ctx, queries, id)
	if err != nil {
		return domain.Rule{}, err
	}

	return mapRuleFields(ruleFieldsRow{
		ID:            row.ID,
		Name:          row.Name,
		Description:   row.Description,
		RuleType:      row.RuleType,
		Identifier:    row.Identifier,
		CustomMessage: row.CustomMessage,
		CustomURL:     row.CustomUrl,
		Enabled:       row.Enabled,
		Targets:       targets,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	})
}

func (store *Store) CreateRule(ctx context.Context, input apprules.WriteInput) (domain.Rule, error) {
	ruleID, err := uuid.NewV7()
	if err != nil {
		return domain.Rule{}, fmt.Errorf("create rule id: %w", err)
	}

	runErr := store.store.RunInTx(ctx, func(queries *db.Queries) error {
		_, createErr := queries.CreateRule(ctx, db.CreateRuleParams{
			ID:            ruleID,
			Name:          input.Name,
			Description:   input.Description,
			RuleType:      string(input.RuleType),
			Identifier:    input.Identifier,
			CustomMessage: input.CustomMessage,
			CustomUrl:     input.CustomURL,
			Enabled:       input.Enabled,
		})
		if createErr != nil {
			return createErr
		}

		return store.replaceRuleTargets(ctx, queries, ruleID, input.Targets)
	})
	if runErr != nil {
		return domain.Rule{}, runErr
	}

	return store.GetRule(ctx, ruleID)
}

func (store *Store) UpdateRule(ctx context.Context, id uuid.UUID, input apprules.WriteInput) (domain.Rule, error) {
	runErr := store.store.RunInTx(ctx, func(queries *db.Queries) error {
		_, updateErr := queries.UpdateRule(ctx, db.UpdateRuleParams{
			ID:            id,
			Name:          input.Name,
			Description:   input.Description,
			RuleType:      string(input.RuleType),
			Identifier:    input.Identifier,
			CustomMessage: input.CustomMessage,
			CustomUrl:     input.CustomURL,
			Enabled:       input.Enabled,
		})
		if updateErr != nil {
			return updateErr
		}

		return store.replaceRuleTargets(ctx, queries, id, input.Targets)
	})
	if runErr != nil {
		return domain.Rule{}, runErr
	}

	return store.GetRule(ctx, id)
}

func (store *Store) DeleteRule(ctx context.Context, id uuid.UUID) error {
	_, err := store.queries.DeleteRule(ctx, id)
	return err
}

func (store *Store) ListResolvedMachineRules(
	ctx context.Context,
	machineID uuid.UUID,
) ([]domain.MachineResolvedRule, error) {
	rows, err := store.queries.ListResolvedRulesForMachine(ctx, machineID)
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
	Enabled       bool
	Targets       domain.RuleTargets
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
		Enabled:       row.Enabled,
		Targets:       row.Targets,
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
	targets apprules.TargetsWriteInput,
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
	subjectKind, err := pgutil.ToRuleTargetSubjectKind(row.SubjectKind)
	if err != nil {
		return err
	}
	assignment, err := pgutil.ToRuleTargetAssignment(row.Assignment)
	if err != nil {
		return err
	}

	switch assignment {
	case domain.RuleTargetAssignmentInclude:
		if !row.Policy.Valid {
			return errors.New("include rule target missing policy")
		}
		policy, policyErr := pgutil.ToRulePolicy(row.Policy.String)
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
