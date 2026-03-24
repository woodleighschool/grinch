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
	s.logger.DebugContext(
		ctx,
		"santa postflight started",
		syncLogAttrs(
			ctx,
			machineID,
			"rules_received", req.GetRulesReceived(),
			"rules_processed", req.GetRulesProcessed(),
			"rules_hash", req.GetRulesHash(),
		)...,
	)

	snapshotState, err := s.dataStore.GetMachineSyncState(ctx, machineID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.DebugContext(ctx, "santa postflight skipped missing machine", syncLogAttrs(ctx, machineID)...)
			return syncv1.PostflightResponse_builder{}.Build(), nil
		}
		s.logger.ErrorContext(
			ctx,
			"santa postflight load machine sync state failed",
			syncLogAttrs(ctx, machineID, "error", err)...)
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
		s.logger.ErrorContext(ctx, "santa postflight record failed", syncLogAttrs(ctx, machineID, "error", err)...)
		return nil, err
	}

	if snapshotState.PendingPreflightAt == nil ||
		int64(req.GetRulesProcessed()) != snapshotState.PendingPayloadRuleCount {
		s.logger.DebugContext(
			ctx,
			"santa postflight completed without promotion",
			syncLogAttrs(
				ctx,
				machineID,
				"pending_snapshot", snapshotState.PendingPreflightAt != nil,
				"pending_payload_rule_count", snapshotState.PendingPayloadRuleCount,
				"rules_processed", req.GetRulesProcessed(),
			)...,
		)
		return syncv1.PostflightResponse_builder{}.Build(), nil
	}

	if err = s.dataStore.PromotePendingSnapshot(ctx, machineID, now); err != nil {
		s.logger.ErrorContext(
			ctx,
			"santa postflight promote pending snapshot failed",
			syncLogAttrs(ctx, machineID, "error", err)...)
		return nil, err
	}

	s.logger.DebugContext(ctx, "santa postflight promoted pending snapshot", syncLogAttrs(ctx, machineID)...)
	return syncv1.PostflightResponse_builder{}.Build(), nil
}
