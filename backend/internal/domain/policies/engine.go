package policies

import (
	"context"

	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain/machines"
)

func resolveForMachine(ctx context.Context, svc Service, machine machines.Machine) (Policy, error) {
	userID := uuid.Nil
	if machine.UserID != nil {
		userID = *machine.UserID
	}

	groupIDs, err := lookupGroupIDs(ctx, svc.groups, userID)
	if err != nil {
		return Policy{}, err
	}

	enabled, err := svc.ListEnabled(ctx)
	if err != nil {
		return Policy{}, err
	}
	if len(enabled) == 0 {
		return Policy{}, nil
	}

	policyIDs := extractPolicyIDs(enabled)

	targets, err := svc.ListPolicyTargetsByPolicyIDs(ctx, policyIDs)
	if err != nil {
		return Policy{}, err
	}

	targetsByPolicy := groupTargetsByPolicy(targets)
	groupSet := makeUUIDSet(groupIDs)

	return selectBestPolicy(machine, userID, groupSet, enabled, targetsByPolicy), nil
}

func lookupGroupIDs(ctx context.Context, groups GroupLookup, userID uuid.UUID) ([]uuid.UUID, error) {
	if groups == nil || userID == uuid.Nil {
		return nil, nil
	}
	return groups.GroupIDsForUser(ctx, userID)
}

func extractPolicyIDs(policies []Policy) []uuid.UUID {
	ids := make([]uuid.UUID, len(policies))
	for i := range policies {
		ids[i] = policies[i].ID
	}
	return ids
}

func groupTargetsByPolicy(targets []Target) map[uuid.UUID][]Target {
	byPolicy := make(map[uuid.UUID][]Target)
	for _, t := range targets {
		byPolicy[t.PolicyID] = append(byPolicy[t.PolicyID], t)
	}
	return byPolicy
}

func makeUUIDSet(ids []uuid.UUID) map[uuid.UUID]bool {
	set := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set
}

func selectBestPolicy(
	machine machines.Machine,
	userID uuid.UUID,
	groupSet map[uuid.UUID]bool,
	policies []Policy,
	targetsByPolicy map[uuid.UUID][]Target,
) Policy {
	var best Policy
	var found bool

	for _, p := range policies {
		if !matchesAnyTarget(machine, userID, groupSet, targetsByPolicy[p.ID]) {
			continue
		}
		if !found || p.Priority > best.Priority {
			best = p
			found = true
		}
	}

	return best
}

func matchesAnyTarget(machine machines.Machine, userID uuid.UUID, groupSet map[uuid.UUID]bool, targets []Target) bool {
	for _, t := range targets {
		if matchesTarget(machine, userID, groupSet, t) {
			return true
		}
	}
	return false
}

func matchesTarget(machine machines.Machine, userID uuid.UUID, groupSet map[uuid.UUID]bool, target Target) bool {
	refID := target.RefID
	switch target.Kind {
	case TargetAll:
		return true
	case TargetMachine:
		return refID != nil && *refID == machine.ID
	case TargetUser:
		return refID != nil && userID != uuid.Nil && *refID == userID
	case TargetGroup:
		return refID != nil && groupSet[*refID]
	default:
		return false
	}
}
