package admin

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/domain"
	pgutil "github.com/woodleighschool/grinch/internal/store/postgres/shared"
)

func (store *Store) ListMachines(
	ctx context.Context,
	options domain.MachineListOptions,
) ([]domain.MachineSummary, int32, error) {
	orderBy, err := pgutil.OrderBy(options.Sort, options.Order, map[string]string{
		"id":               "m.machine_id",
		"hostname":         "m.hostname",
		"serial_number":    "m.serial_number",
		"model_identifier": "m.model_identifier",
		"os_version":       "m.os_version",
		"created_at":       "m.created_at",
		"updated_at":       "m.updated_at",
		"last_seen_at":     "m.last_seen_at",
	}, []string{"m.last_seen_at DESC", "m.machine_id ASC"})
	if err != nil {
		return nil, 0, err
	}

	whereClauses := []string{
		`($1 = '' OR
  m.hostname ILIKE $1 OR
  m.serial_number ILIKE $1 OR
  m.model_identifier ILIKE $1 OR
  m.os_version ILIKE $1 OR
  m.primary_user ILIKE $1 OR
  COALESCE(u.display_name, '') ILIKE $1)`,
	}
	args := []any{pgutil.SearchPattern(options.Search)}
	if len(options.IDs) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("m.machine_id = ANY($%d)", len(args)+1))
		args = append(args, options.IDs)
	}
	if options.UserID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("u.id = $%d::uuid", len(args)+1))
		args = append(args, *options.UserID)
	}
	if len(options.RuleSyncStatuses) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf(`(
  CASE
    WHEN ms.pending_preflight_at IS NULL THEN 'synced'
    WHEN ms.last_rule_sync_attempt_at IS NULL OR ms.last_rule_sync_attempt_at < ms.pending_preflight_at THEN 'pending'
    ELSE 'issue'
  END
) = ANY($%d)`, len(args)+1))
		args = append(args, pgutil.Strings(options.RuleSyncStatuses))
	}
	if len(options.ClientModes) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("COALESCE(ms.client_mode, 'unknown') = ANY($%d)", len(args)+1))
		args = append(args, pgutil.Strings(options.ClientModes))
	}
	limitParam := len(args) + 1
	offsetParam := limitParam + 1

	query := fmt.Sprintf(machineListQuery, strings.Join(whereClauses, " AND "), orderBy, limitParam, offsetParam)
	args = append(args, options.Limit, options.Offset)

	rows, err := store.store.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return pgutil.CollectRows(rows, scanMachineSummary)
}

const machineListQuery = `
SELECT
  m.machine_id,
  m.serial_number,
  m.hostname,
  m.model_identifier,
  m.os_version,
	m.santa_version,
	m.primary_user,
	u.id AS primary_user_id,
  ms.pending_preflight_at,
  ms.last_rule_sync_attempt_at,
  m.last_seen_at,
  m.created_at,
  m.updated_at,
  COUNT(*) OVER()::INT4 AS total
FROM machines AS m
LEFT JOIN machine_sync_states AS ms
  ON ms.machine_id = m.machine_id
LEFT JOIN users AS u
  ON u.upn = m.primary_user
  AND m.primary_user <> ''
WHERE %s
ORDER BY %s
LIMIT NULLIF($%d::INT, 0)
OFFSET $%d
`

func scanMachineSummary(rows pgx.Rows) (domain.MachineSummary, int32, error) {
	var (
		item                  domain.MachineSummary
		pendingPreflightAt    *time.Time
		lastRuleSyncAttemptAt *time.Time
		total                 int32
	)

	if scanErr := rows.Scan(
		&item.ID,
		&item.SerialNumber,
		&item.Hostname,
		&item.ModelIdentifier,
		&item.OSVersion,
		&item.SantaVersion,
		&item.PrimaryUser,
		&item.PrimaryUserID,
		&pendingPreflightAt,
		&lastRuleSyncAttemptAt,
		&item.LastSeenAt,
		&item.CreatedAt,
		&item.UpdatedAt,
		&total,
	); scanErr != nil {
		return domain.MachineSummary{}, 0, scanErr
	}

	item.RuleSyncStatus = domain.DeriveMachineRuleSyncStatus(pendingPreflightAt, lastRuleSyncAttemptAt)

	return item, total, nil
}

func (store *Store) GetMachine(ctx context.Context, id uuid.UUID) (domain.Machine, error) {
	return pgutil.GetMachine(ctx, store.store.Queries(), id)
}

func (store *Store) DeleteMachine(ctx context.Context, id uuid.UUID) error {
	return store.store.Queries().DeleteMachine(ctx, id)
}
