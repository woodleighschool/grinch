package policies

import (
	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"

	"github.com/woodleighschool/grinch/internal/domain/policies"
	"github.com/woodleighschool/grinch/internal/store/db/pgconv"
	"github.com/woodleighschool/grinch/internal/store/db/sqlc"
)

func toDomainPolicy(row sqlc.Policy) policies.Policy {
	p := policies.Policy{
		ID:                           row.ID,
		Name:                         row.Name,
		Description:                  row.Description,
		Enabled:                      row.Enabled,
		Priority:                     row.Priority,
		SettingsVersion:              row.SettingsVersion,
		RulesVersion:                 row.RulesVersion,
		SetClientMode:                syncv1.ClientMode(row.SetClientMode),
		SetBatchSize:                 row.SetBatchSize,
		SetEnableBundles:             row.SetEnableBundles,
		SetEnableTransitiveRules:     row.SetEnableTransitiveRules,
		SetEnableAllEventUpload:      row.SetEnableAllEventUpload,
		SetDisableUnknownEventUpload: row.SetDisableUnknownEventUpload,
		SetFullSyncIntervalSeconds:   row.SetFullSyncIntervalSeconds,
		SetPushNotificationFullSyncIntervalSeconds:       row.SetPushNotificationFullSyncIntervalSeconds,
		SetPushNotificationGlobalRuleSyncDeadlineSeconds: row.SetPushNotificationGlobalRuleSyncDeadlineSeconds,
		SetAllowedPathRegex:                              row.SetAllowedPathRegex,
		SetBlockedPathRegex:                              row.SetBlockedPathRegex,
		SetBlockUSBMount:                                 row.SetBlockUsbMount,
		SetRemountUSBMode:                                pgconv.TextArray(row.SetRemountUsbMode),
		SetOverrideFileAccessAction:                      syncv1.FileAccessAction(row.SetOverrideFileAccessAction),
	}

	return p
}

func toDomainTargets(rows []sqlc.PolicyTarget) []policies.Target {
	out := make([]policies.Target, len(rows))
	for i, row := range rows {
		t := policies.Target{
			ID:       row.ID,
			PolicyID: row.PolicyID,
			Kind:     policies.TargetKind(row.Kind),
		}
		switch t.Kind {
		case policies.TargetUser:
			t.RefID = row.UserID
		case policies.TargetGroup:
			t.RefID = row.GroupID
		case policies.TargetMachine:
			t.RefID = row.MachineID
		case policies.TargetAll:
		}
		out[i] = t
	}
	return out
}

func toDomainAttachments(rows []sqlc.PolicyRule) []policies.Attachment {
	out := make([]policies.Attachment, len(rows))
	for i, row := range rows {
		out[i] = policies.Attachment{
			RuleID:  row.RuleID,
			Action:  syncv1.Policy(row.Action),
			CELExpr: pgconv.TextVal(row.CelExpr),
		}
	}
	return out
}

func toCreateParams(p policies.Policy) sqlc.CreatePolicyParams {
	return sqlc.CreatePolicyParams{
		Name:                         p.Name,
		Description:                  p.Description,
		Enabled:                      p.Enabled,
		Priority:                     p.Priority,
		SettingsVersion:              p.SettingsVersion,
		RulesVersion:                 p.RulesVersion,
		SetClientMode:                int32(p.SetClientMode),
		SetBatchSize:                 p.SetBatchSize,
		SetEnableBundles:             p.SetEnableBundles,
		SetEnableTransitiveRules:     p.SetEnableTransitiveRules,
		SetEnableAllEventUpload:      p.SetEnableAllEventUpload,
		SetDisableUnknownEventUpload: p.SetDisableUnknownEventUpload,
		SetFullSyncIntervalSeconds:   p.SetFullSyncIntervalSeconds,
		SetPushNotificationFullSyncIntervalSeconds:       p.SetPushNotificationFullSyncIntervalSeconds,
		SetPushNotificationGlobalRuleSyncDeadlineSeconds: p.SetPushNotificationGlobalRuleSyncDeadlineSeconds,
		SetAllowedPathRegex:                              p.SetAllowedPathRegex,
		SetBlockedPathRegex:                              p.SetBlockedPathRegex,
		SetBlockUsbMount:                                 p.SetBlockUSBMount,
		SetRemountUsbMode:                                pgconv.TextArray(p.SetRemountUSBMode),
		SetOverrideFileAccessAction:                      int32(p.SetOverrideFileAccessAction),
	}
}

func toUpdateParams(p policies.Policy) sqlc.UpdatePolicyByIDParams {
	return sqlc.UpdatePolicyByIDParams{
		ID:                           p.ID,
		Name:                         p.Name,
		Description:                  p.Description,
		Enabled:                      p.Enabled,
		Priority:                     p.Priority,
		SettingsVersion:              p.SettingsVersion,
		RulesVersion:                 p.RulesVersion,
		SetClientMode:                int32(p.SetClientMode),
		SetBatchSize:                 p.SetBatchSize,
		SetEnableBundles:             p.SetEnableBundles,
		SetEnableTransitiveRules:     p.SetEnableTransitiveRules,
		SetEnableAllEventUpload:      p.SetEnableAllEventUpload,
		SetDisableUnknownEventUpload: p.SetDisableUnknownEventUpload,
		SetFullSyncIntervalSeconds:   p.SetFullSyncIntervalSeconds,
		SetPushNotificationFullSyncIntervalSeconds:       p.SetPushNotificationFullSyncIntervalSeconds,
		SetPushNotificationGlobalRuleSyncDeadlineSeconds: p.SetPushNotificationGlobalRuleSyncDeadlineSeconds,
		SetAllowedPathRegex:                              p.SetAllowedPathRegex,
		SetBlockedPathRegex:                              p.SetBlockedPathRegex,
		SetBlockUsbMount:                                 p.SetBlockUSBMount,
		SetRemountUsbMode:                                pgconv.TextArray(p.SetRemountUSBMode),
		SetOverrideFileAccessAction:                      int32(p.SetOverrideFileAccessAction),
	}
}
