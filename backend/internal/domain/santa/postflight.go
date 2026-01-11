package santa

import (
	"context"
	"fmt"
	"time"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"
	"github.com/woodleighschool/grinch/internal/domain/machines"
)

// Postflight processes the Santa postflight stage and records the applied policy state.
func (s SyncService) Postflight(
	ctx context.Context,
	machineID uuid.UUID,
	req *syncv1.PostflightRequest,
) (*syncv1.PostflightResponse, error) {
	machine, err := s.machines.Get(ctx, machineID)
	if err != nil {
		return nil, fmt.Errorf("postflight: get machine: %w", err)
	}

	machine.LastSeen = time.Now().UTC()

	// If the machine has no policy, clear any applied state and persist the update.
	if machine.PolicyID == nil {
		machine = clearAppliedState(machine)
		if _, err = s.machines.Upsert(ctx, machine); err != nil {
			return nil, fmt.Errorf("postflight: upsert machine: %w", err)
		}
		return &syncv1.PostflightResponse{}, nil
	}

	policy, err := s.policies.Get(ctx, *machine.PolicyID)
	if err != nil {
		return nil, fmt.Errorf("postflight: get policy: %w", err)
	}

	machine.AppliedPolicyID = machine.PolicyID

	// Santa only reports a rules hash when rules were applied; use it as the signal to record the rules version.
	if rulesHash := req.GetRulesHash(); rulesHash != "" {
		machine.AppliedRulesVersion = &policy.RulesVersion
	}

	if machine.AppliedRulesVersion != nil && *machine.AppliedRulesVersion == policy.RulesVersion {
		machine.PolicyStatus = machines.PolicyStatusUpToDate
	} else {
		machine.PolicyStatus = machines.PolicyStatusPending
	}

	if _, err = s.machines.Upsert(ctx, machine); err != nil {
		return nil, fmt.Errorf("postflight: upsert machine: %w", err)
	}

	return &syncv1.PostflightResponse{}, nil
}
