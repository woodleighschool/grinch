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
		return nil, fmt.Errorf("upsert machine: %w", err)
	}

	if err = s.dataStore.UpdateMachineDesiredTargets(ctx, machineID); err != nil {
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
		return nil, fmt.Errorf("prepare pending rule snapshot: %w", err)
	}

	syncType := snapshot.SyncTypeFromPendingFullSync(pendingSnapshot.FullSync)
	return syncv1.PreflightResponse_builder{
		ClientMode: req.GetClientMode(),
		SyncType:   &syncType,
	}.Build(), nil
}
