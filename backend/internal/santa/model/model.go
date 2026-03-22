// Package model defines the Santa sync state shared by the root santa service,
// the planning logic, and the Postgres adapter.
package model

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain"
)

type MachineUpsert struct {
	MachineID            uuid.UUID
	SerialNumber         string
	Hostname             string
	ModelIdentifier      string
	OSVersion            string
	OSBuild              string
	SantaVersion         string
	PrimaryUser          string
	PrimaryUserGroupsRaw []byte
	LastSeenAt           time.Time
}

type MachineSyncState struct {
	MachineID                   uuid.UUID
	RulesHash                   string
	DesiredTargets              []AppliedRuleTarget
	AppliedTargets              []AppliedRuleTarget
	PendingTargets              []AppliedRuleTarget
	PendingPayload              []SyncRule
	PendingPayloadRuleCount     int64
	PendingFullSync             bool
	PendingPreflightAt          *time.Time
	DesiredBinaryRuleCount      int32
	DesiredCertificateRuleCount int32
	DesiredTeamIDRuleCount      int32
	DesiredSigningIDRuleCount   int32
	DesiredCDHashRuleCount      int32
	ClientMode                  domain.MachineClientMode
	BinaryRuleCount             int32
	CertificateRuleCount        int32
	CompilerRuleCount           int32
	TransitiveRuleCount         int32
	TeamIDRuleCount             int32
	SigningIDRuleCount          int32
	CDHashRuleCount             int32
	RulesReceived               int32
	RulesProcessed              int32
	LastRuleSyncAttemptAt       *time.Time
	LastRuleSyncSuccessAt       *time.Time
	LastCleanSyncAt             *time.Time
	LastReportedCountsMatchAt   *time.Time
}

type (
	AppliedRuleTarget = domain.AppliedRuleTarget
	PendingRuleTarget = domain.PendingRuleTarget
)

type SyncRule struct {
	domain.MachineRuleTarget `json:",inline"`

	Removed bool `json:"removed"`
}

type PendingSnapshotWrite struct {
	MachineID                   uuid.UUID
	RulesHash                   string
	DesiredTargets              []AppliedRuleTarget
	AppliedTargets              []AppliedRuleTarget
	PendingTargets              []AppliedRuleTarget
	PendingPayload              []SyncRule
	PendingPayloadRuleCount     int64
	PendingFullSync             bool
	PendingPreflightAt          time.Time
	DesiredBinaryRuleCount      int32
	DesiredCertificateRuleCount int32
	DesiredTeamIDRuleCount      int32
	DesiredSigningIDRuleCount   int32
	DesiredCDHashRuleCount      int32
	ClientMode                  domain.MachineClientMode
	BinaryRuleCount             int32
	CertificateRuleCount        int32
	CompilerRuleCount           int32
	TransitiveRuleCount         int32
	TeamIDRuleCount             int32
	SigningIDRuleCount          int32
	CDHashRuleCount             int32
	RulesReceived               int32
	RulesProcessed              int32
	LastRuleSyncAttemptAt       *time.Time
	LastRuleSyncSuccessAt       *time.Time
	LastReportedCountsMatchAt   *time.Time
}

type PostflightWrite struct {
	MachineID             uuid.UUID
	RulesHash             string
	RulesReceived         int32
	RulesProcessed        int32
	LastRuleSyncAttemptAt time.Time
	LastRuleSyncSuccessAt *time.Time
}

// ExecutableWrite holds pre-decoded fields for an executable ingested during
// an event upload. SigningChain and Entitlements are already JSON-encoded.
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

// ProcessWrite holds pre-decoded fields for a single process in a file access
// event process chain. SigningChain is already JSON-encoded.
type ProcessWrite struct {
	Pid          int32
	FilePath     string
	FileSHA256   string
	SigningID    string
	TeamID       string
	CDHash       string
	SigningChain []byte
}

// ExecutionEventWrite is a single decoded execution event ready for storage.
type ExecutionEventWrite struct {
	Executable      ExecutableWrite
	FilePath        string
	ExecutingUser   string
	LoggedInUsers   []string
	CurrentSessions []string
	Decision        domain.EventDecision
	OccurredAt      *time.Time
}

// FileAccessEventWrite is a single decoded file access event ready for storage.
type FileAccessEventWrite struct {
	RuleVersion string
	RuleName    string
	Target      string
	Decision    domain.FileAccessDecision
	Processes   []ProcessWrite
	OccurredAt  *time.Time
}

// DataStore keeps the two-phase sync state:
// preflight writes a pending snapshot, and postflight promotes that snapshot
// once the client reports it processed the full pending payload.
type DataStore interface {
	UpsertMachine(context.Context, MachineUpsert) error
	SyncMachineDesiredRuleTargets(context.Context, uuid.UUID) error
	GetMachineSyncState(context.Context, uuid.UUID) (MachineSyncState, error)
	ReplacePendingSnapshot(context.Context, PendingSnapshotWrite) error
	RecordPostflight(context.Context, PostflightWrite) error
	PromotePendingSnapshot(context.Context, uuid.UUID, time.Time) error
	IngestEvents(
		context.Context,
		uuid.UUID,
		[]ExecutionEventWrite,
		[]FileAccessEventWrite,
	) error
}

type RuleResolver interface {
	ResolveMachineRuleTargets(context.Context, uuid.UUID) ([]domain.MachineResolvedRule, error)
}
