package jobs

import (
	"context"
	"errors"
	"time"

	"github.com/riverqueue/river"

	"github.com/woodleighschool/grinch/internal/integra/entra"
	"github.com/woodleighschool/grinch/internal/service/events"
	"github.com/woodleighschool/grinch/internal/service/policies"
)

// EntraSyncArgs triggers a directory sync.
type EntraSyncArgs struct{}

// Kind identifies the Entra sync job.
func (EntraSyncArgs) Kind() string { return "entra_sync" }

// EntraSyncWorker runs directory syncs.
type EntraSyncWorker struct {
	river.WorkerDefaults[EntraSyncArgs]

	Syncer *entra.Syncer
}

// Work executes a directory sync and schedules the next run.
func (w *EntraSyncWorker) Work(ctx context.Context, _ *river.Job[EntraSyncArgs]) error {
	if w.Syncer == nil {
		return errors.New("entra syncer not configured")
	}
	return w.Syncer.Sync(ctx)
}

// PolicyReconcileArgs triggers a full policy assignment refresh.
type PolicyReconcileArgs struct{}

// Kind identifies the policy reconcile job.
func (PolicyReconcileArgs) Kind() string { return "policy_reconcile" }

// PolicyReconcileWorker recalculates assignments.
type PolicyReconcileWorker struct {
	river.WorkerDefaults[PolicyReconcileArgs]

	Policies *policies.PolicyService
}

// Work recalculates machine policy assignments.
func (w *PolicyReconcileWorker) Work(ctx context.Context, _ *river.Job[PolicyReconcileArgs]) error {
	if w.Policies == nil {
		return errors.New("policy service not configured")
	}
	return w.Policies.RefreshAssignments(ctx)
}

// PruneEventsArgs deletes events older than Before.
type PruneEventsArgs struct {
	Before time.Time `json:"before"`
}

// Kind identifies the prune job.
func (PruneEventsArgs) Kind() string { return "prune_events" }

// PruneEventsWorker prunes old events.
type PruneEventsWorker struct {
	river.WorkerDefaults[PruneEventsArgs]

	Events    *events.EventService
	Retention time.Duration
}

// Work removes events older than the configured threshold.
func (w *PruneEventsWorker) Work(ctx context.Context, job *river.Job[PruneEventsArgs]) error {
	if w.Events == nil {
		return errors.New("event service not configured")
	}

	before := job.Args.Before
	if before.IsZero() {
		before = time.Now().Add(-w.Retention)
	}

	if _, err := w.Events.Prune(ctx, before); err != nil {
		return err
	}

	return nil
}
