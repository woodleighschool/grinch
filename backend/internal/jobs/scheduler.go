package jobs

import (
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

// RecurringConfig holds intervals and retention for recurring jobs.
type RecurringConfig struct {
	EntraInterval   time.Duration
	PruneInterval   time.Duration
	RetentionPeriod time.Duration
}

// PeriodicJobs defines the recurring River jobs with sensible uniqueness.
func PeriodicJobs(cfg RecurringConfig) []*river.PeriodicJob {
	return []*river.PeriodicJob{
		newPeriodicJob(
			"entra-sync",
			cfg.EntraInterval,
			func() (river.JobArgs, *river.InsertOpts) {
				return EntraSyncArgs{}, &river.InsertOpts{
					UniqueOpts:  activeUniqueOpts(),
					MaxAttempts: 1,
				}
			},
		),
		newPeriodicJob(
			"prune-events",
			cfg.PruneInterval,
			func() (river.JobArgs, *river.InsertOpts) {
				return PruneEventsArgs{Before: time.Now().Add(-cfg.RetentionPeriod)},
					&river.InsertOpts{
						UniqueOpts:  activeUniqueOpts(),
						MaxAttempts: 1,
					}
			},
		),
	}
}

// newPeriodicJob wraps common options for recurring jobs.
func newPeriodicJob(
	id string,
	interval time.Duration,
	constructor river.PeriodicJobConstructor,
) *river.PeriodicJob {
	return river.NewPeriodicJob(
		river.PeriodicInterval(interval),
		constructor,
		&river.PeriodicJobOpts{
			ID:         id,
			RunOnStart: true,
		},
	)
}

// activeUniqueOpts enforces one in-flight job per kind but allows new runs after completion.
func activeUniqueOpts() river.UniqueOpts {
	return river.UniqueOpts{
		ByState: []rivertype.JobState{
			rivertype.JobStatePending,
			rivertype.JobStateScheduled,
			rivertype.JobStateAvailable,
			rivertype.JobStateRunning,
			rivertype.JobStateRetryable,
		},
	}
}
