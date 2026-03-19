package santa

import "github.com/woodleighschool/grinch/internal/app/santa/model"

type (
	MachineUpsert        = model.MachineUpsert
	MachineSyncState     = model.MachineSyncState
	StoredRuleTarget     = model.StoredRuleTarget
	SyncRule             = model.SyncRule
	PendingSnapshotWrite = model.PendingSnapshotWrite
	PostflightWrite      = model.PostflightWrite
	DataStore            = model.DataStore
	RuleResolver         = model.RuleResolver
)
