package domain

import (
	"time"

	"github.com/google/uuid"
)

type PrincipalSource string

const (
	PrincipalSourceLocal PrincipalSource = "local"
	PrincipalSourceEntra PrincipalSource = "entra"
)

type MemberKind string

const (
	MemberKindUser    MemberKind = "user"
	MemberKindMachine MemberKind = "machine"
)

type GroupMembershipOrigin string

const (
	GroupMembershipOriginExplicit GroupMembershipOrigin = "explicit"
	GroupMembershipOriginSynced   GroupMembershipOrigin = "synced"
)

type RuleType string

const (
	RuleTypeBinary      RuleType = "binary"
	RuleTypeCertificate RuleType = "certificate"
	RuleTypeTeamID      RuleType = "team_id"
	RuleTypeSigningID   RuleType = "signing_id"
	RuleTypeCDHash      RuleType = "cd_hash"
)

type RulePolicy string

const (
	RulePolicyAllowlist       RulePolicy = "allowlist"
	RulePolicyBlocklist       RulePolicy = "blocklist"
	RulePolicySilentBlocklist RulePolicy = "silent_blocklist"
	RulePolicyCEL             RulePolicy = "cel"
)

type RuleTargetAssignment string

const (
	RuleTargetAssignmentInclude RuleTargetAssignment = "include"
	RuleTargetAssignmentExclude RuleTargetAssignment = "exclude"
)

type RuleTargetSubjectKind string

const (
	RuleTargetSubjectKindGroup      RuleTargetSubjectKind = "group"
	RuleTargetSubjectKindAllDevices RuleTargetSubjectKind = "all_devices"
	RuleTargetSubjectKindAllUsers   RuleTargetSubjectKind = "all_users"
)

type MachineRuleSyncStatus string

const (
	MachineRuleSyncStatusSynced  MachineRuleSyncStatus = "synced"
	MachineRuleSyncStatusPending MachineRuleSyncStatus = "pending"
	MachineRuleSyncStatusIssue   MachineRuleSyncStatus = "issue"
)

type MachineClientMode string

const (
	MachineClientModeUnknown    MachineClientMode = "unknown"
	MachineClientModeMonitor    MachineClientMode = "monitor"
	MachineClientModeLockdown   MachineClientMode = "lockdown"
	MachineClientModeStandalone MachineClientMode = "standalone"
)

func ParseMachineClientMode(value string) MachineClientMode {
	switch MachineClientMode(value) {
	case MachineClientModeMonitor:
		return MachineClientModeMonitor
	case MachineClientModeLockdown:
		return MachineClientModeLockdown
	case MachineClientModeStandalone:
		return MachineClientModeStandalone
	default:
		return MachineClientModeUnknown
	}
}

func DeriveMachineRuleSyncStatus(
	pendingPreflightAt *time.Time,
	lastRuleSyncAttemptAt *time.Time,
) MachineRuleSyncStatus {
	if pendingPreflightAt == nil {
		return MachineRuleSyncStatusSynced
	}
	if lastRuleSyncAttemptAt == nil || lastRuleSyncAttemptAt.Before(*pendingPreflightAt) {
		return MachineRuleSyncStatusPending
	}
	return MachineRuleSyncStatusIssue
}

type Machine struct {
	ID                   uuid.UUID             `json:"id"`
	SerialNumber         string                `json:"serial_number"`
	Hostname             string                `json:"hostname"`
	ModelIdentifier      string                `json:"model_identifier"`
	OSVersion            string                `json:"os_version"`
	OSBuild              string                `json:"os_build"`
	SantaVersion         string                `json:"santa_version"`
	PrimaryUser          string                `json:"primary_user"`
	PrimaryUserID        *uuid.UUID            `json:"primary_user_id,omitempty"`
	RuleSyncStatus       MachineRuleSyncStatus `json:"rule_sync_status"`
	ClientMode           MachineClientMode     `json:"client_mode"`
	BinaryRuleCount      int32                 `json:"binary_rule_count"`
	CertificateRuleCount int32                 `json:"certificate_rule_count"`
	CompilerRuleCount    int32                 `json:"compiler_rule_count"`
	TransitiveRuleCount  int32                 `json:"transitive_rule_count"`
	TeamIDRuleCount      int32                 `json:"teamid_rule_count"`
	SigningIDRuleCount   int32                 `json:"signingid_rule_count"`
	CDHashRuleCount      int32                 `json:"cdhash_rule_count"`
	LastSeenAt           time.Time             `json:"last_seen_at"`
	CreatedAt            time.Time             `json:"created_at"`
	UpdatedAt            time.Time             `json:"updated_at"`
}

type MachineSummary struct {
	ID              uuid.UUID             `json:"id"`
	SerialNumber    string                `json:"serial_number"`
	Hostname        string                `json:"hostname"`
	ModelIdentifier string                `json:"model_identifier"`
	OSVersion       string                `json:"os_version"`
	SantaVersion    string                `json:"santa_version"`
	PrimaryUser     string                `json:"primary_user"`
	PrimaryUserID   *uuid.UUID            `json:"primary_user_id,omitempty"`
	RuleSyncStatus  MachineRuleSyncStatus `json:"rule_sync_status"`
	LastSeenAt      time.Time             `json:"last_seen_at"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
}

type ExecutableSource string

const (
	ExecutableSourceEvent   ExecutableSource = "event"
	ExecutableSourceProcess ExecutableSource = "process"
)

type Executable struct {
	ID             uuid.UUID           `json:"id"`
	Source         ExecutableSource    `json:"source"`
	FileSHA256     string              `json:"file_sha256"`
	FileName       string              `json:"file_name"`
	FilePath       string              `json:"file_path"`
	FileBundleID   string              `json:"file_bundle_id"`
	FileBundlePath string              `json:"file_bundle_path"`
	SigningID      string              `json:"signing_id"`
	TeamID         string              `json:"team_id"`
	CDHash         string              `json:"cdhash"`
	Entitlements   map[string]any      `json:"entitlements"`
	SigningChain   []SigningChainEntry `json:"signing_chain"`
	CreatedAt      time.Time           `json:"created_at"`
}

type SigningChainEntry struct {
	CommonName         string    `json:"common_name"`
	Organization       string    `json:"organization"`
	OrganizationalUnit string    `json:"organizational_unit"`
	SHA256             string    `json:"sha256"`
	ValidFrom          time.Time `json:"valid_from"`
	ValidUntil         time.Time `json:"valid_until"`
}

type ExecutableSummary struct {
	ID             uuid.UUID        `json:"id"`
	Source         ExecutableSource `json:"source"`
	FileSHA256     string           `json:"file_sha256"`
	FileName       string           `json:"file_name"`
	FilePath       string           `json:"file_path"`
	FileBundleID   string           `json:"file_bundle_id"`
	FileBundlePath string           `json:"file_bundle_path"`
	SigningID      string           `json:"signing_id"`
	TeamID         string           `json:"team_id"`
	CDHash         string           `json:"cdhash"`
	CreatedAt      time.Time        `json:"created_at"`
}

type ExecutionEvent struct {
	ID              uuid.UUID           `json:"id"`
	MachineID       uuid.UUID           `json:"machine_id"`
	ExecutableID    uuid.UUID           `json:"executable_id"`
	Decision        EventDecision       `json:"decision"`
	FilePath        string              `json:"file_path"`
	FileName        string              `json:"file_name"`
	FileSHA256      string              `json:"file_sha256"`
	FileBundleID    string              `json:"file_bundle_id"`
	FileBundlePath  string              `json:"file_bundle_path"`
	SigningID       string              `json:"signing_id"`
	TeamID          string              `json:"team_id"`
	CDHash          string              `json:"cdhash"`
	ExecutingUser   string              `json:"executing_user"`
	LoggedInUsers   []string            `json:"logged_in_users"`
	CurrentSessions []string            `json:"current_sessions"`
	SigningChain    []SigningChainEntry `json:"signing_chain"`
	Entitlements    map[string]any      `json:"entitlements"`
	OccurredAt      *time.Time          `json:"occurred_at,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
}

type ExecutionEventSummary struct {
	ID           uuid.UUID     `json:"id"`
	MachineID    uuid.UUID     `json:"machine_id"`
	ExecutableID uuid.UUID     `json:"executable_id"`
	Decision     EventDecision `json:"decision"`
	FilePath     string        `json:"file_path"`
	FileName     string        `json:"file_name"`
	SigningID    string        `json:"signing_id"`
	OccurredAt   *time.Time    `json:"occurred_at,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
}

type FileAccessEventProcess struct {
	Pid          int32     `json:"pid"`
	FilePath     string    `json:"file_path"`
	ExecutableID uuid.UUID `json:"executable_id"`
	FileName     string    `json:"file_name"`
}

type FileAccessEvent struct {
	ID           uuid.UUID                `json:"id"`
	MachineID    uuid.UUID                `json:"machine_id"`
	ExecutableID *uuid.UUID               `json:"executable_id,omitempty"`
	RuleVersion  string                   `json:"rule_version"`
	RuleName     string                   `json:"rule_name"`
	Target       string                   `json:"target"`
	Decision     FileAccessDecision       `json:"decision"`
	FileName     string                   `json:"file_name"`
	FileSHA256   string                   `json:"file_sha256"`
	SigningID    string                   `json:"signing_id"`
	TeamID       string                   `json:"team_id"`
	CDHash       string                   `json:"cdhash"`
	ProcessChain []FileAccessEventProcess `json:"process_chain"`
	OccurredAt   *time.Time               `json:"occurred_at,omitempty"`
	CreatedAt    time.Time                `json:"created_at"`
}

type FileAccessEventSummary struct {
	ID           uuid.UUID          `json:"id"`
	MachineID    uuid.UUID          `json:"machine_id"`
	ExecutableID *uuid.UUID         `json:"executable_id,omitempty"`
	Decision     FileAccessDecision `json:"decision"`
	RuleName     string             `json:"rule_name"`
	Target       string             `json:"target"`
	FileName     string             `json:"file_name"`
	FileSHA256   string             `json:"file_sha256"`
	SigningID    string             `json:"signing_id"`
	TeamID       string             `json:"team_id"`
	CDHash       string             `json:"cdhash"`
	OccurredAt   *time.Time         `json:"occurred_at,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
}

type Rule struct {
	ID            uuid.UUID   `json:"id"`
	Name          string      `json:"name"`
	Description   string      `json:"description"`
	RuleType      RuleType    `json:"rule_type"`
	Identifier    string      `json:"identifier"`
	CustomMessage string      `json:"custom_message"`
	CustomURL     string      `json:"custom_url"`
	Enabled       bool        `json:"enabled"`
	Targets       RuleTargets `json:"targets"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

type RuleSummary struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	RuleType    RuleType  `json:"rule_type"`
	Identifier  string    `json:"identifier"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type GroupTarget struct {
	SubjectID  *uuid.UUID
	Assignment RuleTargetAssignment
	Priority   *int32
}

type RuleTargets struct {
	Include []IncludeRuleTarget `json:"include"`
	Exclude []ExcludedGroup     `json:"exclude"`
}

type IncludeRuleTarget struct {
	SubjectKind   RuleTargetSubjectKind `json:"subject_kind"`
	SubjectID     *uuid.UUID            `json:"subject_id,omitempty"`
	SubjectName   string                `json:"subject_name,omitempty"`
	Policy        RulePolicy            `json:"policy"`
	CELExpression string                `json:"cel_expression,omitempty"`
}

type ExcludedGroup struct {
	GroupID   uuid.UUID `json:"group_id"`
	GroupName string    `json:"group_name,omitempty"`
}

type MachineRuleTarget struct {
	RuleType      RuleType
	Identifier    string
	IdentifierKey string
	Policy        RulePolicy
	CustomMessage string
	CustomURL     string
	CELExpression string
}

type User struct {
	ID          uuid.UUID       `json:"id"`
	UPN         string          `json:"upn"`
	DisplayName string          `json:"display_name"`
	Source      PrincipalSource `json:"source"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type Group struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Source      PrincipalSource `json:"source"`
	MemberCount int32           `json:"member_count"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type GroupMembershipKind string

const (
	GroupMembershipKindActual    GroupMembershipKind = "actual"
	GroupMembershipKindEffective GroupMembershipKind = "effective"
)

type GroupMembershipGroup struct {
	ID     uuid.UUID       `json:"id"`
	Name   string          `json:"name"`
	Source PrincipalSource `json:"source"`
}

type GroupMembershipMember struct {
	Kind MemberKind `json:"kind"`
	ID   uuid.UUID  `json:"id"`
	Name string     `json:"name,omitempty"`
}

type GroupMembership struct {
	ID        string                `json:"id"`
	Kind      GroupMembershipKind   `json:"kind"`
	Group     GroupMembershipGroup  `json:"group"`
	Member    GroupMembershipMember `json:"member"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
}

type MachineResolvedRule struct {
	MachineRuleTarget

	RuleID uuid.UUID
	Name   string
}

type MachineRule struct {
	ID        string     `json:"id"`
	MachineID uuid.UUID  `json:"machine_id"`
	RuleID    *uuid.UUID `json:"rule_id,omitempty"`
	Policy    RulePolicy `json:"policy"`
	Applied   bool       `json:"applied"`
}

type RuleMachine struct {
	ID        string     `json:"id"`
	RuleID    uuid.UUID  `json:"rule_id"`
	MachineID uuid.UUID  `json:"machine_id"`
	Policy    RulePolicy `json:"policy"`
	Applied   bool       `json:"applied"`
}
