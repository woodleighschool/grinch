package ruletargets

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	appruletargets "github.com/woodleighschool/grinch/internal/app/ruletargets"
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

func (store *Store) ListRuleTargets(
	ctx context.Context,
	options domain.RuleTargetListOptions,
) ([]domain.RuleTargetSummary, int32, error) {
	orderBy, err := pgutil.OrderBy(
		options.Sort,
		options.Order,
		ruleTargetSortColumns(),
		[]string{"r.name ASC", "rt.priority ASC NULLS LAST", "g.name ASC", "rt.id ASC"},
	)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
SELECT
  rt.id,
  rt.rule_id,
  rt.subject_kind,
  rt.subject_id,
  rt.assignment,
  rt.priority,
  rt.policy,
  rt.created_at,
  rt.updated_at,
  COUNT(*) OVER()::INT4 AS total
FROM rule_targets AS rt
JOIN rules AS r ON r.id = rt.rule_id
JOIN groups AS g ON g.id = rt.subject_id
WHERE ($1::uuid IS NULL OR rt.rule_id = $1::uuid)
  AND ($2 = '' OR rt.subject_kind = $2)
  AND ($3::uuid IS NULL OR rt.subject_id = $3::uuid)
  AND ($4 = '' OR rt.assignment = $4)
  AND ($5 = '' OR rt.policy = $5)
  AND ($6 = '' OR
    r.name ILIKE $6 OR
    r.identifier ILIKE $6 OR
    g.name ILIKE $6)
ORDER BY %s
LIMIT NULLIF($7::INT, 0)
OFFSET $8
`, orderBy)

	rows, err := store.store.Pool().Query(ctx, query, ruleTargetListArguments(options)...)
	if err != nil {
		return nil, 0, err
	}

	return pgutil.CollectRows(rows, scanRuleTargetSummary)
}

func ruleTargetSortColumns() map[string]string {
	return map[string]string{
		"id":           "rt.id",
		"subject_name": "g.name",
		"assignment":   "rt.assignment",
		"priority":     "rt.priority",
		"policy":       "rt.policy",
		"created_at":   "rt.created_at",
		"updated_at":   "rt.updated_at",
	}
}

func ruleTargetListArguments(options domain.RuleTargetListOptions) []any {
	var ruleID any
	if options.RuleID != nil {
		ruleID = *options.RuleID
	}

	var subjectID any
	if options.SubjectID != nil {
		subjectID = *options.SubjectID
	}

	subjectKind := ""
	if options.SubjectKind != nil {
		subjectKind = string(*options.SubjectKind)
	}

	assignment := ""
	if options.Assignment != nil {
		assignment = string(*options.Assignment)
	}

	policy := ""
	if options.Policy != nil {
		policy = string(*options.Policy)
	}

	return []any{
		ruleID,
		subjectKind,
		subjectID,
		assignment,
		policy,
		pgutil.SearchPattern(options.Search),
		options.Limit,
		options.Offset,
	}
}

func scanRuleTargetSummary(rows pgx.Rows) (domain.RuleTargetSummary, int32, error) {
	var (
		item            domain.RuleTargetSummary
		subjectKindText string
		assignmentText  string
		policyText      pgtype.Text
		priorityValue   pgtype.Int4
		total           int32
	)

	if scanErr := rows.Scan(
		&item.ID,
		&item.RuleID,
		&subjectKindText,
		&item.SubjectID,
		&assignmentText,
		&priorityValue,
		&policyText,
		&item.CreatedAt,
		&item.UpdatedAt,
		&total,
	); scanErr != nil {
		return domain.RuleTargetSummary{}, 0, scanErr
	}

	subjectKindValue, subjectKindErr := pgutil.ToRuleTargetSubjectKind(subjectKindText)
	if subjectKindErr != nil {
		return domain.RuleTargetSummary{}, 0, subjectKindErr
	}
	assignmentValue, assignmentErr := pgutil.ToRuleTargetAssignment(assignmentText)
	if assignmentErr != nil {
		return domain.RuleTargetSummary{}, 0, assignmentErr
	}

	item.SubjectKind = subjectKindValue
	item.Assignment = assignmentValue
	if priorityValue.Valid {
		priority := priorityValue.Int32
		item.Priority = &priority
	}
	if policyText.Valid {
		policyValue, policyErr := pgutil.ToRulePolicy(policyText.String)
		if policyErr != nil {
			return domain.RuleTargetSummary{}, 0, policyErr
		}
		item.Policy = &policyValue
	}

	return item, total, nil
}

func (store *Store) GetRuleTarget(ctx context.Context, id uuid.UUID) (domain.RuleTarget, error) {
	row, err := store.queries.GetRuleTarget(ctx, id)
	if err != nil {
		return domain.RuleTarget{}, err
	}
	return mapRuleTarget(row)
}

func (store *Store) CreateRuleTarget(ctx context.Context, input appruletargets.WriteInput) (domain.RuleTarget, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return domain.RuleTarget{}, err
	}

	row, err := store.queries.CreateRuleTarget(ctx, db.CreateRuleTargetParams{
		ID:            id,
		RuleID:        input.RuleID,
		SubjectKind:   string(domain.RuleTargetSubjectKindGroup),
		SubjectID:     input.SubjectID,
		Assignment:    string(input.Assignment),
		Priority:      nullableInt4(input.Priority),
		Policy:        nullableRulePolicy(input.Policy),
		CelExpression: input.CELExpression,
	})
	if err != nil {
		return domain.RuleTarget{}, err
	}
	return mapRuleTarget(row)
}

func (store *Store) PatchRuleTarget(
	ctx context.Context,
	id uuid.UUID,
	input appruletargets.WriteInput,
) (domain.RuleTarget, error) {
	row, err := store.queries.UpdateRuleTarget(ctx, db.UpdateRuleTargetParams{
		ID:            id,
		SubjectKind:   string(domain.RuleTargetSubjectKindGroup),
		SubjectID:     input.SubjectID,
		Assignment:    string(input.Assignment),
		Priority:      nullableInt4(input.Priority),
		Policy:        nullableRulePolicy(input.Policy),
		CelExpression: input.CELExpression,
	})
	if err != nil {
		return domain.RuleTarget{}, err
	}
	return mapRuleTarget(row)
}

func (store *Store) DeleteRuleTarget(ctx context.Context, id uuid.UUID) error {
	_, err := store.queries.DeleteRuleTarget(ctx, id)
	return err
}

func mapRuleTarget(row db.RuleTarget) (domain.RuleTarget, error) {
	subjectKind, err := pgutil.ToRuleTargetSubjectKind(row.SubjectKind)
	if err != nil {
		return domain.RuleTarget{}, err
	}
	assignment, err := pgutil.ToRuleTargetAssignment(row.Assignment)
	if err != nil {
		return domain.RuleTarget{}, err
	}

	result := domain.RuleTarget{
		ID:            row.ID,
		RuleID:        row.RuleID,
		SubjectKind:   subjectKind,
		SubjectID:     row.SubjectID,
		Assignment:    assignment,
		CELExpression: row.CelExpression,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}

	if row.Priority.Valid {
		priority := row.Priority.Int32
		result.Priority = &priority
	}
	if row.Policy.Valid {
		policy, policyErr := pgutil.ToRulePolicy(row.Policy.String)
		if policyErr != nil {
			return domain.RuleTarget{}, policyErr
		}
		result.Policy = &policy
	}

	return result, nil
}

func nullableInt4(value *int32) pgtype.Int4 {
	if value == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *value, Valid: true}
}

func nullableRulePolicy(value *domain.RulePolicy) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: string(*value), Valid: true}
}
