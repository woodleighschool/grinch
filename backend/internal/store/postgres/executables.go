package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/store/db"
)

var (
	executableListSortColumns = map[string]string{ //nolint:gochecknoglobals // package-level lookup table, not mutable state
		"id":          "e.id",
		"file_name":   "e.file_name",
		"file_sha256": "e.file_sha256",
		"occurrences": "occurrences",
		"created_at":  "e.created_at",
		"signing_id":  "e.signing_id",
		"team_id":     "e.team_id",
	}

	executableListDefaultOrder = []string{ //nolint:gochecknoglobals // package-level lookup table, not mutable state
		"e.created_at DESC",
		"e.id DESC",
	}
)

func (s *Store) ListExecutables(
	ctx context.Context,
	opts domain.ListOptions,
) ([]domain.ExecutableSummary, int32, error) {
	orderBy, err := orderBy(
		opts.Sort,
		opts.Order,
		executableListSortColumns,
		executableListDefaultOrder,
	)
	if err != nil {
		return nil, 0, err
	}

	where := []string{
		`($1 = '' OR
  e.file_name ILIKE $1 OR
  e.file_sha256 ILIKE $1 OR
  e.signing_id ILIKE $1 OR
  e.team_id ILIKE $1 OR
  e.cdhash ILIKE $1)`,
	}
	args := []any{searchPattern(opts.Search)}

	if len(opts.IDs) > 0 {
		where = append(where, fmt.Sprintf("e.id = ANY($%d)", len(args)+1))
		args = append(args, opts.IDs)
	}

	limitArg := len(args) + 1
	offsetArg := limitArg + 1

	query := fmt.Sprintf(
		executableListQuery,
		strings.Join(where, " AND "),
		orderBy,
		limitArg,
		offsetArg,
	)
	args = append(args, opts.Limit, opts.Offset)

	rows, err := s.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list executables: %w", err)
	}

	return collectRows(rows, scanExecutableSummaryRow)
}

func (s *Store) GetExecutable(ctx context.Context, id uuid.UUID) (domain.Executable, error) {
	row, err := s.Queries().GetExecutable(ctx, id)
	if err != nil {
		return domain.Executable{}, err
	}

	return mapExecutable(row)
}

func scanExecutableSummaryRow(rows pgx.Rows) (domain.ExecutableSummary, int32, error) {
	var (
		item  domain.ExecutableSummary
		total int32
	)

	if err := rows.Scan(
		&item.ID,
		&item.FileSHA256,
		&item.FileName,
		&item.FileBundleID,
		&item.FileBundlePath,
		&item.SigningID,
		&item.TeamID,
		&item.CDHash,
		&item.Occurrences,
		&item.CreatedAt,
		&total,
	); err != nil {
		return domain.ExecutableSummary{}, 0, err
	}

	return item, total, nil
}

func mapExecutable(row db.GetExecutableRow) (domain.Executable, error) {
	entitlements, err := unmarshalEntitlements(row.Entitlements)
	if err != nil {
		return domain.Executable{}, fmt.Errorf("unmarshal entitlements: %w", err)
	}

	signingChain, err := unmarshalSigningChain(row.SigningChain)
	if err != nil {
		return domain.Executable{}, fmt.Errorf("unmarshal signing chain: %w", err)
	}

	return domain.Executable{
		ID:             row.ID,
		FileSHA256:     row.FileSHA256,
		FileName:       row.FileName,
		FileBundleID:   row.FileBundleID,
		FileBundlePath: row.FileBundlePath,
		SigningID:      row.SigningID,
		TeamID:         row.TeamID,
		CDHash:         row.Cdhash,
		Occurrences:    row.Occurrences,
		Entitlements:   entitlements,
		SigningChain:   signingChain,
		CreatedAt:      row.CreatedAt,
	}, nil
}

const executableListQuery = `
SELECT
  e.id,
  e.file_sha256,
  e.file_name,
  e.file_bundle_id,
  e.file_bundle_path,
  e.signing_id,
  e.team_id,
  e.cdhash,
  COALESCE(event_counts.occurrences, 0)::INT4 AS occurrences,
  e.created_at,
  COUNT(*) OVER()::INT4 AS total
FROM executables AS e
LEFT JOIN (
  SELECT
    executable_id,
    COUNT(*)::INT4 AS occurrences
  FROM execution_events
  GROUP BY executable_id
) AS event_counts
  ON event_counts.executable_id = e.id
WHERE %s
ORDER BY %s
LIMIT NULLIF($%d::INT, 0)
OFFSET $%d
`
