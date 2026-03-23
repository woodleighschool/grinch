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
	pendingSnapshot, _, err := snapshot.LoadPendingSnapshot(ctx, s.dataStore, machineID)
	if err != nil {
		if errors.Is(err, snapshot.ErrPendingSnapshotNotFound) {
			return nil, fmt.Errorf("%w: %w", ErrInvalidSyncRequest, err)
		}
		return nil, fmt.Errorf("get pending machine rule snapshot: %w", err)
	}

	return snapshot.BuildRuleDownloadResponse(pendingSnapshot.Payload)
}
