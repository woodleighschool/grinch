package santa

import "github.com/woodleighschool/grinch/internal/app/santa/model"

type (
	MachineUpsert        = model.MachineUpsert
	RuleSyncType         = model.RuleSyncType
	MachineRuleSyncState = model.MachineRuleSyncState
	StoredRuleTarget     = model.StoredRuleTarget
	SyncRule             = model.SyncRule
	PendingSnapshotWrite = model.PendingSnapshotWrite
	DataStore            = model.DataStore
	RuleResolver         = model.RuleResolver
)

const (
	RuleSyncTypeNone       = model.RuleSyncTypeNone
	RuleSyncTypeNormal     = model.RuleSyncTypeNormal
	RuleSyncTypeCleanRules = model.RuleSyncTypeCleanRules
)
