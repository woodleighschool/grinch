package policies

import (
	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"
)

// Policy describes a policy and its Santa settings, attachments, and targets.
type Policy struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"             validate:"required"`
	Description     string    `json:"description"`
	Enabled         bool      `json:"enabled"`
	Priority        int32     `json:"priority"         validate:"gte=0"`
	SettingsVersion int32     `json:"settings_version"`
	RulesVersion    int32     `json:"rules_version"`

	// Santa PreflightResponse fields.
	SetClientMode                                    syncv1.ClientMode       `json:"set_client_mode"                                         validate:"required"`
	SetBatchSize                                     int32                   `json:"set_batch_size"                                          validate:"gte=1"`
	SetEnableBundles                                 bool                    `json:"set_enable_bundles"`
	SetEnableTransitiveRules                         bool                    `json:"set_enable_transitive_rules"`
	SetEnableAllEventUpload                          bool                    `json:"set_enable_all_event_upload"`
	SetDisableUnknownEventUpload                     bool                    `json:"set_disable_unknown_event_upload"`
	SetFullSyncIntervalSeconds                       int32                   `json:"set_full_sync_interval_seconds"                          validate:"gte=60"`
	SetPushNotificationFullSyncIntervalSeconds       int32                   `json:"set_push_notification_full_sync_interval_seconds"        validate:"gte=60"`
	SetPushNotificationGlobalRuleSyncDeadlineSeconds int32                   `json:"set_push_notification_global_rule_sync_deadline_seconds" validate:"gte=0"`
	SetAllowedPathRegex                              string                  `json:"set_allowed_path_regex"`
	SetBlockedPathRegex                              string                  `json:"set_blocked_path_regex"`
	SetBlockUSBMount                                 bool                    `json:"set_block_usb_mount"`
	SetRemountUSBMode                                []string                `json:"set_remount_usb_mode"`
	SetOverrideFileAccessAction                      syncv1.FileAccessAction `json:"set_override_file_access_action"                         validate:"required"`

	Attachments []PolicyAttachment `json:"attachments"`
	Targets     []PolicyTarget     `json:"targets"`
}

// PolicyAttachment links a rule to a policy with an action and an optional CEL condition.
type PolicyAttachment struct {
	RuleID  uuid.UUID     `json:"rule_id"  validate:"required"`
	Action  syncv1.Policy `json:"action"`
	CELExpr *string       `json:"cel_expr"`
}

// PolicyTargetKind identifies the type of policy target.
type PolicyTargetKind string

const (
	TargetAll     PolicyTargetKind = "all"
	TargetUser    PolicyTargetKind = "user"
	TargetGroup   PolicyTargetKind = "group"
	TargetMachine PolicyTargetKind = "machine"
)

// PolicyTarget describes an entity a policy applies to.
type PolicyTarget struct {
	ID       uuid.UUID        `json:"id"`
	PolicyID uuid.UUID        `json:"policy_id"`
	Kind     PolicyTargetKind `json:"kind"      validate:"required,oneof=all user group machine"`
	RefID    *uuid.UUID       `json:"ref_id"`
}

// PolicyListItem summarises a policy for list endpoints.
type PolicyListItem struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled"`
	Priority    int32     `json:"priority"`
}
