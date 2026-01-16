package sync

import (
	"time"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"

	coremachines "github.com/woodleighschool/grinch/internal/core/machines"
	"github.com/woodleighschool/grinch/internal/core/policies"
	"github.com/woodleighschool/grinch/internal/store/db/pgconv"
)

type machineBuilder struct {
	id       uuid.UUID
	existing coremachines.Machine
	now      time.Time
}

func newMachineBuilder(id uuid.UUID, existing coremachines.Machine, now time.Time) machineBuilder {
	return machineBuilder{
		id:       id,
		existing: existing,
		now:      now,
	}
}

func (b machineBuilder) fromPreflight(req *syncv1.PreflightRequest) coremachines.Machine {
	primaryUser := stringPtr(req.GetPrimaryUser())
	groups := req.GetPrimaryUserGroups()
	if groups == nil {
		groups = []string{}
	}

	machine := coremachines.Machine{
		ID:                     b.id,
		SerialNumber:           req.GetSerialNumber(),
		Hostname:               req.GetHostname(),
		Model:                  req.GetModelIdentifier(),
		OSVersion:              req.GetOsVersion(),
		OSBuild:                req.GetOsBuild(),
		SantaVersion:           req.GetSantaVersion(),
		PrimaryUser:            primaryUser,
		PrimaryUserGroups:      groups,
		PushToken:              stringPtr(req.GetPushNotificationToken()),
		SIPStatus:              pgconv.Uint32ToInt32(req.GetSipStatus()),
		ClientMode:             req.GetClientMode(),
		RequestCleanSync:       req.GetRequestCleanSync(),
		PushNotificationSync:   req.GetPushNotificationSync(),
		BinaryRuleCount:        pgconv.Uint32ToInt32(req.GetBinaryRuleCount()),
		CertificateRuleCount:   pgconv.Uint32ToInt32(req.GetCertificateRuleCount()),
		CompilerRuleCount:      pgconv.Uint32ToInt32(req.GetCompilerRuleCount()),
		TransitiveRuleCount:    pgconv.Uint32ToInt32(req.GetTransitiveRuleCount()),
		TeamIDRuleCount:        pgconv.Uint32ToInt32(req.GetTeamidRuleCount()),
		SigningIDRuleCount:     pgconv.Uint32ToInt32(req.GetSigningidRuleCount()),
		CDHashRuleCount:        pgconv.Uint32ToInt32(req.GetCdhashRuleCount()),
		RulesHash:              stringPtr(req.GetRulesHash()),
		LastSeen:               b.now,
		PolicyID:               b.existing.PolicyID,
		AppliedPolicyID:        b.existing.AppliedPolicyID,
		AppliedSettingsVersion: b.existing.AppliedSettingsVersion,
		AppliedRulesVersion:    b.existing.AppliedRulesVersion,
	}

	if same(primaryUser, b.existing.PrimaryUser) {
		machine.UserID = b.existing.UserID
	}

	return machine
}

func (b machineBuilder) clearApplied(machine coremachines.Machine) coremachines.Machine {
	machine.AppliedPolicyID = nil
	machine.AppliedSettingsVersion = nil
	machine.AppliedRulesVersion = nil
	machine.PolicyStatus = policies.StatusUnassigned
	return machine
}

func computePostflightStatus(machine coremachines.Machine, policy policies.Policy) policies.Status {
	if machine.AppliedRulesVersion != nil && *machine.AppliedRulesVersion == policy.RulesVersion {
		return policies.StatusUpToDate
	}
	return policies.StatusPending
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func same(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func nowUTC() time.Time {
	return time.Now().UTC()
}
