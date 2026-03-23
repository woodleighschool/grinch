package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/santa/model"
	"github.com/woodleighschool/grinch/internal/store/db"
)

var (
	machineListSortColumns = map[string]string{ //nolint:gochecknoglobals // package-level lookup table, not mutable state
		"id":               "m.id",
		"hostname":         "m.hostname",
		"serial_number":    "m.serial_number",
		"model_identifier": "m.model_identifier",
		"os_version":       "m.os_version",
		"created_at":       "m.created_at",
		"updated_at":       "m.updated_at",
		"last_seen_at":     "m.last_seen_at",
	}

	machineListDefaultOrder = []string{ //nolint:gochecknoglobals // package-level lookup table, not mutable state
		"m.last_seen_at DESC",
		"m.id ASC",
	}
)

func (s *Store) ListMachines(
	ctx context.Context,
	opts domain.MachineListOptions,
) ([]domain.MachineSummary, int32, error) {
	orderBy, err := orderBy(
		opts.Sort,
		opts.Order,
		machineListSortColumns,
		machineListDefaultOrder,
	)
	if err != nil {
		return nil, 0, err
	}

	where := []string{
		`($1 = '' OR
  m.hostname ILIKE $1 OR
  m.serial_number ILIKE $1 OR
  m.model_identifier ILIKE $1 OR
  m.os_version ILIKE $1 OR
  m.primary_user ILIKE $1 OR
  COALESCE(u.display_name, '') ILIKE $1)`,
	}
	args := []any{searchPattern(opts.Search)}

	if len(opts.IDs) > 0 {
		where = append(where, fmt.Sprintf("m.id = ANY($%d)", len(args)+1))
		args = append(args, opts.IDs)
	}
	if opts.UserID != nil {
		where = append(where, fmt.Sprintf("u.id = $%d::uuid", len(args)+1))
		args = append(args, *opts.UserID)
	}
	if len(opts.RuleSyncStatuses) > 0 {
		where = append(where, fmt.Sprintf(`machine_rule_sync_status(
  ms.pending_preflight_at,
  ms.desired_targets,
  ms.applied_targets,
  ms.desired_binary_rule_count,
  ms.binary_rule_count,
  ms.desired_certificate_rule_count,
  ms.certificate_rule_count,
  ms.desired_teamid_rule_count,
  ms.teamid_rule_count,
  ms.desired_signingid_rule_count,
  ms.signingid_rule_count,
  ms.desired_cdhash_rule_count,
  ms.cdhash_rule_count,
  ms.last_clean_sync_at,
  ms.last_reported_counts_match_at
) = ANY($%d)`, len(args)+1))
		args = append(args, toStrings(opts.RuleSyncStatuses))
	}
	if len(opts.ClientModes) > 0 {
		where = append(
			where,
			fmt.Sprintf("m.client_mode::text = ANY($%d)", len(args)+1),
		)
		args = append(args, toStrings(opts.ClientModes))
	}

	limitArg := len(args) + 1
	offsetArg := limitArg + 1

	query := fmt.Sprintf(
		machineListQuery,
		strings.Join(where, " AND "),
		orderBy,
		limitArg,
		offsetArg,
	)
	args = append(args, opts.Limit, opts.Offset)

	rows, err := s.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list machines: %w", err)
	}

	return collectRows(rows, scanMachineSummaryRow)
}

func (s *Store) GetMachine(ctx context.Context, id uuid.UUID) (domain.Machine, error) {
	row, err := s.Queries().GetMachine(ctx, id)
	if err != nil {
		return domain.Machine{}, err
	}

	return mapMachine(row)
}

func (s *Store) DeleteMachine(ctx context.Context, id uuid.UUID) error {
	return s.Queries().DeleteMachine(ctx, id)
}

func (s *Store) UpsertMachine(ctx context.Context, machine model.MachineUpsert) error {
	_, err := s.Queries().UpsertMachine(ctx, db.UpsertMachineParams{
		MachineID:         machine.MachineID,
		SerialNumber:      machine.SerialNumber,
		Hostname:          machine.Hostname,
		ModelIdentifier:   machine.ModelIdentifier,
		OsVersion:         machine.OSVersion,
		OsBuild:           machine.OSBuild,
		SantaVersion:      machine.SantaVersion,
		PrimaryUser:       machine.PrimaryUser,
		PrimaryUserGroups: machine.PrimaryUserGroups,
		ClientMode:        db.SantaClientMode(machine.ClientMode),
		LastSeenAt:        machine.LastSeenAt,
	})
	if err != nil {
		return fmt.Errorf("upsert machine: %w", err)
	}

	return nil
}

func (s *Store) GetMachineSyncState(
	ctx context.Context,
	machineID uuid.UUID,
) (model.MachineSyncState, error) {
	row, err := s.Queries().GetMachineSyncState(ctx, machineID)
	if err != nil {
		return model.MachineSyncState{}, err
	}

	return mapMachineSyncState(row)
}

func (s *Store) ReplacePendingSnapshot(
	ctx context.Context,
	snapshot model.PendingSnapshotWrite,
) error {
	desiredTargets, err := json.Marshal(snapshot.DesiredTargets)
	if err != nil {
		return fmt.Errorf("marshal desired targets: %w", err)
	}

	appliedTargets, err := json.Marshal(snapshot.AppliedTargets)
	if err != nil {
		return fmt.Errorf("marshal applied targets: %w", err)
	}

	pendingTargets, err := json.Marshal(snapshot.SentTargets)
	if err != nil {
		return fmt.Errorf("marshal pending targets: %w", err)
	}

	pendingPayload, err := json.Marshal(snapshot.PendingPayload)
	if err != nil {
		return fmt.Errorf("marshal pending payload: %w", err)
	}

	err = s.Queries().UpsertMachineSyncState(ctx, db.UpsertMachineSyncStateParams{
		MachineID:                   snapshot.MachineID,
		RulesHash:                   snapshot.RulesHash,
		DesiredTargets:              desiredTargets,
		AppliedTargets:              appliedTargets,
		PendingTargets:              pendingTargets,
		PendingPayload:              pendingPayload,
		PendingPayloadRuleCount:     snapshot.PendingPayloadRuleCount,
		PendingFullSync:             snapshot.PendingFullSync,
		PendingPreflightAt:          &snapshot.PendingPreflightAt,
		DesiredBinaryRuleCount:      snapshot.DesiredBinaryRuleCount,
		DesiredCertificateRuleCount: snapshot.DesiredCertificateRuleCount,
		DesiredTeamIDRuleCount:      snapshot.DesiredTeamIDRuleCount,
		DesiredSigningIDRuleCount:   snapshot.DesiredSigningIDRuleCount,
		DesiredCDHashRuleCount:      snapshot.DesiredCDHashRuleCount,
		BinaryRuleCount:             snapshot.BinaryRuleCount,
		CertificateRuleCount:        snapshot.CertificateRuleCount,
		TeamIDRuleCount:             snapshot.TeamIDRuleCount,
		SigningIDRuleCount:          snapshot.SigningIDRuleCount,
		CDHashRuleCount:             snapshot.CDHashRuleCount,
		RulesReceived:               snapshot.RulesReceived,
		RulesProcessed:              snapshot.RulesProcessed,
		LastRuleSyncAttemptAt:       snapshot.LastRuleSyncAttemptAt,
		LastRuleSyncSuccessAt:       snapshot.LastRuleSyncSuccessAt,
		LastReportedCountsMatchAt:   snapshot.LastReportedCountsMatchAt,
	})
	if err != nil {
		return fmt.Errorf("replace pending snapshot: %w", err)
	}

	return nil
}

func (s *Store) RecordPostflight(
	ctx context.Context,
	write model.PostflightWrite,
) error {
	updated, err := s.Queries().RecordMachineSyncPostflight(ctx, db.RecordMachineSyncPostflightParams{
		MachineID:             write.MachineID,
		RulesHash:             write.RulesHash,
		RulesReceived:         write.RulesReceived,
		RulesProcessed:        write.RulesProcessed,
		LastRuleSyncAttemptAt: &write.LastRuleSyncAttemptAt,
	})
	if err != nil {
		return err
	}
	if updated == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (s *Store) PromotePendingSnapshot(
	ctx context.Context,
	machineID uuid.UUID,
	completedAt time.Time,
) error {
	updated, err := s.Queries().PromoteMachineSyncPendingSnapshot(
		ctx,
		db.PromoteMachineSyncPendingSnapshotParams{
			MachineID:             machineID,
			LastRuleSyncSuccessAt: &completedAt,
		},
	)
	if err != nil {
		return err
	}
	if updated == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func scanMachineSummaryRow(rows pgx.Rows) (domain.MachineSummary, int32, error) {
	var (
		item               domain.MachineSummary
		ruleSyncStatusText string
		total              int32
	)

	if err := rows.Scan(
		&item.ID,
		&item.SerialNumber,
		&item.Hostname,
		&item.ModelIdentifier,
		&item.OSVersion,
		&item.SantaVersion,
		&item.PrimaryUser,
		&item.PrimaryUserID,
		&ruleSyncStatusText,
		&item.LastSeenAt,
		&item.CreatedAt,
		&item.UpdatedAt,
		&total,
	); err != nil {
		return domain.MachineSummary{}, 0, err
	}

	ruleSyncStatus, err := domain.ParseMachineRuleSyncStatus(ruleSyncStatusText)
	if err != nil {
		return domain.MachineSummary{}, 0, fmt.Errorf("parse machine rule sync status: %w", err)
	}

	item.RuleSyncStatus = ruleSyncStatus

	return item, total, nil
}

func mapMachine(row db.GetMachineRow) (domain.Machine, error) {
	clientMode, err := domain.ParseMachineClientMode(string(row.ClientMode))
	if err != nil {
		return domain.Machine{}, fmt.Errorf("parse machine client mode: %w", err)
	}

	ruleSyncStatus, err := domain.ParseMachineRuleSyncStatus(row.RuleSyncStatus)
	if err != nil {
		return domain.Machine{}, fmt.Errorf("parse machine rule sync status: %w", err)
	}

	return domain.Machine{
		ID:                   row.ID,
		SerialNumber:         row.SerialNumber,
		Hostname:             row.Hostname,
		ModelIdentifier:      row.ModelIdentifier,
		OSVersion:            row.OsVersion,
		OSBuild:              row.OsBuild,
		SantaVersion:         row.SantaVersion,
		PrimaryUser:          row.PrimaryUser,
		PrimaryUserID:        row.PrimaryUserID,
		RuleSyncStatus:       ruleSyncStatus,
		ClientMode:           clientMode,
		BinaryRuleCount:      row.BinaryRuleCount,
		CertificateRuleCount: row.CertificateRuleCount,
		TeamIDRuleCount:      row.TeamIDRuleCount,
		SigningIDRuleCount:   row.SigningIDRuleCount,
		CDHashRuleCount:      row.CDHashRuleCount,
		LastSeenAt:           row.LastSeenAt,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
	}, nil
}

func mapMachineSyncState(row db.GetMachineSyncStateRow) (model.MachineSyncState, error) {
	desiredTargets, err := unmarshalJSONSlice[model.AppliedRuleTarget](row.DesiredTargets)
	if err != nil {
		return model.MachineSyncState{}, fmt.Errorf("unmarshal desired targets: %w", err)
	}

	appliedTargets, err := unmarshalJSONSlice[model.AppliedRuleTarget](row.AppliedTargets)
	if err != nil {
		return model.MachineSyncState{}, fmt.Errorf("unmarshal applied targets: %w", err)
	}

	sentTargets, err := unmarshalJSONSlice[model.AppliedRuleTarget](row.PendingTargets)
	if err != nil {
		return model.MachineSyncState{}, fmt.Errorf("unmarshal pending targets: %w", err)
	}

	pendingPayload, err := unmarshalJSONSlice[model.SyncRule](row.PendingPayload)
	if err != nil {
		return model.MachineSyncState{}, fmt.Errorf("unmarshal pending payload: %w", err)
	}

	return model.MachineSyncState{
		MachineID:                   row.ID,
		RulesHash:                   row.RulesHash,
		DesiredTargets:              desiredTargets,
		AppliedTargets:              appliedTargets,
		SentTargets:                 sentTargets,
		PendingPayload:              pendingPayload,
		PendingPayloadRuleCount:     row.PendingPayloadRuleCount,
		PendingFullSync:             row.PendingFullSync,
		PendingPreflightAt:          row.PendingPreflightAt,
		DesiredBinaryRuleCount:      row.DesiredBinaryRuleCount,
		DesiredCertificateRuleCount: row.DesiredCertificateRuleCount,
		DesiredTeamIDRuleCount:      row.DesiredTeamIDRuleCount,
		DesiredSigningIDRuleCount:   row.DesiredSigningIDRuleCount,
		DesiredCDHashRuleCount:      row.DesiredCDHashRuleCount,
		BinaryRuleCount:             row.BinaryRuleCount,
		CertificateRuleCount:        row.CertificateRuleCount,
		TeamIDRuleCount:             row.TeamIDRuleCount,
		SigningIDRuleCount:          row.SigningIDRuleCount,
		CDHashRuleCount:             row.CDHashRuleCount,
		RulesReceived:               row.RulesReceived,
		RulesProcessed:              row.RulesProcessed,
		LastRuleSyncAttemptAt:       row.LastRuleSyncAttemptAt,
		LastRuleSyncSuccessAt:       row.LastRuleSyncSuccessAt,
		LastCleanSyncAt:             row.LastCleanSyncAt,
		LastReportedCountsMatchAt:   row.LastReportedCountsMatchAt,
	}, nil
}

func unmarshalJSONSlice[T any](data []byte) ([]T, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var items []T
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}

	return items, nil
}

const machineListQuery = `
SELECT
  m.id,
  m.serial_number,
  m.hostname,
  m.model_identifier,
  m.os_version,
  m.santa_version,
  m.primary_user,
  u.id AS primary_user_id,
  machine_rule_sync_status(
    ms.pending_preflight_at,
    ms.desired_targets,
    ms.applied_targets,
    ms.desired_binary_rule_count,
    ms.binary_rule_count,
    ms.desired_certificate_rule_count,
    ms.certificate_rule_count,
    ms.desired_teamid_rule_count,
    ms.teamid_rule_count,
    ms.desired_signingid_rule_count,
    ms.signingid_rule_count,
    ms.desired_cdhash_rule_count,
    ms.cdhash_rule_count,
    ms.last_clean_sync_at,
    ms.last_reported_counts_match_at
  ) AS rule_sync_status,
  m.last_seen_at,
  m.created_at,
  m.updated_at,
  COUNT(*) OVER()::INT4 AS total
FROM machines AS m
LEFT JOIN machine_sync_states AS ms
  ON ms.machine_id = m.id
LEFT JOIN users AS u
  ON u.upn = m.primary_user
  AND m.primary_user <> ''
WHERE %s
ORDER BY %s
LIMIT NULLIF($%d::INT, 0)
OFFSET $%d
`
