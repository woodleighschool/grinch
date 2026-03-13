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

type RuleSyncType string

const (
	RuleSyncTypeNone       RuleSyncType = ""
	RuleSyncTypeNormal     RuleSyncType = "normal"
	RuleSyncTypeCleanRules RuleSyncType = "clean_rules"
)

type MachineRuleSyncState struct {
	MachineID                uuid.UUID
	RequestCleanSync         bool
	LastClientRulesHash      string
	AcknowledgedTargets      []StoredRuleTarget
	PendingTargets           []StoredRuleTarget
	PendingExpectedRulesHash string
	PendingPayloadRuleCount  int64
	PendingSyncType          RuleSyncType
	PendingPreflightAt       *time.Time
	LastPostflightAt         *time.Time
}

type StoredRuleTarget struct {
	domain.MachineRuleTarget

	PayloadHash string
}

type SyncRule struct {
	StoredRuleTarget

	Removed bool
}

type PendingSnapshotWrite struct {
	MachineID                uuid.UUID
	RequestCleanSync         bool
	LastClientRulesHash      string
	AcknowledgedTargets      []StoredRuleTarget
	PendingTargets           []StoredRuleTarget
	PendingExpectedRulesHash string
	PendingPayloadRuleCount  int64
	PendingSyncType          RuleSyncType
	PendingPreflightAt       time.Time
	LastPostflightAt         *time.Time
}

// DataStore keeps the two-phase sync state:
// preflight writes a pending snapshot, and postflight promotes that snapshot
// once the client reports the expected final rules hash and processed count.
type DataStore interface {
	UpsertMachine(context.Context, MachineUpsert) error
	GetMachineRuleSyncState(context.Context, uuid.UUID) (MachineRuleSyncState, error)
	ReplacePendingSnapshot(context.Context, PendingSnapshotWrite) error
	PromotePendingSnapshot(context.Context, uuid.UUID, string, time.Time) error
	IngestEvents(
		context.Context,
		uuid.UUID,
		[]*syncv1.Event,
		[]*syncv1.FileAccessEvent,
		map[domain.EventDecision]struct{},
	) (int, error)
	DeleteEventsBefore(context.Context, time.Time) (int64, error)
}

type RuleResolver interface {
	ResolveMachineRuleTargets(context.Context, uuid.UUID) ([]domain.MachineRuleTarget, error)
}
