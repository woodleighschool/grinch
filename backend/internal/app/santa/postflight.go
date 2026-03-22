package santa

import (
	"context"
	"errors"
	"time"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/santa/model"
	"github.com/woodleighschool/grinch/internal/santa/snapshot"
)

// HandlePostflight promotes the pending snapshot only after the client reports
// it processed the frozen payload for this sync cycle.
func (service *Service) HandlePostflight(
	ctx context.Context,
	machineID uuid.UUID,
	request *syncv1.PostflightRequest,
) (*syncv1.PostflightResponse, error) {
	now := time.Now().UTC()
	snapshotState, stateErr := service.dataStore.GetMachineSyncState(ctx, machineID)
	if stateErr != nil {
		if errors.Is(stateErr, pgx.ErrNoRows) {
			return syncv1.PostflightResponse_builder{}.Build(), nil
		}
		return nil, stateErr
	}

	recordErr := service.dataStore.RecordPostflight(ctx, model.PostflightWrite{
		MachineID:             machineID,
		RulesHash:             request.GetRulesHash(),
		RulesReceived:         snapshot.SafeCount(request.GetRulesReceived()),
		RulesProcessed:        snapshot.SafeCount(request.GetRulesProcessed()),
		LastRuleSyncAttemptAt: now,
	})
	if recordErr != nil {
		return nil, recordErr
	}

	if snapshotState.PendingPreflightAt == nil ||
		int64(request.GetRulesProcessed()) != snapshotState.PendingPayloadRuleCount {
		return syncv1.PostflightResponse_builder{}.Build(), nil
	}

	if promoteErr := service.dataStore.PromotePendingSnapshot(
		ctx,
		machineID,
		now,
	); promoteErr != nil {
		return nil, promoteErr
	}

	return syncv1.PostflightResponse_builder{}.Build(), nil
}
