// Package events owns event-lifecycle app concerns that are not part of the
// Santa sync protocol itself, such as retention cleanup.
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
	logger        *slog.Logger
	retentionDays int
	store         Store
}

func New(logger *slog.Logger, store Store, retentionDays int) *Service {
	return &Service{
		logger:        logger,
		retentionDays: retentionDays,
		store:         store,
	}
}

func (service *Service) CleanupExpiredEvents(ctx context.Context) (int64, error) {
	if service.retentionDays <= 0 {
		return 0, nil
	}

	cutoff := time.Now().UTC().AddDate(0, 0, -service.retentionDays)
	deleted, err := service.store.DeleteEventsBefore(ctx, cutoff)
	if err != nil {
		return 0, fmt.Errorf("delete events before %s: %w", cutoff.Format(time.RFC3339), err)
	}

	return deleted, nil
}

func (service *Service) RunRetention(ctx context.Context, interval time.Duration) {
	service.cleanupAndLog(ctx)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			service.logger.InfoContext(ctx, "event retention worker stopped")
			return
		case <-ticker.C:
			service.cleanupAndLog(ctx)
		}
	}
}

func (service *Service) cleanupAndLog(ctx context.Context) {
	started := time.Now()

	deleted, err := service.CleanupExpiredEvents(ctx)
	if err != nil {
		service.logger.ErrorContext(
			ctx,
			"event retention cleanup failed",
			"error",
			err,
			"duration",
			time.Since(started),
		)
		return
	}

	service.logger.InfoContext(
		ctx,
		"event retention cleanup complete",
		"retention_days",
		service.retentionDays,
		"deleted",
		deleted,
		"duration",
		time.Since(started),
	)
}
