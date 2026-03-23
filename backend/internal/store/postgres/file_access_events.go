package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/santa/model"
	"github.com/woodleighschool/grinch/internal/store/db"
)

var (
	fileAccessEventListSortColumns = map[string]string{ //nolint:gochecknoglobals // package-level lookup table, not mutable state
		"id":          "fe.id",
		"occurred_at": "fe.occurred_at",
		"decision":    "fe.decision",
		"rule_name":   "fe.rule_name",
		"target":      "fe.target",
		"created_at":  "fe.created_at",
	}

	fileAccessEventListDefaultOrder = []string{ //nolint:gochecknoglobals // package-level lookup table, not mutable state
		"fe.created_at DESC",
		"fe.id DESC",
	}
)

func (s *Store) ListFileAccessEvents(
	ctx context.Context,
	opts domain.FileAccessEventListOptions,
) ([]domain.FileAccessEventSummary, int32, error) {
	orderBy, err := orderBy(
		opts.Sort,
		opts.Order,
		fileAccessEventListSortColumns,
		fileAccessEventListDefaultOrder,
	)
	if err != nil {
		return nil, 0, err
	}

	where := []string{
		`($1 = '' OR
  fe.target ILIKE $1 OR
  fe.rule_name ILIKE $1 OR
  fe.process_chain->0->>'file_name' ILIKE $1 OR
  fe.process_chain->0->>'signing_id' ILIKE $1 OR
  m.hostname ILIKE $1)`,
	}
	args := []any{searchPattern(opts.Search)}

	if len(opts.IDs) > 0 {
		where = append(where, fmt.Sprintf("fe.id = ANY($%d)", len(args)+1))
		args = append(args, opts.IDs)
	}
	if opts.MachineID != nil {
		where = append(where, fmt.Sprintf("fe.machine_id = $%d::uuid", len(args)+1))
		args = append(args, *opts.MachineID)
	}
	if len(opts.Decisions) > 0 {
		where = append(where, fmt.Sprintf("fe.decision = ANY($%d)", len(args)+1))
		args = append(args, toStrings(opts.Decisions))
	}

	limitArg := len(args) + 1
	offsetArg := limitArg + 1

	query := fmt.Sprintf(
		fileAccessEventListQuery,
		strings.Join(where, " AND "),
		orderBy,
		limitArg,
		offsetArg,
	)
	args = append(args, opts.Limit, opts.Offset)

	rows, err := s.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list file access events: %w", err)
	}

	return collectRows(rows, scanFileAccessEventSummaryRow)
}

func (s *Store) GetFileAccessEvent(ctx context.Context, id uuid.UUID) (domain.FileAccessEvent, error) {
	row, err := s.Queries().GetFileAccessEvent(ctx, id)
	if err != nil {
		return domain.FileAccessEvent{}, err
	}

	return mapFileAccessEvent(row)
}

func (s *Store) DeleteFileAccessEvent(ctx context.Context, id uuid.UUID) error {
	return s.Queries().DeleteFileAccessEvent(ctx, id)
}

func ingestFileAccessEvent(
	ctx context.Context,
	queries *db.Queries,
	machineID uuid.UUID,
	event model.FileAccessEventWrite,
) error {
	processChain := make([]domain.FileAccessEventProcess, 0, len(event.Processes))
	for _, process := range event.Processes {
		processChain = append(processChain, domain.FileAccessEventProcess{
			Pid:        process.Pid,
			FilePath:   process.FilePath,
			FileName:   baseFileName(process.FilePath),
			FileSHA256: process.FileSHA256,
			SigningID:  process.SigningID,
			TeamID:     process.TeamID,
			CDHash:     process.CDHash,
		})
	}

	processChainJSON, err := marshalFileAccessProcessChain(processChain)
	if err != nil {
		return err
	}

	_, err = queries.CreateFileAccessEvent(ctx, db.CreateFileAccessEventParams{
		MachineID:    machineID,
		RuleVersion:  event.RuleVersion,
		RuleName:     event.RuleName,
		Target:       event.Target,
		Decision:     db.FileAccessDecision(event.Decision),
		ProcessChain: processChainJSON,
		OccurredAt:   event.OccurredAt,
	})
	if err != nil {
		return fmt.Errorf("ingest file access event: %w", err)
	}

	return nil
}

func scanFileAccessEventSummaryRow(rows pgx.Rows) (domain.FileAccessEventSummary, int32, error) {
	var (
		item         domain.FileAccessEventSummary
		decisionText string
		total        int32
	)

	if err := rows.Scan(
		&item.ID,
		&item.MachineID,
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
	); err != nil {
		return domain.FileAccessEventSummary{}, 0, err
	}

	decision, err := domain.ParseFileAccessDecision(decisionText)
	if err != nil {
		return domain.FileAccessEventSummary{}, 0, fmt.Errorf("parse file access decision: %w", err)
	}

	item.Decision = decision

	return item, total, nil
}

func mapFileAccessEvent(row db.FileAccessEvent) (domain.FileAccessEvent, error) {
	decision, err := domain.ParseFileAccessDecision(string(row.Decision))
	if err != nil {
		return domain.FileAccessEvent{}, fmt.Errorf("parse file access decision: %w", err)
	}

	processChain, err := unmarshalFileAccessProcessChain(row.ProcessChain)
	if err != nil {
		return domain.FileAccessEvent{}, err
	}

	event := domain.FileAccessEvent{
		ID:           row.ID,
		MachineID:    row.MachineID,
		RuleVersion:  row.RuleVersion,
		RuleName:     row.RuleName,
		Target:       row.Target,
		Decision:     decision,
		ProcessChain: processChain,
		OccurredAt:   row.OccurredAt,
		CreatedAt:    row.CreatedAt,
	}

	applyPrimaryFileAccessProcess(&event, processChain)

	return event, nil
}

func marshalFileAccessProcessChain(processes []domain.FileAccessEventProcess) ([]byte, error) {
	data, err := json.Marshal(processes)
	if err != nil {
		return nil, fmt.Errorf("marshal file access process chain: %w", err)
	}

	return data, nil
}

func unmarshalFileAccessProcessChain(data []byte) ([]domain.FileAccessEventProcess, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var processes []domain.FileAccessEventProcess
	if err := json.Unmarshal(data, &processes); err != nil {
		return nil, fmt.Errorf("unmarshal file access process chain: %w", err)
	}

	return processes, nil
}

func applyPrimaryFileAccessProcess(
	event *domain.FileAccessEvent,
	processChain []domain.FileAccessEventProcess,
) {
	if len(processChain) == 0 {
		return
	}

	primary := processChain[0]
	event.FileName = primary.FileName
	event.FileSHA256 = primary.FileSHA256
	event.SigningID = primary.SigningID
	event.TeamID = primary.TeamID
	event.CDHash = primary.CDHash
}

func baseFileName(path string) string {
	if path == "" {
		return ""
	}

	return filepath.Base(path)
}

const fileAccessEventListQuery = `
SELECT
  fe.id,
  fe.machine_id,
  fe.decision,
  fe.rule_name,
  fe.target,
  COALESCE(fe.process_chain->0->>'file_name', '') AS file_name,
  COALESCE(fe.process_chain->0->>'file_sha256', '') AS file_sha256,
  COALESCE(fe.process_chain->0->>'signing_id', '') AS signing_id,
  COALESCE(fe.process_chain->0->>'team_id', '') AS team_id,
  COALESCE(fe.process_chain->0->>'cdhash', '') AS cdhash,
  fe.occurred_at,
  fe.created_at,
  COUNT(*) OVER()::INT4 AS total
FROM file_access_events AS fe
JOIN machines AS m
  ON m.id = fe.machine_id
WHERE %s
ORDER BY %s
LIMIT NULLIF($%d::INT, 0)
OFFSET $%d
`
