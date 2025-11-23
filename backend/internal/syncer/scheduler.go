package syncer

import (
	"context"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

type Job func(context.Context) error

type Scheduler struct {
	cron   *cron.Cron
	logger *slog.Logger
}

func NewScheduler(logger *slog.Logger) *Scheduler {
	return &Scheduler{
		cron:   cron.New(),
		logger: logger,
	}
}

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

func (s *Scheduler) Start() {
	s.cron.Start()
}

func (s *Scheduler) Stop() context.Context {
	return s.cron.Stop()
}
