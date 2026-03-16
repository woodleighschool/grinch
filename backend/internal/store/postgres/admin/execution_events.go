package admin

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/domain"
	pgutil "github.com/woodleighschool/grinch/internal/store/postgres/shared"
)

func (store *Store) ListExecutionEvents(
	ctx context.Context,
	options domain.ExecutionEventListOptions,
) ([]domain.ExecutionEventSummary, int32, error) {
	machineID, userID, executableID := executionEventListFilterValues(options)
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
WHERE ($1 = '' OR
  ee.file_path ILIKE $1 OR
  x.file_name ILIKE $1 OR
  x.signing_id ILIKE $1 OR
  x.team_id ILIKE $1 OR
  x.cdhash ILIKE $1 OR
  ee.executing_user ILIKE $1 OR
  m.hostname ILIKE $1)
  AND ($2::uuid IS NULL OR ee.machine_id = $2::uuid)
  AND ($3::uuid IS NULL OR u.id = $3::uuid)
  AND ($4::uuid IS NULL OR ee.executable_id = $4::uuid)
ORDER BY %s
LIMIT NULLIF($5::INT, 0)
OFFSET $6
`, orderBy)

	rows, queryErr := store.store.Pool().Query(
		ctx,
		query,
		pgutil.SearchPattern(options.Search),
		machineID,
		userID,
		executableID,
		options.Limit,
		options.Offset,
	)
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

func executionEventListFilterValues(options domain.ExecutionEventListOptions) (any, any, any) {
	var machineID any
	if options.MachineID != nil {
		machineID = *options.MachineID
	}

	var userID any
	if options.UserID != nil {
		userID = *options.UserID
	}

	var executableID any
	if options.ExecutableID != nil {
		executableID = *options.ExecutableID
	}

	return machineID, userID, executableID
}

func (store *Store) GetExecutionEvent(ctx context.Context, id uuid.UUID) (domain.ExecutionEvent, error) {
	row, err := store.queries.GetExecutionEvent(ctx, id)
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
	return store.queries.DeleteExecutionEvent(ctx, id)
}
