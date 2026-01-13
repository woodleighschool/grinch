package santa

import (
	"context"
	"fmt"
	"time"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain/errx"
	"github.com/woodleighschool/grinch/internal/domain/machines"
	"github.com/woodleighschool/grinch/internal/domain/policies"
	"github.com/woodleighschool/grinch/internal/store/db/pgconv"
)

// Preflight processes the Santa preflight stage and returns the sync parameters for the client.
func (s SyncService) Preflight(
	ctx context.Context,
	machineID uuid.UUID,
	req *syncv1.PreflightRequest,
) (*syncv1.PreflightResponse, error) {
	existing, err := s.machines.Get(ctx, machineID)
	if err != nil && !errx.IsCode(err, errx.CodeNotFound) {
		return nil, fmt.Errorf("preflight: get machine: %w", err)
	}

	now := time.Now().UTC()
	machine := buildMachineFromRequest(machineID, req, existing, now)

	if machine.PolicyID == nil {
		machine = clearAppliedState(machine)
		if _, err = s.machines.Upsert(ctx, machine); err != nil {
			return nil, fmt.Errorf("preflight: upsert machine: %w", err)
		}
		return &syncv1.PreflightResponse{}, nil
	}

	policy, err := s.policies.Get(ctx, *machine.PolicyID)
	if err != nil {
		return nil, fmt.Errorf("preflight: get policy: %w", err)
	}

	machine.AppliedSettingsVersion = &policy.SettingsVersion
	machine.PolicyStatus = policies.ComputePolicyStatus(machine, policy)

	resolved, err := s.machines.Upsert(ctx, machine)
	if err != nil {
		return nil, fmt.Errorf("preflight: upsert machine: %w", err)
	}

	resp := buildPreflightResponse(policy)
	resp.SyncType = determineSyncType(resolved, policy, req)
	return resp, nil
}

func buildMachineFromRequest(
	machineID uuid.UUID,
	req *syncv1.PreflightRequest,
	existing machines.Machine,
	now time.Time,
) machines.Machine {
	primaryUserPtr := pgconv.StrPtr(req.GetPrimaryUser())
	primaryUserGroups := req.GetPrimaryUserGroups()
	if primaryUserGroups == nil {
		primaryUserGroups = []string{}
	}

	return machines.Machine{
		ID:                   machineID,
		SerialNumber:         req.GetSerialNumber(),
		Hostname:             req.GetHostname(),
		Model:                req.GetModelIdentifier(),
		OSVersion:            req.GetOsVersion(),
		OSBuild:              req.GetOsBuild(),
		SantaVersion:         req.GetSantaVersion(),
		PrimaryUser:          primaryUserPtr,
		PrimaryUserGroups:    primaryUserGroups,
		PushToken:            pgconv.StrPtr(req.GetPushNotificationToken()),
		SIPStatus:            pgconv.Uint32ToInt32(req.GetSipStatus()),
		ClientMode:           req.GetClientMode(),
		RequestCleanSync:     req.GetRequestCleanSync(),
		PushNotificationSync: req.GetPushNotificationSync(),
		BinaryRuleCount:      pgconv.Uint32ToInt32(req.GetBinaryRuleCount()),
		CertificateRuleCount: pgconv.Uint32ToInt32(req.GetCertificateRuleCount()),
		CompilerRuleCount:    pgconv.Uint32ToInt32(req.GetCompilerRuleCount()),
		TransitiveRuleCount:  pgconv.Uint32ToInt32(req.GetTransitiveRuleCount()),
		TeamIDRuleCount:      pgconv.Uint32ToInt32(req.GetTeamidRuleCount()),
		SigningIDRuleCount:   pgconv.Uint32ToInt32(req.GetSigningidRuleCount()),
		CDHashRuleCount:      pgconv.Uint32ToInt32(req.GetCdhashRuleCount()),
		RulesHash:            pgconv.StrPtr(req.GetRulesHash()),
		LastSeen:             now,

		PolicyID:               existing.PolicyID,
		AppliedPolicyID:        existing.AppliedPolicyID,
		AppliedSettingsVersion: existing.AppliedSettingsVersion,
		AppliedRulesVersion:    existing.AppliedRulesVersion,
		UserID: preserveUserID(
			primaryUserPtr != nil && existing.PrimaryUser != nil && *primaryUserPtr == *existing.PrimaryUser,
			existing,
		),
	}
}

func preserveUserID(samePrimary bool, existing machines.Machine) *uuid.UUID {
	if samePrimary {
		return existing.UserID
	}
	return nil
}

func clearAppliedState(m machines.Machine) machines.Machine {
	m.AppliedPolicyID = nil
	m.AppliedSettingsVersion = nil
	m.AppliedRulesVersion = nil
	m.PolicyStatus = machines.PolicyStatusUnassigned
	return m
}

func determineSyncType(
	machine machines.Machine,
	policy policies.Policy,
	req *syncv1.PreflightRequest,
) *syncv1.SyncType {
	if req.GetRequestCleanSync() {
		return syncv1.SyncType_CLEAN.Enum()
	}

	if policy.ID == uuid.Nil {
		if machine.AppliedPolicyID != nil || req.GetRulesHash() != "" {
			return syncv1.SyncType_CLEAN_ALL.Enum()
		}
		return nil
	}

	if machine.AppliedPolicyID == nil || *machine.AppliedPolicyID != policy.ID {
		return syncv1.SyncType_CLEAN_ALL.Enum()
	}

	if policy.RulesVersion != 0 && machine.AppliedRulesVersion != nil &&
		policy.RulesVersion == *machine.AppliedRulesVersion {
		return nil
	}

	return syncv1.SyncType_CLEAN_ALL.Enum()
}

func buildPreflightResponse(p policies.Policy) *syncv1.PreflightResponse {
	resp := &syncv1.PreflightResponse{
		ClientMode:                              p.SetClientMode,
		BatchSize:                               pgconv.Int32ToUint32(p.SetBatchSize),
		FullSyncIntervalSeconds:                 pgconv.Int32ToUint32(p.SetFullSyncIntervalSeconds),
		PushNotificationFullSyncIntervalSeconds: pgconv.Int32ToUint32(p.SetPushNotificationFullSyncIntervalSeconds),
		PushNotificationGlobalRuleSyncDeadlineSeconds: pgconv.Int32ToUint32(
			p.SetPushNotificationGlobalRuleSyncDeadlineSeconds,
		),
		EnableBundles:             &p.SetEnableBundles,
		EnableTransitiveRules:     &p.SetEnableTransitiveRules,
		EnableAllEventUpload:      &p.SetEnableAllEventUpload,
		DisableUnknownEventUpload: &p.SetDisableUnknownEventUpload,
		AllowedPathRegex:          &p.SetAllowedPathRegex,
		BlockedPathRegex:          &p.SetBlockedPathRegex,
		BlockUsbMount:             &p.SetBlockUSBMount,
		OverrideFileAccessAction:  &p.SetOverrideFileAccessAction,
	}
	resp.RemountUsbMode = p.SetRemountUSBMode

	return resp
}
