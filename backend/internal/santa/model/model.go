// Package model defines the Santa sync types shared by the service, planning
// logic, and storage adapters.
package model

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain"
)

// MachineUpsert contains the latest machine identity and inventory details
// reported by a client.
type MachineUpsert struct {
	MachineID         uuid.UUID
	SerialNumber      string
	Hostname          string
	ModelIdentifier   string
	OSVersion         string
	OSBuild           string
	SantaVersion      string
	PrimaryUser       string
	PrimaryUserGroups []string
	ClientMode        domain.MachineClientMode
	LastSeenAt        time.Time
}

// MachineSyncState is the persisted two-phase sync state for a machine.
type MachineSyncState struct {
	MachineID uuid.UUID
	RulesHash string

	DesiredTargets []AppliedRuleTarget
	AppliedTargets []AppliedRuleTarget
	SentTargets    []AppliedRuleTarget
	PendingPayload []SyncRule

	PendingPayloadRuleCount int64
	PendingFullSync         bool
	PendingPreflightAt      *time.Time

	DesiredBinaryRuleCount      int32
	DesiredCertificateRuleCount int32
	DesiredTeamIDRuleCount      int32
	DesiredSigningIDRuleCount   int32
	DesiredCDHashRuleCount      int32

	BinaryRuleCount      int32
	CertificateRuleCount int32
	TeamIDRuleCount      int32
	SigningIDRuleCount   int32
	CDHashRuleCount      int32

	RulesReceived  int32
	RulesProcessed int32

	LastRuleSyncAttemptAt     *time.Time
	LastRuleSyncSuccessAt     *time.Time
	LastCleanSyncAt           *time.Time
	LastReportedCountsMatchAt *time.Time
}

// AppliedRuleTarget is the rule fingerprint stored once a target has been
// planned, sent, or acknowledged.
type AppliedRuleTarget struct {
	RuleType    domain.RuleType `json:"rule_type"`
	Identifier  string          `json:"identifier"`
	PayloadHash string          `json:"payload_hash"`
}

// PendingRuleTarget is a resolved machine rule target with its payload hash,
// ready for sync planning.
type PendingRuleTarget struct {
	domain.MachineRuleTarget

	PayloadHash string `json:"payload_hash"`
}

// SyncRule is a single rule mutation prepared for a sync payload.
type SyncRule struct {
	domain.MachineRuleTarget `json:",inline"`

	Removed bool `json:"removed"`
}

// PendingSnapshotWrite is the frozen preflight state written before rule
// download begins.
type PendingSnapshotWrite struct {
	MachineID uuid.UUID
	RulesHash string

	DesiredTargets []AppliedRuleTarget
	AppliedTargets []AppliedRuleTarget
	SentTargets    []AppliedRuleTarget
	PendingPayload []SyncRule

	PendingPayloadRuleCount int64
	PendingFullSync         bool
	PendingPreflightAt      time.Time

	DesiredBinaryRuleCount      int32
	DesiredCertificateRuleCount int32
	DesiredTeamIDRuleCount      int32
	DesiredSigningIDRuleCount   int32
	DesiredCDHashRuleCount      int32

	BinaryRuleCount      int32
	CertificateRuleCount int32
	TeamIDRuleCount      int32
	SigningIDRuleCount   int32
	CDHashRuleCount      int32

	RulesReceived  int32
	RulesProcessed int32

	LastRuleSyncAttemptAt     *time.Time
	LastRuleSyncSuccessAt     *time.Time
	LastReportedCountsMatchAt *time.Time
}

// PostflightWrite contains the client-reported sync result for a completed
// snapshot.
type PostflightWrite struct {
	MachineID uuid.UUID
	RulesHash string

	RulesReceived  int32
	RulesProcessed int32

	LastRuleSyncAttemptAt time.Time
	LastRuleSyncSuccessAt *time.Time
}

// ExecutableWrite contains a decoded executable ready for storage. Entitlements
// and SigningChain are already JSON-encoded.
type ExecutableWrite struct {
	FileSHA256     string
	FileName       string
	FileBundleID   string
	FileBundlePath string
	SigningID      string
	TeamID         string
	CDHash         string
	Entitlements   []byte
	SigningChain   []byte
}

// ProcessWrite contains a decoded process entry from a file access event chain.
// SigningChain is already JSON-encoded.
type ProcessWrite struct {
	Pid          int32
	FilePath     string
	FileSHA256   string
	SigningID    string
	TeamID       string
	CDHash       string
	SigningChain []byte
}

// ExecutionEventWrite is a decoded execution event ready for storage.
type ExecutionEventWrite struct {
	Executable      ExecutableWrite
	FilePath        string
	ExecutingUser   string
	LoggedInUsers   []string
	CurrentSessions []string
	Decision        domain.ExecutionDecision
	OccurredAt      *time.Time
}

// FileAccessEventWrite is a decoded file access event ready for storage.
type FileAccessEventWrite struct {
	RuleVersion string
	RuleName    string
	Target      string
	Decision    domain.FileAccessDecision
	Processes   []ProcessWrite
	OccurredAt  *time.Time
}

// DataStore stores two-phase sync state and ingested Santa events.
type DataStore interface {
	UpsertMachine(context.Context, MachineUpsert) error
	UpdateMachineDesiredTargets(context.Context, uuid.UUID) error
	GetMachineSyncState(context.Context, uuid.UUID) (MachineSyncState, error)
	ReplacePendingSnapshot(context.Context, PendingSnapshotWrite) error
	RecordPostflight(context.Context, PostflightWrite) error
	PromotePendingSnapshot(context.Context, uuid.UUID, time.Time) error
	IngestEvents(context.Context, uuid.UUID, []ExecutionEventWrite, []FileAccessEventWrite) error
}

// RuleResolver resolves the desired machine rule targets used during snapshot
// planning.
type RuleResolver interface {
	ResolveMachineRuleTargets(context.Context, uuid.UUID) ([]domain.MachineResolvedRule, error)
}
