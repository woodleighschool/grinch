package policies

import (
	"context"

	"github.com/google/uuid"
	"github.com/woodleighschool/grinch/internal/domain/machines"
	"github.com/woodleighschool/grinch/internal/listing"
)

const defaultPageSize = 200

// Reconciler recomputes machine policy assignments.
type Reconciler struct {
	machines machines.Service
	policies Service
	pageSize int
}

// NewReconciler returns a Reconciler with default settings.
func NewReconciler(machinesSvc machines.Service, policiesSvc Service) *Reconciler {
	return &Reconciler{
		machines: machinesSvc,
		policies: policiesSvc,
		pageSize: defaultPageSize,
	}
}

// RefreshAll recomputes policy assignments for all machines.
func (r *Reconciler) RefreshAll(ctx context.Context) error {
	offset := 0
	for {
		items, _, err := r.machines.List(ctx, listing.Query{
			Limit:  r.pageSize,
			Offset: offset,
		})
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}

		for _, item := range items {
			if err = r.refreshMachine(ctx, item); err != nil {
				return err
			}
		}

		if len(items) < r.pageSize {
			return nil
		}
		offset += r.pageSize
	}
}

func (r *Reconciler) refreshMachine(ctx context.Context, item machines.ListItem) error {
	machine := machines.Machine{
		ID:                     item.ID,
		UserID:                 item.UserID,
		PolicyID:               item.PolicyID,
		AppliedPolicyID:        item.AppliedPolicyID,
		AppliedSettingsVersion: item.AppliedSettingsVersion,
		AppliedRulesVersion:    item.AppliedRulesVersion,
	}

	policy, err := r.policies.ResolveForMachine(ctx, machine)
	if err != nil {
		return err
	}

	var desiredPolicyID *uuid.UUID
	if policy.ID != uuid.Nil {
		desiredPolicyID = &policy.ID
	}

	status := ComputePolicyStatus(machine, policy)

	if ptrEqual(item.PolicyID, desiredPolicyID) && item.PolicyStatus == status {
		return nil
	}

	return r.machines.UpdatePolicyState(ctx, item.ID, desiredPolicyID, status)
}
