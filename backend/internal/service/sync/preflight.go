package sync

import (
	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"

	coremachines "github.com/woodleighschool/grinch/internal/core/machines"
	"github.com/woodleighschool/grinch/internal/core/policies"
	"github.com/woodleighschool/grinch/internal/store/db/pgconv"
)

func buildPreflightResponse(policy policies.Policy) *syncv1.PreflightResponse {
	resp := &syncv1.PreflightResponse{
		ClientMode:              policy.SetClientMode,
		BatchSize:               pgconv.Int32ToUint32(policy.SetBatchSize),
		FullSyncIntervalSeconds: pgconv.Int32ToUint32(policy.SetFullSyncIntervalSeconds),
		PushNotificationFullSyncIntervalSeconds: pgconv.Int32ToUint32(
			policy.SetPushNotificationFullSyncIntervalSeconds,
		),
		PushNotificationGlobalRuleSyncDeadlineSeconds: pgconv.Int32ToUint32(
			policy.SetPushNotificationGlobalRuleSyncDeadlineSeconds,
		),
		EnableBundles:             &policy.SetEnableBundles,
		EnableTransitiveRules:     &policy.SetEnableTransitiveRules,
		EnableAllEventUpload:      &policy.SetEnableAllEventUpload,
		DisableUnknownEventUpload: &policy.SetDisableUnknownEventUpload,
		AllowedPathRegex:          &policy.SetAllowedPathRegex,
		BlockedPathRegex:          &policy.SetBlockedPathRegex,
		BlockUsbMount:             &policy.SetBlockUSBMount,
		OverrideFileAccessAction:  &policy.SetOverrideFileAccessAction,
	}
	resp.RemountUsbMode = policy.SetRemountUSBMode
	return resp
}

func determineSyncType(
	machine coremachines.Machine,
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
