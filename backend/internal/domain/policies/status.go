package policies

import (
	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain/machines"
)

// ComputePolicyStatus derives the effective policy status for a machine
// by comparing applied identifiers and version markers.
func ComputePolicyStatus(machine machines.Machine, policy Policy) machines.PolicyStatus {
	if policy.ID == uuid.Nil {
		return machines.PolicyStatusUnassigned
	}

	if machine.AppliedPolicyID == nil || *machine.AppliedPolicyID != policy.ID {
		return machines.PolicyStatusPending
	}

	if machine.AppliedSettingsVersion == nil || machine.AppliedRulesVersion == nil {
		return machines.PolicyStatusPending
	}

	if *machine.AppliedSettingsVersion != policy.SettingsVersion ||
		*machine.AppliedRulesVersion != policy.RulesVersion {
		return machines.PolicyStatusPending
	}

	return machines.PolicyStatusUpToDate
}
