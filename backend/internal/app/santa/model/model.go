// Package model defines the Santa sync state shared by the root santa service,
// the planning logic, and the Postgres adapter.
package model

import (
	"context"
	"time"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
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
	MachineID               uuid.UUID
	RulesHash               string
	AppliedTargets          []StoredRuleTarget
	PendingTargets          []StoredRuleTarget
	ExpectedRulesHash       string
	PendingPayloadRuleCount int64
	PendingFullSync         bool
	PendingPreflightAt      *time.Time
	ClientMode              domain.MachineClientMode
	BinaryRuleCount         int32
	CertificateRuleCount    int32
	CompilerRuleCount       int32
	TransitiveRuleCount     int32
	TeamIDRuleCount         int32
	SigningIDRuleCount      int32
	CDHashRuleCount         int32
	RulesReceived           int32
	RulesProcessed          int32
	LastRuleSyncAttemptAt   *time.Time
	LastRuleSyncSuccessAt   *time.Time
}

type StoredRuleTarget = domain.StoredRuleTarget

type SyncRule struct {
	StoredRuleTarget

	Removed bool
}

type PendingSnapshotWrite struct {
	MachineID               uuid.UUID
	RulesHash               string
	AppliedTargets          []StoredRuleTarget
	PendingTargets          []StoredRuleTarget
	ExpectedRulesHash       string
	PendingPayloadRuleCount int64
	PendingFullSync         bool
	PendingPreflightAt      time.Time
	ClientMode              domain.MachineClientMode
	BinaryRuleCount         int32
	CertificateRuleCount    int32
	CompilerRuleCount       int32
	TransitiveRuleCount     int32
	TeamIDRuleCount         int32
	SigningIDRuleCount      int32
	CDHashRuleCount         int32
	RulesReceived           int32
	RulesProcessed          int32
	LastRuleSyncAttemptAt   *time.Time
	LastRuleSyncSuccessAt   *time.Time
}

type PostflightWrite struct {
	MachineID             uuid.UUID
	RulesHash             string
	RulesReceived         int32
	RulesProcessed        int32
	LastRuleSyncAttemptAt time.Time
	LastRuleSyncSuccessAt *time.Time
}

// DataStore keeps the two-phase sync state:
// preflight writes a pending snapshot, and postflight promotes that snapshot
// once the client reports the expected final rules hash and processed count.
type DataStore interface {
	UpsertMachine(context.Context, MachineUpsert) error
	GetMachineSyncState(context.Context, uuid.UUID) (MachineSyncState, error)
	ReplacePendingSnapshot(context.Context, PendingSnapshotWrite) error
	RecordPostflight(context.Context, PostflightWrite) error
	PromotePendingSnapshot(context.Context, uuid.UUID, time.Time) error
	IngestEvents(
		context.Context,
		uuid.UUID,
		[]*syncv1.Event,
		[]*syncv1.FileAccessEvent,
		map[domain.EventDecision]struct{},
	) (int, error)
}

type RuleResolver interface {
	ResolveMachineRuleTargets(context.Context, uuid.UUID) ([]domain.MachineResolvedRule, error)
}
