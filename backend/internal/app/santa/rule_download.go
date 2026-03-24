package santa

import (
	"context"
	"errors"
	"fmt"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/santa/snapshot"
)

// HandleRuleDownload serves the frozen pending snapshot from preflight.
func (s *Service) HandleRuleDownload(
	ctx context.Context,
	machineID uuid.UUID,
	_ *syncv1.RuleDownloadRequest,
) (*syncv1.RuleDownloadResponse, error) {
	s.logger.DebugContext(ctx, "santa rule download started", syncLogAttrs(ctx, machineID)...)

	pendingSnapshot, _, err := snapshot.LoadPendingSnapshot(ctx, s.dataStore, machineID)
	if err != nil {
		if errors.Is(err, snapshot.ErrPendingSnapshotNotFound) {
			s.logger.WarnContext(ctx, "santa rule download rejected", syncLogAttrs(ctx, machineID, "error", err)...)
			return nil, fmt.Errorf("%w: %w", ErrInvalidSyncRequest, err)
		}
		s.logger.ErrorContext(ctx, "santa rule download failed", syncLogAttrs(ctx, machineID, "error", err)...)
		return nil, fmt.Errorf("get pending machine rule snapshot: %w", err)
	}

	s.logger.DebugContext(
		ctx,
		"santa rule download completed",
		syncLogAttrs(
			ctx,
			machineID,
			"payload_rule_count", pendingSnapshot.PayloadRuleCount,
			"full_sync", pendingSnapshot.FullSync,
		)...,
	)

	resp, err := snapshot.BuildRuleDownloadResponse(pendingSnapshot.Payload)
	if err != nil {
		s.logger.ErrorContext(
			ctx,
			"santa rule download build response failed",
			syncLogAttrs(ctx, machineID, "error", err)...)
		return nil, err
	}

	return resp, nil
}
