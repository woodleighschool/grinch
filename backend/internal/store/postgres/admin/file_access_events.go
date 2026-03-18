package admin

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/domain"
	pgutil "github.com/woodleighschool/grinch/internal/store/postgres/shared"
)

func (store *Store) ListFileAccessEvents(
	ctx context.Context,
	options domain.FileAccessEventListOptions,
) ([]domain.FileAccessEventSummary, int32, error) {
	machineID, executableID := fileAccessEventListFilterValues(options)
	orderBy, err := pgutil.OrderBy(options.Sort, options.Order, map[string]string{
		"id":          "fe.id",
		"occurred_at": "fe.occurred_at",
		"decision":    "fe.decision",
		"rule_name":   "fe.rule_name",
		"target":      "fe.target",
		"created_at":  "fe.created_at",
	}, []string{"fe.created_at DESC", "fe.id DESC"})
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
SELECT
  fe.id,
  fe.machine_id,
  fe.executable_id,
  fe.decision,
  fe.rule_name,
  fe.target,
  COALESCE(x.file_name, '') AS file_name,
  COALESCE(x.file_sha256, '') AS file_sha256,
  COALESCE(x.signing_id, '') AS signing_id,
  COALESCE(x.team_id, '') AS team_id,
  COALESCE(x.cdhash, '') AS cdhash,
  fe.occurred_at,
  fe.created_at,
  COUNT(*) OVER()::INT4 AS total
FROM file_access_events AS fe
JOIN machines AS m ON m.machine_id = fe.machine_id
LEFT JOIN executables AS x ON x.id = fe.executable_id
WHERE ($1 = '' OR
  fe.target ILIKE $1 OR
  fe.rule_name ILIKE $1 OR
  x.file_name ILIKE $1 OR
  x.signing_id ILIKE $1 OR
  m.hostname ILIKE $1)
  AND ($2::uuid IS NULL OR fe.machine_id = $2::uuid)
  AND ($3::uuid IS NULL OR fe.executable_id = $3::uuid)
ORDER BY %s
LIMIT NULLIF($4::INT, 0)
OFFSET $5
`, orderBy)

	rows, queryErr := store.store.Pool().Query(
		ctx,
		query,
		pgutil.SearchPattern(options.Search),
		machineID,
		executableID,
		options.Limit,
		options.Offset,
	)
	if queryErr != nil {
		return nil, 0, queryErr
	}

	return pgutil.CollectRows(rows, scanFileAccessEventSummary)
}

func scanFileAccessEventSummary(rows pgx.Rows) (domain.FileAccessEventSummary, int32, error) {
	var (
		item         domain.FileAccessEventSummary
		decisionText string
		total        int32
	)

	scanErr := rows.Scan(
		&item.ID,
		&item.MachineID,
		&item.ExecutableID,
		&decisionText,
		&item.RuleName,
		&item.Target,
		&item.FileName,
		&item.FileSHA256,
		&item.SigningID,
		&item.TeamID,
		&item.CDHash,
		&item.OccurredAt,
		&item.CreatedAt,
		&total,
	)
	if scanErr != nil {
		return domain.FileAccessEventSummary{}, 0, scanErr
	}

	decision, parseErr := domain.ParseFileAccessDecision(decisionText)
	if parseErr != nil {
		return domain.FileAccessEventSummary{}, 0, parseErr
	}

	item.Decision = decision
	return item, total, nil
}

func fileAccessEventListFilterValues(options domain.FileAccessEventListOptions) (any, any) {
	var machineID any
	if options.MachineID != nil {
		machineID = *options.MachineID
	}

	var executableID any
	if options.ExecutableID != nil {
		executableID = *options.ExecutableID
	}

	return machineID, executableID
}

func (store *Store) GetFileAccessEvent(ctx context.Context, id uuid.UUID) (domain.FileAccessEvent, error) {
	row, err := store.queries.GetFileAccessEvent(ctx, id)
	if err != nil {
		return domain.FileAccessEvent{}, err
	}

	decision, decisionErr := domain.ParseFileAccessDecision(row.Decision)
	if decisionErr != nil {
		return domain.FileAccessEvent{}, decisionErr
	}

	processChain, processErr := pgutil.UnmarshalFileAccessProcessChain(row.ProcessChain)
	if processErr != nil {
		return domain.FileAccessEvent{}, processErr
	}

	processChain, processNameErr := store.populateProcessNames(ctx, processChain)
	if processNameErr != nil {
		return domain.FileAccessEvent{}, processNameErr
	}

	return domain.FileAccessEvent{
		ID:           row.ID,
		MachineID:    row.MachineID,
		ExecutableID: row.ExecutableID,
		RuleVersion:  row.RuleVersion,
		RuleName:     row.RuleName,
		Target:       row.Target,
		Decision:     decision,
		FileName:     row.FileName,
		FileSHA256:   row.FileSha256,
		SigningID:    row.SigningID,
		TeamID:       row.TeamID,
		CDHash:       row.Cdhash,
		ProcessChain: processChain,
		OccurredAt:   row.OccurredAt,
		CreatedAt:    row.CreatedAt,
	}, nil
}

func (store *Store) populateProcessNames(
	ctx context.Context,
	processChain []domain.FileAccessEventProcess,
) ([]domain.FileAccessEventProcess, error) {
	if len(processChain) == 0 {
		return processChain, nil
	}

	executableIDs := make([]uuid.UUID, 0, len(processChain))
	seen := make(map[uuid.UUID]struct{}, len(processChain))
	for _, process := range processChain {
		if _, exists := seen[process.ExecutableID]; exists {
			continue
		}
		seen[process.ExecutableID] = struct{}{}
		executableIDs = append(executableIDs, process.ExecutableID)
	}

	rows, err := store.queries.GetExecutableNamesByIds(ctx, executableIDs)
	if err != nil {
		return nil, err
	}

	fileNamesByID := make(map[uuid.UUID]string, len(rows))
	for _, row := range rows {
		fileNamesByID[row.ID] = row.FileName
	}

	enriched := make([]domain.FileAccessEventProcess, 0, len(processChain))
	for _, process := range processChain {
		process.FileName = fileNamesByID[process.ExecutableID]
		enriched = append(enriched, process)
	}

	return enriched, nil
}

func (store *Store) DeleteFileAccessEvent(ctx context.Context, id uuid.UUID) error {
	return store.queries.DeleteFileAccessEvent(ctx, id)
}
