// Package events owns event-lifecycle concerns outside the Santa sync protocol,
// such as retention cleanup.
package events

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type Store interface {
	DeleteEventsBefore(context.Context, time.Time) (int64, error)
}

type Service struct {
	logger *slog.Logger
	store  Store

	retentionDays int
}

func New(logger *slog.Logger, store Store, retentionDays int) *Service {
	return &Service{
		logger:        logger,
		store:         store,
		retentionDays: retentionDays,
	}
}

func (s *Service) CleanupExpiredEvents(ctx context.Context) (int64, error) {
	cutoff := time.Now().UTC().AddDate(0, 0, -s.retentionDays)

	deleted, err := s.store.DeleteEventsBefore(ctx, cutoff)
	if err != nil {
		return 0, fmt.Errorf("delete events before %s: %w", cutoff.Format(time.RFC3339), err)
	}

	return deleted, nil
}

func (s *Service) RunRetention(ctx context.Context, interval time.Duration) {
	s.runCleanup(ctx)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.InfoContext(ctx, "event retention worker stopped")
			return
		case <-ticker.C:
			s.runCleanup(ctx)
		}
	}
}

func (s *Service) runCleanup(ctx context.Context) {
	start := time.Now()

	deleted, err := s.CleanupExpiredEvents(ctx)
	if err != nil {
		s.logger.ErrorContext(
			ctx,
			"event retention cleanup failed",
			"error", err,
			"duration", time.Since(start),
		)
		return
	}

	s.logger.InfoContext(
		ctx,
		"event retention cleanup complete",
		"retention_days", s.retentionDays,
		"deleted", deleted,
		"duration", time.Since(start),
	)
}
