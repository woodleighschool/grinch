package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/santa/model"
	"github.com/woodleighschool/grinch/internal/store/db"
)

var (
	executionEventListSortColumns = map[string]string{ //nolint:gochecknoglobals // package-level lookup table, not mutable state
		"id":          "ee.id",
		"occurred_at": "ee.occurred_at",
		"decision":    "ee.decision",
		"file_name":   "x.file_name",
		"created_at":  "ee.created_at",
	}

	executionEventListDefaultOrder = []string{ //nolint:gochecknoglobals // package-level lookup table, not mutable state
		"ee.created_at DESC",
		"ee.id DESC",
	}
)

func (s *Store) ListExecutionEvents(
	ctx context.Context,
	opts domain.ExecutionEventListOptions,
) ([]domain.ExecutionEventSummary, int32, error) {
	orderBy, err := orderBy(
		opts.Sort,
		opts.Order,
		executionEventListSortColumns,
		executionEventListDefaultOrder,
	)
	if err != nil {
		return nil, 0, err
	}

	where := []string{
		`($1 = '' OR
  ee.file_path ILIKE $1 OR
  x.file_name ILIKE $1 OR
  x.signing_id ILIKE $1 OR
  x.team_id ILIKE $1 OR
  x.cdhash ILIKE $1 OR
  ee.executing_user ILIKE $1 OR
  m.hostname ILIKE $1)`,
	}
	args := []any{searchPattern(opts.Search)}

	if len(opts.IDs) > 0 {
		where = append(where, fmt.Sprintf("ee.id = ANY($%d)", len(args)+1))
		args = append(args, opts.IDs)
	}
	if opts.MachineID != nil {
		where = append(where, fmt.Sprintf("ee.machine_id = $%d::uuid", len(args)+1))
		args = append(args, *opts.MachineID)
	}
	if opts.UserID != nil {
		where = append(where, fmt.Sprintf("u.id = $%d::uuid", len(args)+1))
		args = append(args, *opts.UserID)
	}
	if opts.ExecutableID != nil {
		where = append(where, fmt.Sprintf("ee.executable_id = $%d::uuid", len(args)+1))
		args = append(args, *opts.ExecutableID)
	}
	if len(opts.Decisions) > 0 {
		where = append(where, fmt.Sprintf("ee.decision = ANY($%d)", len(args)+1))
		args = append(args, toStrings(opts.Decisions))
	}

	limitArg := len(args) + 1
	offsetArg := limitArg + 1

	query := fmt.Sprintf(
		executionEventListQuery,
		strings.Join(where, " AND "),
		orderBy,
		limitArg,
		offsetArg,
	)
	args = append(args, opts.Limit, opts.Offset)

	rows, err := s.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list execution events: %w", err)
	}

	return collectRows(rows, scanExecutionEventSummaryRow)
}

func (s *Store) GetExecutionEvent(ctx context.Context, id uuid.UUID) (domain.ExecutionEvent, error) {
	row, err := s.Queries().GetExecutionEvent(ctx, id)
	if err != nil {
		return domain.ExecutionEvent{}, err
	}

	return mapExecutionEvent(row)
}

func (s *Store) DeleteExecutionEvent(ctx context.Context, id uuid.UUID) error {
	return s.Queries().DeleteExecutionEvent(ctx, id)
}

func (s *Store) IngestEvents(
	ctx context.Context,
	machineID uuid.UUID,
	events []model.ExecutionEventWrite,
	fileAccessEvents []model.FileAccessEventWrite,
) error {
	if err := s.RunInTx(ctx, func(q *db.Queries) error {
		for _, event := range events {
			executableID, err := upsertEventExecutable(ctx, q, event.Executable)
			if err != nil {
				return err
			}

			if err = ingestExecutionEvent(ctx, q, machineID, executableID, event); err != nil {
				return err
			}
		}

		for _, event := range fileAccessEvents {
			if err := ingestFileAccessEvent(ctx, q, machineID, event); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return fmt.Errorf("ingest events: %w", err)
	}

	return nil
}

func scanExecutionEventSummaryRow(rows pgx.Rows) (domain.ExecutionEventSummary, int32, error) {
	var (
		item         domain.ExecutionEventSummary
		decisionText string
		total        int32
	)

	if err := rows.Scan(
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
	); err != nil {
		return domain.ExecutionEventSummary{}, 0, err
	}

	decision, err := domain.ParseExecutionDecision(decisionText)
	if err != nil {
		return domain.ExecutionEventSummary{}, 0, fmt.Errorf("parse execution event decision: %w", err)
	}

	item.Decision = decision

	return item, total, nil
}

func mapExecutionEvent(row db.GetExecutionEventRow) (domain.ExecutionEvent, error) {
	decision, err := domain.ParseExecutionDecision(string(row.Decision))
	if err != nil {
		return domain.ExecutionEvent{}, fmt.Errorf("parse execution event decision: %w", err)
	}

	signingChain, err := unmarshalSigningChain(row.SigningChain)
	if err != nil {
		return domain.ExecutionEvent{}, fmt.Errorf("unmarshal signing chain: %w", err)
	}

	entitlements, err := unmarshalEntitlements(row.Entitlements)
	if err != nil {
		return domain.ExecutionEvent{}, fmt.Errorf("unmarshal entitlements: %w", err)
	}

	return domain.ExecutionEvent{
		ID:              row.ID,
		MachineID:       row.MachineID,
		ExecutableID:    row.ExecutableID,
		Decision:        decision,
		FilePath:        row.FilePath,
		FileName:        row.FileName,
		FileSHA256:      row.FileSHA256,
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

func upsertEventExecutable(
	ctx context.Context,
	queries *db.Queries,
	exe model.ExecutableWrite,
) (uuid.UUID, error) {
	row, err := queries.GetOrCreateExecutable(ctx, db.GetOrCreateExecutableParams{
		FileSHA256:     exe.FileSHA256,
		FileName:       exe.FileName,
		FileBundleID:   exe.FileBundleID,
		FileBundlePath: exe.FileBundlePath,
		SigningID:      exe.SigningID,
		TeamID:         exe.TeamID,
		Cdhash:         exe.CDHash,
		Entitlements:   exe.Entitlements,
		SigningChain:   exe.SigningChain,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("upsert event executable: %w", err)
	}

	return row.ID, nil
}

func ingestExecutionEvent(
	ctx context.Context,
	queries *db.Queries,
	machineID uuid.UUID,
	executableID uuid.UUID,
	event model.ExecutionEventWrite,
) error {
	_, err := queries.CreateExecutionEvent(ctx, db.CreateExecutionEventParams{
		MachineID:       machineID,
		ExecutableID:    executableID,
		Decision:        db.ExecutionDecision(event.Decision),
		FilePath:        event.FilePath,
		ExecutingUser:   event.ExecutingUser,
		LoggedInUsers:   event.LoggedInUsers,
		CurrentSessions: event.CurrentSessions,
		OccurredAt:      event.OccurredAt,
	})
	if err != nil {
		return fmt.Errorf("ingest execution event: %w", err)
	}

	return nil
}

const executionEventListQuery = `
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
JOIN machines AS m
  ON m.id = ee.machine_id
JOIN executables AS x
  ON x.id = ee.executable_id
LEFT JOIN users AS u
  ON u.upn = m.primary_user
  AND m.primary_user <> ''
WHERE %s
ORDER BY %s
LIMIT NULLIF($%d::INT, 0)
OFFSET $%d
`
