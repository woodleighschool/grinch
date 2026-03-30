package santa

import (
	"context"
	"fmt"
	"time"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/santa/model"
	"github.com/woodleighschool/grinch/internal/santa/snapshot"
)

// HandlePreflight snapshots machine state and freezes the next pending sync
// snapshot. Later stages read that stored snapshot instead of recomputing live
// state.
func (s *Service) HandlePreflight(
	ctx context.Context,
	machineID uuid.UUID,
	req *syncv1.PreflightRequest,
) (*syncv1.PreflightResponse, error) {
	s.logger.DebugContext(
		ctx,
		"santa preflight started",
		syncLogAttrs(
			ctx,
			machineID,
			"hostname", req.GetHostname(),
			"serial_number", req.GetSerialNumber(),
			"client_mode", req.GetClientMode().String(),
			"request_clean_sync", req.GetRequestCleanSync(),
		)...,
	)

	err := s.dataStore.UpsertMachine(ctx, model.MachineUpsert{
		MachineID:         machineID,
		SerialNumber:      req.GetSerialNumber(),
		Hostname:          req.GetHostname(),
		ModelIdentifier:   req.GetModelIdentifier(),
		OSVersion:         req.GetOsVersion(),
		OSBuild:           req.GetOsBuild(),
		SantaVersion:      req.GetSantaVersion(),
		PrimaryUser:       req.GetPrimaryUser(),
		PrimaryUserGroups: normalizeStrings(req.GetPrimaryUserGroups()),
		ClientMode:        snapshot.MachineClientModeFromProto(req.GetClientMode()),
		LastSeenAt:        time.Now().UTC(),
	})
	if err != nil {
		s.logger.ErrorContext(
			ctx,
			"santa preflight upsert machine failed",
			syncLogAttrs(ctx, machineID, "error", err)...)
		return nil, fmt.Errorf("upsert machine: %w", err)
	}

	if err = s.dataStore.UpdateMachineDesiredTargets(ctx, machineID); err != nil {
		s.logger.ErrorContext(
			ctx,
			"santa preflight sync desired targets failed",
			syncLogAttrs(ctx, machineID, "error", err)...)
		return nil, fmt.Errorf("sync machine desired rule targets: %w", err)
	}

	pendingSnapshot, err := snapshot.PreparePendingSnapshot(
		ctx,
		s.dataStore,
		s.ruleResolver,
		machineID,
		req,
		time.Now().UTC(),
	)
	if err != nil {
		s.logger.ErrorContext(
			ctx,
			"santa preflight prepare pending snapshot failed",
			syncLogAttrs(ctx, machineID, "error", err)...)
		return nil, fmt.Errorf("prepare pending rule snapshot: %w", err)
	}

	syncType := snapshot.SyncTypeFromPendingFullSync(pendingSnapshot.FullSync)
	s.logger.DebugContext(
		ctx,
		"santa preflight completed",
		syncLogAttrs(
			ctx,
			machineID,
			"sync_type", syncType.String(),
			"payload_rule_count", pendingSnapshot.PayloadRuleCount,
			"full_sync", pendingSnapshot.FullSync,
		)...,
	)

	return syncv1.PreflightResponse_builder{
		// Does this cook the client if we send this...?
		// ClientMode: req.GetClientMode(),
		SyncType: &syncType,
	}.Build(), nil
}
