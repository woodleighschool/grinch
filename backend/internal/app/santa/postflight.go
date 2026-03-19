package santa

import (
	"context"
	"errors"
	"time"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/app/santa/protocol"
	"github.com/woodleighschool/grinch/internal/app/santa/snapshot"
)

// HandlePostflight promotes the pending snapshot only after the client reports
// the expected final rules hash and processed count for the frozen payload.
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

	recordErr := service.dataStore.RecordPostflight(ctx, PostflightWrite{
		MachineID:             machineID,
		RulesHash:             request.GetRulesHash(),
		RulesReceived:         protocol.SafeCount(request.GetRulesReceived()),
		RulesProcessed:        protocol.SafeCount(request.GetRulesProcessed()),
		LastRuleSyncAttemptAt: now,
	})
	if recordErr != nil {
		return nil, recordErr
	}

	if !snapshot.PostflightMatchesSnapshot(request, snapshotState) {
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
