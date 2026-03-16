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
	RuleTargetSubjectKindGroup RuleTargetSubjectKind = "group"
)

type Machine struct {
	ID               uuid.UUID
	SerialNumber     string
	Hostname         string
	ModelIdentifier  string
	OSVersion        string
	OSBuild          string
	SantaVersion     string
	PrimaryUser      string
	PrimaryUserID    *uuid.UUID
	RequestCleanSync bool
	LastSeenAt       time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type MachineSummary struct {
	ID              uuid.UUID
	SerialNumber    string
	Hostname        string
	ModelIdentifier string
	OSVersion       string
	SantaVersion    string
	PrimaryUser     string
	PrimaryUserID   *uuid.UUID
	LastSeenAt      time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Entitlement struct {
	Value any
}

type ExecutableSource string

const (
	ExecutableSourceEvent   ExecutableSource = "event"
	ExecutableSourceProcess ExecutableSource = "process"
)

type Executable struct {
	ID             uuid.UUID
	Source         ExecutableSource
	FileSHA256     string
	FileName       string
	FilePath       string
	FileBundleID   string
	FileBundlePath string
	SigningID      string
	TeamID         string
	CDHash         string
	Entitlements   map[string]Entitlement
	SigningChain   []SigningChainEntry
	CreatedAt      time.Time
}

type SigningChainEntry struct {
	CommonName         string
	Organization       string
	OrganizationalUnit string
	SHA256             string
	ValidFrom          time.Time
	ValidUntil         time.Time
}

type ExecutableSummary struct {
	ID             uuid.UUID
	Source         ExecutableSource
	FileSHA256     string
	FileName       string
	FilePath       string
	FileBundleID   string
	FileBundlePath string
	SigningID      string
	TeamID         string
	CDHash         string
	CreatedAt      time.Time
}

type ExecutionEvent struct {
	ID              uuid.UUID
	MachineID       uuid.UUID
	ExecutableID    uuid.UUID
	Decision        EventDecision
	FilePath        string
	FileName        string
	FileSHA256      string
	FileBundleID    string
	FileBundlePath  string
	SigningID       string
	TeamID          string
	CDHash          string
	ExecutingUser   string
	LoggedInUsers   []string
	CurrentSessions []string
	SigningChain    []SigningChainEntry
	Entitlements    map[string]Entitlement
	OccurredAt      *time.Time
	CreatedAt       time.Time
}

type ExecutionEventSummary struct {
	ID           uuid.UUID
	MachineID    uuid.UUID
	ExecutableID uuid.UUID
	Decision     EventDecision
	FilePath     string
	FileName     string
	SigningID    string
	OccurredAt   *time.Time
	CreatedAt    time.Time
}

type FileAccessEventProcess struct {
	Pid          int32
	FilePath     string
	ExecutableID uuid.UUID
	FileName     string
}

type FileAccessEvent struct {
	ID           uuid.UUID
	MachineID    uuid.UUID
	ExecutableID *uuid.UUID
	RuleVersion  string
	RuleName     string
	Target       string
	Decision     FileAccessDecision
	FileName     string
	FileSHA256   string
	SigningID    string
	TeamID       string
	CDHash       string
	ProcessChain []FileAccessEventProcess
	OccurredAt   *time.Time
	CreatedAt    time.Time
}

type FileAccessEventSummary struct {
	ID           uuid.UUID
	MachineID    uuid.UUID
	ExecutableID *uuid.UUID
	Decision     FileAccessDecision
	RuleName     string
	Target       string
	FileName     string
	FileSHA256   string
	SigningID    string
	TeamID       string
	CDHash       string
	OccurredAt   *time.Time
	CreatedAt    time.Time
}

type Rule struct {
	ID            uuid.UUID
	Name          string
	Description   string
	RuleType      RuleType
	Identifier    string
	CustomMessage string
	CustomURL     string
	Enabled       bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type RuleSummary struct {
	ID          uuid.UUID
	Name        string
	Description string
	RuleType    RuleType
	Identifier  string
	Enabled     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type GroupTarget struct {
	SubjectID  uuid.UUID
	Assignment RuleTargetAssignment
	Priority   *int32
}

type RuleTarget struct {
	ID            uuid.UUID
	RuleID        uuid.UUID
	SubjectKind   RuleTargetSubjectKind
	SubjectID     uuid.UUID
	Assignment    RuleTargetAssignment
	Priority      *int32
	Policy        *RulePolicy
	CELExpression string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type RuleTargetSummary struct {
	ID          uuid.UUID
	RuleID      uuid.UUID
	SubjectKind RuleTargetSubjectKind
	SubjectID   uuid.UUID
	Assignment  RuleTargetAssignment
	Priority    *int32
	Policy      *RulePolicy
	CreatedAt   time.Time
	UpdatedAt   time.Time
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
	ID          uuid.UUID
	UPN         string
	DisplayName string
	Source      PrincipalSource
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Group struct {
	ID          uuid.UUID
	Name        string
	Description string
	Source      PrincipalSource
	MemberCount int32
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type GroupMembershipKind string

const (
	GroupMembershipKindActual    GroupMembershipKind = "actual"
	GroupMembershipKindEffective GroupMembershipKind = "effective"
)

type GroupMembershipGroup struct {
	ID     uuid.UUID
	Name   string
	Source PrincipalSource
}

type GroupMembershipMember struct {
	Kind MemberKind
	ID   uuid.UUID
	Name string
}

type GroupMembership struct {
	ID        string
	Kind      GroupMembershipKind
	Group     GroupMembershipGroup
	Member    GroupMembershipMember
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MachineResolvedRule struct {
	MachineRuleTarget

	RuleID uuid.UUID
}
