package syncer

import (
	"context"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

// Job represents a single sync task (users, groups, rules, etc).
type Job func(context.Context) error

// Scheduler wraps robfig/cron with context-aware helpers.
type Scheduler struct {
	cron   *cron.Cron
	logger *slog.Logger
}

// NewScheduler creates a scheduler with the default cron configuration.
func NewScheduler(logger *slog.Logger) *Scheduler {
	return &Scheduler{
		cron:   cron.New(),
		logger: logger,
	}
}

// Add registers a cron entry that enforces a timeout per run.
func (s *Scheduler) Add(spec, name string, timeout time.Duration, job Job) error {
	_, err := s.cron.AddFunc(spec, func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := job(ctx); err != nil {
			s.logger.Error("sync job failed", "job", name, "err", err)
		} else {
			s.logger.Debug("sync job complete", "job", name)
		}
	})
	return err
}

// Start launches the cron scheduler.
func (s *Scheduler) Start() {
	s.cron.Start()
}

// Stop halts the scheduler and returns a context that closes once jobs finish.
func (s *Scheduler) Stop() context.Context {
	return s.cron.Stop()
}
