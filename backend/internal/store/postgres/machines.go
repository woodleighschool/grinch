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

func (store *Store) ListMachines(
	ctx context.Context,
	options domain.MachineListOptions,
) ([]domain.MachineSummary, int32, error) {
	orderBy, err := orderBy(options.Sort, options.Order, map[string]string{
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
	args := []any{searchPattern(options.Search)}
	if len(options.IDs) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("m.machine_id = ANY($%d)", len(args)+1))
		args = append(args, options.IDs)
	}
	if options.UserID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("u.id = $%d::uuid", len(args)+1))
		args = append(args, *options.UserID)
	}
	if len(options.RuleSyncStatuses) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf(`machine_rule_sync_status(
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
		args = append(args, toStrings(options.RuleSyncStatuses))
	}
	if len(options.ClientModes) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("COALESCE(ms.client_mode, 'unknown') = ANY($%d)", len(args)+1))
		args = append(args, toStrings(options.ClientModes))
	}
	limitParam := len(args) + 1
	offsetParam := limitParam + 1

	query := fmt.Sprintf(machineListQuery, strings.Join(whereClauses, " AND "), orderBy, limitParam, offsetParam)
	args = append(args, options.Limit, options.Offset)

	rows, err := store.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return collectRows(rows, scanMachineSummary)
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
		item               domain.MachineSummary
		ruleSyncStatusText string
		total              int32
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
		&ruleSyncStatusText,
		&item.LastSeenAt,
		&item.CreatedAt,
		&item.UpdatedAt,
		&total,
	); scanErr != nil {
		return domain.MachineSummary{}, 0, scanErr
	}

	ruleSyncStatus, err := domain.ParseMachineRuleSyncStatus(ruleSyncStatusText)
	if err != nil {
		return domain.MachineSummary{}, 0, err
	}
	item.RuleSyncStatus = ruleSyncStatus

	return item, total, nil
}

func (store *Store) GetMachine(ctx context.Context, id uuid.UUID) (domain.Machine, error) {
	row, err := store.Queries().GetMachine(ctx, id)
	if err != nil {
		return domain.Machine{}, err
	}

	clientMode, err := domain.ParseMachineClientMode(row.ClientMode)
	if err != nil {
		return domain.Machine{}, err
	}
	ruleSyncStatus, err := domain.ParseMachineRuleSyncStatus(row.RuleSyncStatus)
	if err != nil {
		return domain.Machine{}, err
	}

	return domain.Machine{
		ID:                   row.MachineID,
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
		CompilerRuleCount:    row.CompilerRuleCount,
		TransitiveRuleCount:  row.TransitiveRuleCount,
		TeamIDRuleCount:      row.TeamIDRuleCount,
		SigningIDRuleCount:   row.SigningIDRuleCount,
		CDHashRuleCount:      row.CDHashRuleCount,
		LastSeenAt:           row.LastSeenAt,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
	}, nil
}

func (store *Store) DeleteMachine(ctx context.Context, id uuid.UUID) error {
	return store.Queries().DeleteMachine(ctx, id)
}

func (store *Store) UpsertMachine(ctx context.Context, machine model.MachineUpsert) error {
	return store.RunInTx(ctx, func(queries *db.Queries) error {
		_, err := queries.UpsertMachine(ctx, db.UpsertMachineParams{
			MachineID:            machine.MachineID,
			SerialNumber:         machine.SerialNumber,
			Hostname:             machine.Hostname,
			ModelIdentifier:      machine.ModelIdentifier,
			OsVersion:            machine.OSVersion,
			OsBuild:              machine.OSBuild,
			SantaVersion:         machine.SantaVersion,
			PrimaryUser:          machine.PrimaryUser,
			PrimaryUserGroupsRaw: machine.PrimaryUserGroupsRaw,
			LastSeenAt:           machine.LastSeenAt,
		})
		return err
	})
}

func (store *Store) GetMachineSyncState(
	ctx context.Context,
	machineID uuid.UUID,
) (model.MachineSyncState, error) {
	row, err := store.Queries().GetMachineSyncState(ctx, machineID)
	if err != nil {
		return model.MachineSyncState{}, err
	}
	return mapMachineSyncState(row)
}

func (store *Store) ReplacePendingSnapshot(
	ctx context.Context,
	snapshot model.PendingSnapshotWrite,
) error {
	desiredTargets, err := json.Marshal(snapshot.DesiredTargets)
	if err != nil {
		return err
	}
	appliedTargets, err := json.Marshal(snapshot.AppliedTargets)
	if err != nil {
		return err
	}
	pendingTargets, err := json.Marshal(snapshot.PendingTargets)
	if err != nil {
		return err
	}
	pendingPayload, err := json.Marshal(snapshot.PendingPayload)
	if err != nil {
		return err
	}

	_, err = store.Queries().UpsertMachineSyncState(ctx, db.UpsertMachineSyncStateParams{
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
		ClientMode:                  string(snapshot.ClientMode),
		BinaryRuleCount:             snapshot.BinaryRuleCount,
		CertificateRuleCount:        snapshot.CertificateRuleCount,
		CompilerRuleCount:           snapshot.CompilerRuleCount,
		TransitiveRuleCount:         snapshot.TransitiveRuleCount,
		TeamIDRuleCount:             snapshot.TeamIDRuleCount,
		SigningIDRuleCount:          snapshot.SigningIDRuleCount,
		CDHashRuleCount:             snapshot.CDHashRuleCount,
		RulesReceived:               snapshot.RulesReceived,
		RulesProcessed:              snapshot.RulesProcessed,
		LastRuleSyncAttemptAt:       snapshot.LastRuleSyncAttemptAt,
		LastRuleSyncSuccessAt:       snapshot.LastRuleSyncSuccessAt,
		LastReportedCountsMatchAt:   snapshot.LastReportedCountsMatchAt,
	})
	return err
}

func (store *Store) RecordPostflight(
	ctx context.Context,
	write model.PostflightWrite,
) error {
	updated, err := store.Queries().RecordMachineSyncPostflight(ctx, db.RecordMachineSyncPostflightParams{
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

func (store *Store) PromotePendingSnapshot(
	ctx context.Context,
	machineID uuid.UUID,
	completedAt time.Time,
) error {
	updated, err := store.Queries().
		PromoteMachineSyncPendingSnapshot(ctx, db.PromoteMachineSyncPendingSnapshotParams{
			MachineID:             machineID,
			LastRuleSyncSuccessAt: &completedAt,
		})
	if err != nil {
		return err
	}
	if updated == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func mapMachineSyncState(row db.GetMachineSyncStateRow) (model.MachineSyncState, error) {
	var desiredTargets []model.AppliedRuleTarget
	if len(row.DesiredTargets) > 0 {
		if err := json.Unmarshal(row.DesiredTargets, &desiredTargets); err != nil {
			return model.MachineSyncState{}, err
		}
	}
	var appliedTargets []model.AppliedRuleTarget
	if len(row.AppliedTargets) > 0 {
		if err := json.Unmarshal(row.AppliedTargets, &appliedTargets); err != nil {
			return model.MachineSyncState{}, err
		}
	}
	var pendingTargets []model.AppliedRuleTarget
	if len(row.PendingTargets) > 0 {
		if err := json.Unmarshal(row.PendingTargets, &pendingTargets); err != nil {
			return model.MachineSyncState{}, err
		}
	}
	var pendingPayload []model.SyncRule
	if len(row.PendingPayload) > 0 {
		if err := json.Unmarshal(row.PendingPayload, &pendingPayload); err != nil {
			return model.MachineSyncState{}, err
		}
	}

	clientMode, err := domain.ParseMachineClientMode(row.ClientMode)
	if err != nil {
		return model.MachineSyncState{}, err
	}

	return model.MachineSyncState{
		MachineID:                   row.MachineID,
		RulesHash:                   row.RulesHash,
		DesiredTargets:              desiredTargets,
		AppliedTargets:              appliedTargets,
		PendingTargets:              pendingTargets,
		PendingPayload:              pendingPayload,
		PendingPayloadRuleCount:     row.PendingPayloadRuleCount,
		PendingFullSync:             row.PendingFullSync,
		PendingPreflightAt:          row.PendingPreflightAt,
		DesiredBinaryRuleCount:      row.DesiredBinaryRuleCount,
		DesiredCertificateRuleCount: row.DesiredCertificateRuleCount,
		DesiredTeamIDRuleCount:      row.DesiredTeamIDRuleCount,
		DesiredSigningIDRuleCount:   row.DesiredSigningIDRuleCount,
		DesiredCDHashRuleCount:      row.DesiredCDHashRuleCount,
		ClientMode:                  clientMode,
		BinaryRuleCount:             row.BinaryRuleCount,
		CertificateRuleCount:        row.CertificateRuleCount,
		CompilerRuleCount:           row.CompilerRuleCount,
		TransitiveRuleCount:         row.TransitiveRuleCount,
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
