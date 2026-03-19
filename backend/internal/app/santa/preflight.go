package santa

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/app/santa/protocol"
	"github.com/woodleighschool/grinch/internal/app/santa/snapshot"
)

// HandlePreflight snapshots machine state and freezes the next pending sync
// snapshot. Later stages read that stored snapshot instead of recomputing live
// state.
func (service *Service) HandlePreflight(
	ctx context.Context,
	machineID uuid.UUID,
	request *syncv1.PreflightRequest,
) (*syncv1.PreflightResponse, error) {
	primaryUserGroupsRaw, err := json.Marshal(request.GetPrimaryUserGroups())
	if err != nil {
		return nil, fmt.Errorf("marshal primary user groups: %w", err)
	}

	err = service.dataStore.UpsertMachine(ctx, MachineUpsert{
		MachineID:            machineID,
		SerialNumber:         request.GetSerialNumber(),
		Hostname:             request.GetHostname(),
		ModelIdentifier:      request.GetModelIdentifier(),
		OSVersion:            request.GetOsVersion(),
		OSBuild:              request.GetOsBuild(),
		SantaVersion:         request.GetSantaVersion(),
		PrimaryUser:          request.GetPrimaryUser(),
		PrimaryUserGroupsRaw: primaryUserGroupsRaw,
		LastSeenAt:           time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("upsert machine: %w", err)
	}

	pendingSnapshot, err := snapshot.PreparePendingSnapshot(
		ctx,
		service.dataStore,
		service.ruleResolver,
		machineID,
		request,
		time.Now().UTC(),
	)
	if err != nil {
		return nil, fmt.Errorf("prepare pending rule snapshot: %w", err)
	}

	syncType := protocol.MapPendingFullSync(pendingSnapshot.FullSync)
	response := syncv1.PreflightResponse_builder{
		ClientMode: request.GetClientMode(),
		SyncType:   &syncType,
	}.Build()

	return response, nil
}
