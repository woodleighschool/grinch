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

type fileAccessProcessRecord struct {
	Pid        int32  `json:"pid"`
	FilePath   string `json:"file_path"`
	FileName   string `json:"file_name"`
	FileSHA256 string `json:"file_sha256"`
	SigningID  string `json:"signing_id"`
	TeamID     string `json:"team_id"`
	CDHash     string `json:"cdhash"`
}

func marshalFileAccessProcessChain(processes []domain.FileAccessEventProcess) ([]byte, error) {
	records := make([]fileAccessProcessRecord, 0, len(processes))
	for _, process := range processes {
		records = append(records, fileAccessProcessRecord{
			Pid:        process.Pid,
			FilePath:   process.FilePath,
			FileName:   process.FileName,
			FileSHA256: process.FileSHA256,
			SigningID:  process.SigningID,
			TeamID:     process.TeamID,
			CDHash:     process.CDHash,
		})
	}

	encoded, err := json.Marshal(records)
	if err != nil {
		return nil, fmt.Errorf("marshal file access process chain: %w", err)
	}

	return encoded, nil
}

func unmarshalFileAccessProcessChain(raw []byte) ([]domain.FileAccessEventProcess, error) {
	records := make([]fileAccessProcessRecord, 0)
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &records); err != nil {
			return nil, fmt.Errorf("decode file access process chain: %w", err)
		}
	}

	processes := make([]domain.FileAccessEventProcess, 0, len(records))
	for _, record := range records {
		processes = append(processes, domain.FileAccessEventProcess{
			Pid:        record.Pid,
			FilePath:   record.FilePath,
			FileName:   record.FileName,
			FileSHA256: record.FileSHA256,
			SigningID:  record.SigningID,
			TeamID:     record.TeamID,
			CDHash:     record.CDHash,
		})
	}

	return processes, nil
}

func (store *Store) ListFileAccessEvents(
	ctx context.Context,
	options domain.FileAccessEventListOptions,
) ([]domain.FileAccessEventSummary, int32, error) {
	orderBy, err := orderBy(options.Sort, options.Order, map[string]string{
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

	whereClauses := []string{
		`($1 = '' OR
  fe.target ILIKE $1 OR
  fe.rule_name ILIKE $1 OR
  fe.process_chain->0->>'file_name' ILIKE $1 OR
  fe.process_chain->0->>'signing_id' ILIKE $1 OR
  m.hostname ILIKE $1)`,
	}
	args := []any{searchPattern(options.Search)}
	if len(options.IDs) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("fe.id = ANY($%d)", len(args)+1))
		args = append(args, options.IDs)
	}
	if options.MachineID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("fe.machine_id = $%d::uuid", len(args)+1))
		args = append(args, *options.MachineID)
	}
	if len(options.Decisions) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("fe.decision = ANY($%d)", len(args)+1))
		args = append(args, toStrings(options.Decisions))
	}
	limitParam := len(args) + 1
	offsetParam := limitParam + 1

	query := fmt.Sprintf(`
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
JOIN machines AS m ON m.machine_id = fe.machine_id
WHERE %s
ORDER BY %s
LIMIT NULLIF($%d::INT, 0)
OFFSET $%d
`, strings.Join(whereClauses, " AND "), orderBy, limitParam, offsetParam)

	args = append(args, options.Limit, options.Offset)

	rows, queryErr := store.Pool().Query(ctx, query, args...)
	if queryErr != nil {
		return nil, 0, queryErr
	}

	return collectRows(rows, scanFileAccessEventSummary)
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

func (store *Store) GetFileAccessEvent(ctx context.Context, id uuid.UUID) (domain.FileAccessEvent, error) {
	row, err := store.Queries().GetFileAccessEvent(ctx, id)
	if err != nil {
		return domain.FileAccessEvent{}, err
	}

	decision, decisionErr := domain.ParseFileAccessDecision(row.Decision)
	if decisionErr != nil {
		return domain.FileAccessEvent{}, decisionErr
	}

	processChain, processErr := unmarshalFileAccessProcessChain(row.ProcessChain)
	if processErr != nil {
		return domain.FileAccessEvent{}, processErr
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

	if len(processChain) > 0 {
		primary := processChain[0]
		event.FileName = primary.FileName
		event.FileSHA256 = primary.FileSHA256
		event.SigningID = primary.SigningID
		event.TeamID = primary.TeamID
		event.CDHash = primary.CDHash
	}

	return event, nil
}

func (store *Store) DeleteFileAccessEvent(ctx context.Context, id uuid.UUID) error {
	return store.Queries().DeleteFileAccessEvent(ctx, id)
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
			FileName:   filepath.Base(process.FilePath),
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
		Decision:     string(event.Decision),
		ProcessChain: processChainJSON,
		OccurredAt:   event.OccurredAt,
	})
	if err != nil {
		return fmt.Errorf("ingest file access event: %w", err)
	}
	return nil
}
