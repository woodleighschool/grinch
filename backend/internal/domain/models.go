package domain

import (
	"time"

	"github.com/google/uuid"
)

type MachineClientMode string

const (
	MachineClientModeUnknown    MachineClientMode = "unknown"
	MachineClientModeMonitor    MachineClientMode = "monitor"
	MachineClientModeLockdown   MachineClientMode = "lockdown"
	MachineClientModeStandalone MachineClientMode = "standalone"
)

type MachineRuleSyncStatus string

const (
	MachineRuleSyncStatusPending MachineRuleSyncStatus = "pending"
	MachineRuleSyncStatusSynced  MachineRuleSyncStatus = "synced"
	MachineRuleSyncStatusIssue   MachineRuleSyncStatus = "issue"
)

type MemberKind string

const (
	MemberKindMachine MemberKind = "machine"
	MemberKindUser    MemberKind = "user"
)

type MembershipOrigin string

const (
	MembershipOriginExplicit MembershipOrigin = "explicit"
	MembershipOriginSynced   MembershipOrigin = "synced"
)

type PrincipalSource string

const (
	PrincipalSourceEntra PrincipalSource = "entra"
	PrincipalSourceLocal PrincipalSource = "local"
)

type RulePolicy string

const (
	RulePolicyAllowlist       RulePolicy = "allowlist"
	RulePolicyBlocklist       RulePolicy = "blocklist"
	RulePolicyCEL             RulePolicy = "cel"
	RulePolicySilentBlocklist RulePolicy = "silent_blocklist"
)

type RuleTargetAssignment string

const (
	RuleTargetAssignmentExclude RuleTargetAssignment = "exclude"
	RuleTargetAssignmentInclude RuleTargetAssignment = "include"
)

type RuleTargetSubjectKind string

const (
	RuleTargetSubjectKindAllDevices RuleTargetSubjectKind = "all_devices"
	RuleTargetSubjectKindAllUsers   RuleTargetSubjectKind = "all_users"
	RuleTargetSubjectKindGroup      RuleTargetSubjectKind = "group"
)

type RuleType string

const (
	RuleTypeBinary      RuleType = "binary"
	RuleTypeCDHash      RuleType = "cd_hash"
	RuleTypeCertificate RuleType = "certificate"
	RuleTypeSigningID   RuleType = "signing_id"
	RuleTypeTeamID      RuleType = "team_id"
)

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
	TeamIDRuleCount      int32                 `json:"teamid_rule_count"`
	SigningIDRuleCount   int32                 `json:"signingid_rule_count"`
	CDHashRuleCount      int32                 `json:"cdhash_rule_count"`
	GroupIDs             []uuid.UUID           `json:"group_ids"`
	Rules                []MachineRule         `json:"rules"`
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

type MachineResolvedRule struct {
	MachineRuleTarget

	RuleID uuid.UUID
	Name   string
}

type MachineRule struct {
	RuleID  *uuid.UUID `json:"rule_id,omitempty"`
	Policy  RulePolicy `json:"policy"`
	Applied bool       `json:"applied"`
}

type RuleMachine struct {
	MachineID uuid.UUID  `json:"machine_id"`
	Policy    RulePolicy `json:"policy"`
	Applied   bool       `json:"applied"`
}

type Executable struct {
	ID             uuid.UUID           `json:"id"`
	FileSHA256     string              `json:"file_sha256"`
	FileName       string              `json:"file_name"`
	FileBundleID   string              `json:"file_bundle_id"`
	FileBundlePath string              `json:"file_bundle_path"`
	SigningID      string              `json:"signing_id"`
	TeamID         string              `json:"team_id"`
	CDHash         string              `json:"cdhash"`
	Occurrences    int32               `json:"occurrences"`
	Entitlements   map[string]any      `json:"entitlements"`
	SigningChain   []SigningChainEntry `json:"signing_chain"`
	CreatedAt      time.Time           `json:"created_at"`
}

type ExecutableSummary struct {
	ID             uuid.UUID `json:"id"`
	FileSHA256     string    `json:"file_sha256"`
	FileName       string    `json:"file_name"`
	FileBundleID   string    `json:"file_bundle_id"`
	FileBundlePath string    `json:"file_bundle_path"`
	SigningID      string    `json:"signing_id"`
	TeamID         string    `json:"team_id"`
	CDHash         string    `json:"cdhash"`
	Occurrences    int32     `json:"occurrences"`
	CreatedAt      time.Time `json:"created_at"`
}

type SigningChainEntry struct {
	CommonName         string    `json:"common_name"`
	Organization       string    `json:"organization"`
	OrganizationalUnit string    `json:"organizational_unit"`
	SHA256             string    `json:"sha256"`
	ValidFrom          time.Time `json:"valid_from"`
	ValidUntil         time.Time `json:"valid_until"`
}

type ExecutionEvent struct {
	ID              uuid.UUID           `json:"id"`
	MachineID       uuid.UUID           `json:"machine_id"`
	ExecutableID    uuid.UUID           `json:"executable_id"`
	Decision        ExecutionDecision   `json:"decision"`
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
	ID           uuid.UUID         `json:"id"`
	MachineID    uuid.UUID         `json:"machine_id"`
	ExecutableID uuid.UUID         `json:"executable_id"`
	Decision     ExecutionDecision `json:"decision"`
	FilePath     string            `json:"file_path"`
	FileName     string            `json:"file_name"`
	SigningID    string            `json:"signing_id"`
	OccurredAt   *time.Time        `json:"occurred_at,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
}

type FileAccessEventProcess struct {
	Pid        int32  `json:"pid"`
	FilePath   string `json:"file_path"`
	FileName   string `json:"file_name"`
	FileSHA256 string `json:"file_sha256"`
	SigningID  string `json:"signing_id"`
	TeamID     string `json:"team_id"`
	CDHash     string `json:"cdhash"`
}

type FileAccessEvent struct {
	ID           uuid.UUID                `json:"id"`
	MachineID    uuid.UUID                `json:"machine_id"`
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
	ID         uuid.UUID          `json:"id"`
	MachineID  uuid.UUID          `json:"machine_id"`
	Decision   FileAccessDecision `json:"decision"`
	RuleName   string             `json:"rule_name"`
	Target     string             `json:"target"`
	FileName   string             `json:"file_name"`
	FileSHA256 string             `json:"file_sha256"`
	SigningID  string             `json:"signing_id"`
	TeamID     string             `json:"team_id"`
	CDHash     string             `json:"cdhash"`
	OccurredAt *time.Time         `json:"occurred_at,omitempty"`
	CreatedAt  time.Time          `json:"created_at"`
}

type Rule struct {
	ID            uuid.UUID     `json:"id"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	RuleType      RuleType      `json:"rule_type"`
	Identifier    string        `json:"identifier"`
	CustomMessage string        `json:"custom_message"`
	CustomURL     string        `json:"custom_url"`
	Enabled       bool          `json:"enabled"`
	Targets       RuleTargets   `json:"targets"`
	Machines      []RuleMachine `json:"machines"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
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
	RuleType      RuleType   `json:"rule_type"`
	Identifier    string     `json:"identifier"`
	Policy        RulePolicy `json:"policy"`
	CustomMessage string     `json:"custom_message"`
	CustomURL     string     `json:"custom_url"`
	CELExpression string     `json:"cel_expression"`
}

type User struct {
	ID          uuid.UUID       `json:"id"`
	UPN         string          `json:"upn"`
	DisplayName string          `json:"display_name"`
	Source      PrincipalSource `json:"source"`
	GroupIDs    []uuid.UUID     `json:"group_ids,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type Group struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Source      PrincipalSource `json:"source"`
	MemberCount int32           `json:"member_count"`
	UserIDs     []uuid.UUID     `json:"user_ids,omitempty"`
	MachineIDs  []uuid.UUID     `json:"machine_ids,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type MembershipGroup struct {
	ID     uuid.UUID       `json:"id"`
	Name   string          `json:"name"`
	Source PrincipalSource `json:"source"`
}

type MembershipMember struct {
	Kind MemberKind `json:"kind"`
	ID   uuid.UUID  `json:"id"`
	Name string     `json:"name,omitempty"`
}

type Membership struct {
	ID        uuid.UUID        `json:"id"`
	Group     MembershipGroup  `json:"group"`
	Member    MembershipMember `json:"member"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

type RuleWriteInput struct {
	Name          string
	Description   string
	RuleType      RuleType
	Identifier    string
	CustomMessage string
	CustomURL     string
	Enabled       bool
	Targets       RuleTargetsWriteInput
}

type RuleTargetsWriteInput struct {
	Include []IncludeRuleTargetWriteInput
	Exclude []ExcludedGroupWriteInput
}

type IncludeRuleTargetWriteInput struct {
	SubjectKind   RuleTargetSubjectKind
	SubjectID     *uuid.UUID
	Policy        RulePolicy
	CELExpression string
}

type ExcludedGroupWriteInput struct {
	GroupID uuid.UUID
}

type EntraSyncResult struct {
	Users       int
	Groups      int
	Memberships int
}
