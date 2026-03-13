package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/store/db"
	pgutil "github.com/woodleighschool/grinch/internal/store/postgres/shared"
)

type signingChainRecord struct {
	CommonName         string    `json:"common_name"`
	Organization       string    `json:"organization"`
	OrganizationalUnit string    `json:"organizational_unit"`
	SHA256             string    `json:"sha256"`
	ValidFrom          time.Time `json:"valid_from"`
	ValidUntil         time.Time `json:"valid_until"`
}

type fileAccessProcessRecord struct {
	Pid          int32     `json:"pid"`
	FilePath     string    `json:"file_path"`
	ExecutableID uuid.UUID `json:"executable_id"`
}

func (store *Store) ListExecutables(
	ctx context.Context,
	options domain.ExecutableListOptions,
) ([]domain.ExecutableSummary, int32, error) {
	orderBy, err := pgutil.OrderBy(options.Sort, options.Order, map[string]string{
		"id":          "e.id",
		"file_name":   "e.file_name",
		"file_sha256": "e.file_sha256",
		"created_at":  "e.created_at",
		"signing_id":  "e.signing_id",
		"team_id":     "e.team_id",
	}, []string{"e.created_at DESC", "e.id DESC"})
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
SELECT
  e.id,
  e.source,
  e.file_sha256,
  e.file_name,
  e.file_path,
  e.file_bundle_id,
  e.file_bundle_path,
  e.signing_id,
  e.team_id,
  e.cdhash,
  e.created_at,
  COUNT(*) OVER()::INT4 AS total
FROM executables AS e
WHERE ($1 = '' OR
  e.file_name ILIKE $1 OR
  e.file_path ILIKE $1 OR
  e.file_sha256 ILIKE $1 OR
  e.signing_id ILIKE $1 OR
  e.team_id ILIKE $1 OR
  e.cdhash ILIKE $1)
ORDER BY %s
LIMIT NULLIF($2::INT, 0)
OFFSET $3
`, orderBy)

	rows, queryErr := store.store.Pool().Query(
		ctx,
		query,
		pgutil.SearchPattern(options.Search),
		options.Limit,
		options.Offset,
	)
	if queryErr != nil {
		return nil, 0, queryErr
	}

	return pgutil.CollectRows(rows, func(rows pgx.Rows) (domain.ExecutableSummary, int32, error) {
		var item domain.ExecutableSummary
		var total int32

		scanErr := rows.Scan(
			&item.ID,
			&item.Source,
			&item.FileSHA256,
			&item.FileName,
			&item.FilePath,
			&item.FileBundleID,
			&item.FileBundlePath,
			&item.SigningID,
			&item.TeamID,
			&item.CDHash,
			&item.CreatedAt,
			&total,
		)
		if scanErr != nil {
			return domain.ExecutableSummary{}, 0, scanErr
		}

		return item, total, nil
	})
}

func (store *Store) GetExecutable(ctx context.Context, id uuid.UUID) (domain.Executable, error) {
	row, err := store.queries.GetExecutable(ctx, id)
	if err != nil {
		return domain.Executable{}, err
	}

	return mapExecutable(row)
}

func mapExecutable(row db.Executable) (domain.Executable, error) {
	entitlements, entitlementsErr := unmarshalEntitlements(row.Entitlements)
	if entitlementsErr != nil {
		return domain.Executable{}, entitlementsErr
	}

	signingChain, signingChainErr := unmarshalSigningChain(row.SigningChain)
	if signingChainErr != nil {
		return domain.Executable{}, signingChainErr
	}

	return domain.Executable{
		ID:             row.ID,
		Source:         domain.ExecutableSource(row.Source),
		FileSHA256:     row.FileSha256,
		FileName:       row.FileName,
		FilePath:       row.FilePath,
		FileBundleID:   row.FileBundleID,
		FileBundlePath: row.FileBundlePath,
		SigningID:      row.SigningID,
		TeamID:         row.TeamID,
		CDHash:         row.Cdhash,
		Entitlements:   entitlements,
		SigningChain:   signingChain,
		CreatedAt:      row.CreatedAt,
	}, nil
}

func unmarshalEntitlements(raw []byte) (map[string]domain.Entitlement, error) {
	records := make(map[string]any)
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &records); err != nil {
			return nil, fmt.Errorf("decode entitlements: %w", err)
		}
	}

	entitlements := make(map[string]domain.Entitlement, len(records))
	for key, value := range records {
		entitlements[key] = domain.Entitlement{Value: value}
	}

	return entitlements, nil
}

func unmarshalSigningChain(raw []byte) ([]domain.SigningChainEntry, error) {
	records := make([]signingChainRecord, 0)
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &records); err != nil {
			return nil, fmt.Errorf("decode signing chain: %w", err)
		}
	}

	signingChain := make([]domain.SigningChainEntry, 0, len(records))
	for _, record := range records {
		signingChain = append(signingChain, domain.SigningChainEntry{
			CommonName:         record.CommonName,
			Organization:       record.Organization,
			OrganizationalUnit: record.OrganizationalUnit,
			SHA256:             record.SHA256,
			ValidFrom:          record.ValidFrom.UTC(),
			ValidUntil:         record.ValidUntil.UTC(),
		})
	}

	return signingChain, nil
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
			Pid:          record.Pid,
			FilePath:     record.FilePath,
			ExecutableID: record.ExecutableID,
		})
	}

	return processes, nil
}
