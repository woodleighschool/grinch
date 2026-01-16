package policies

import "github.com/google/uuid"

// AssignmentState represents the policy versions a machine reports.
type AssignmentState struct {
	AppliedPolicyID        *uuid.UUID
	AppliedSettingsVersion *int32
	AppliedRulesVersion    *int32
}

// Subject identifies the machine and user context for policy selection.
type Subject struct {
	MachineID uuid.UUID
	UserID    uuid.UUID
	GroupIDs  []uuid.UUID
}

// ComputeStatus derives the effective policy status by comparing applied identifiers and version markers.
func ComputeStatus(state AssignmentState, policy Policy) Status {
	if policy.ID == uuid.Nil {
		return StatusUnassigned
	}

	if state.AppliedPolicyID == nil || *state.AppliedPolicyID != policy.ID {
		return StatusPending
	}

	if state.AppliedSettingsVersion == nil || state.AppliedRulesVersion == nil {
		return StatusPending
	}

	if *state.AppliedSettingsVersion != policy.SettingsVersion ||
		*state.AppliedRulesVersion != policy.RulesVersion {
		return StatusPending
	}

	return StatusUpToDate
}

// SelectPolicy chooses the highest priority policy that targets the subject.
func SelectPolicy(subject Subject, policies []Policy, targets []PolicyTarget) Policy {
	if len(policies) == 0 {
		return Policy{}
	}

	targetsByPolicy := groupTargetsByPolicy(targets)
	groupSet := makeUUIDSet(subject.GroupIDs)

	var best Policy
	var found bool

	for _, p := range policies {
		if !p.Enabled {
			continue
		}

		if !matchesAnyTarget(subject, groupSet, targetsByPolicy[p.ID]) {
			continue
		}

		if !found || p.Priority > best.Priority {
			best = p
			found = true
		}
	}

	return best
}

func groupTargetsByPolicy(targets []PolicyTarget) map[uuid.UUID][]PolicyTarget {
	byPolicy := make(map[uuid.UUID][]PolicyTarget)
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

func matchesAnyTarget(
	subject Subject,
	groupSet map[uuid.UUID]bool,
	targets []PolicyTarget,
) bool {
	for _, t := range targets {
		if matchesTarget(subject, groupSet, t) {
			return true
		}
	}
	return false
}

func matchesTarget(
	subject Subject,
	groupSet map[uuid.UUID]bool,
	target PolicyTarget,
) bool {
	refID := target.RefID
	switch target.Kind {
	case TargetAll:
		return true
	case TargetMachine:
		return refID != nil && *refID == subject.MachineID
	case TargetUser:
		return refID != nil && subject.UserID != uuid.Nil && *refID == subject.UserID
	case TargetGroup:
		return refID != nil && groupSet[*refID]
	default:
		return false
	}
}
