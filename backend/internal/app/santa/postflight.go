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
func (s *Service) HandlePostflight(
	ctx context.Context,
	machineID uuid.UUID,
	req *syncv1.PostflightRequest,
) (*syncv1.PostflightResponse, error) {
	now := time.Now().UTC()

	snapshotState, err := s.dataStore.GetMachineSyncState(ctx, machineID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return syncv1.PostflightResponse_builder{}.Build(), nil
		}
		return nil, err
	}

	err = s.dataStore.RecordPostflight(ctx, model.PostflightWrite{
		MachineID:             machineID,
		RulesHash:             req.GetRulesHash(),
		RulesReceived:         snapshot.ClampRuleCount(req.GetRulesReceived()),
		RulesProcessed:        snapshot.ClampRuleCount(req.GetRulesProcessed()),
		LastRuleSyncAttemptAt: now,
	})
	if err != nil {
		return nil, err
	}

	if snapshotState.PendingPreflightAt == nil ||
		int64(req.GetRulesProcessed()) != snapshotState.PendingPayloadRuleCount {
		return syncv1.PostflightResponse_builder{}.Build(), nil
	}

	if err = s.dataStore.PromotePendingSnapshot(ctx, machineID, now); err != nil {
		return nil, err
	}

	return syncv1.PostflightResponse_builder{}.Build(), nil
}
