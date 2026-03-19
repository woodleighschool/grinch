package admin

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/domain"
	pgutil "github.com/woodleighschool/grinch/internal/store/postgres/shared"
)

func (store *Store) ListExecutionEvents(
	ctx context.Context,
	options domain.ExecutionEventListOptions,
) ([]domain.ExecutionEventSummary, int32, error) {
	orderBy, err := pgutil.OrderBy(options.Sort, options.Order, map[string]string{
		"id":          "ee.id",
		"occurred_at": "ee.occurred_at",
		"decision":    "ee.decision",
		"file_name":   "x.file_name",
		"created_at":  "ee.created_at",
	}, []string{"ee.created_at DESC", "ee.id DESC"})
	if err != nil {
		return nil, 0, err
	}

	whereClauses := []string{
		`($1 = '' OR
  ee.file_path ILIKE $1 OR
  x.file_name ILIKE $1 OR
  x.signing_id ILIKE $1 OR
  x.team_id ILIKE $1 OR
  x.cdhash ILIKE $1 OR
  ee.executing_user ILIKE $1 OR
  m.hostname ILIKE $1)`,
	}
	args := []any{pgutil.SearchPattern(options.Search)}
	if len(options.IDs) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("ee.id = ANY($%d)", len(args)+1))
		args = append(args, options.IDs)
	}
	if options.MachineID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("ee.machine_id = $%d::uuid", len(args)+1))
		args = append(args, *options.MachineID)
	}
	if options.UserID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("u.id = $%d::uuid", len(args)+1))
		args = append(args, *options.UserID)
	}
	if options.ExecutableID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("ee.executable_id = $%d::uuid", len(args)+1))
		args = append(args, *options.ExecutableID)
	}
	if len(options.Decisions) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("ee.decision = ANY($%d)", len(args)+1))
		args = append(args, pgutil.Strings(options.Decisions))
	}
	limitParam := len(args) + 1
	offsetParam := limitParam + 1

	query := fmt.Sprintf(`
SELECT
  ee.id,
  ee.machine_id,
  ee.executable_id,
  ee.decision,
  ee.file_path,
  x.file_name,
  x.signing_id,
  ee.occurred_at,
  ee.created_at,
  COUNT(*) OVER()::INT4 AS total
FROM execution_events AS ee
JOIN machines AS m ON m.machine_id = ee.machine_id
JOIN executables AS x ON x.id = ee.executable_id
LEFT JOIN users AS u
  ON u.upn = m.primary_user
  AND m.primary_user <> ''
WHERE %s
ORDER BY %s
LIMIT NULLIF($%d::INT, 0)
OFFSET $%d
`, strings.Join(whereClauses, " AND "), orderBy, limitParam, offsetParam)

	args = append(args, options.Limit, options.Offset)

	rows, queryErr := store.store.Pool().Query(ctx, query, args...)
	if queryErr != nil {
		return nil, 0, queryErr
	}

	return pgutil.CollectRows(rows, scanExecutionEventSummary)
}

func scanExecutionEventSummary(rows pgx.Rows) (domain.ExecutionEventSummary, int32, error) {
	var (
		item         domain.ExecutionEventSummary
		decisionText string
		total        int32
	)

	scanErr := rows.Scan(
		&item.ID,
		&item.MachineID,
		&item.ExecutableID,
		&decisionText,
		&item.FilePath,
		&item.FileName,
		&item.SigningID,
		&item.OccurredAt,
		&item.CreatedAt,
		&total,
	)
	if scanErr != nil {
		return domain.ExecutionEventSummary{}, 0, scanErr
	}

	decision, parseErr := domain.ParseEventDecision(decisionText)
	if parseErr != nil {
		return domain.ExecutionEventSummary{}, 0, parseErr
	}

	item.Decision = decision
	return item, total, nil
}

func (store *Store) GetExecutionEvent(ctx context.Context, id uuid.UUID) (domain.ExecutionEvent, error) {
	row, err := store.store.Queries().GetExecutionEvent(ctx, id)
	if err != nil {
		return domain.ExecutionEvent{}, err
	}

	decision, decisionErr := domain.ParseEventDecision(row.Decision)
	if decisionErr != nil {
		return domain.ExecutionEvent{}, decisionErr
	}

	signingChain, signingChainErr := pgutil.UnmarshalSigningChain(row.SigningChain)
	if signingChainErr != nil {
		return domain.ExecutionEvent{}, signingChainErr
	}

	entitlements, entitlementsErr := pgutil.UnmarshalEntitlements(row.Entitlements)
	if entitlementsErr != nil {
		return domain.ExecutionEvent{}, entitlementsErr
	}

	return domain.ExecutionEvent{
		ID:              row.ID,
		MachineID:       row.MachineID,
		ExecutableID:    row.ExecutableID,
		Decision:        decision,
		FilePath:        row.FilePath,
		FileName:        row.FileName,
		FileSHA256:      row.FileSha256,
		FileBundleID:    row.FileBundleID,
		FileBundlePath:  row.FileBundlePath,
		SigningID:       row.SigningID,
		TeamID:          row.TeamID,
		CDHash:          row.Cdhash,
		ExecutingUser:   row.ExecutingUser,
		LoggedInUsers:   row.LoggedInUsers,
		CurrentSessions: row.CurrentSessions,
		SigningChain:    signingChain,
		Entitlements:    entitlements,
		OccurredAt:      row.OccurredAt,
		CreatedAt:       row.CreatedAt,
	}, nil
}

func (store *Store) DeleteExecutionEvent(ctx context.Context, id uuid.UUID) error {
	return store.store.Queries().DeleteExecutionEvent(ctx, id)
}
