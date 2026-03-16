package admin

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/store/db"
	pgutil "github.com/woodleighschool/grinch/internal/store/postgres/shared"
)

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
	entitlements, entitlementsErr := pgutil.UnmarshalEntitlements(row.Entitlements)
	if entitlementsErr != nil {
		return domain.Executable{}, entitlementsErr
	}

	signingChain, signingChainErr := pgutil.UnmarshalSigningChain(row.SigningChain)
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
