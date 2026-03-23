package entrasync

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	graphsync "github.com/woodleighschool/go-entrasync"

	"github.com/woodleighschool/grinch/internal/domain"
)

type GraphClient interface {
	Snapshot(context.Context) (*graphsync.Snapshot, error)
}

type DataStore interface {
	ReconcileSnapshot(context.Context, *graphsync.Snapshot) (domain.EntraSyncResult, error)
	UpdateAllMachineDesiredTargets(context.Context) error
}

type Service struct {
	logger   *slog.Logger
	client   GraphClient
	store    DataStore
	interval time.Duration
}

func New(logger *slog.Logger, client GraphClient, store DataStore, interval time.Duration) *Service {
	return &Service{
		logger:   logger,
		client:   client,
		store:    store,
		interval: interval,
	}
}

func (s *Service) Run(ctx context.Context) {
	s.syncAndLog(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.InfoContext(ctx, "entra sync stopped")
			return
		case <-ticker.C:
			s.syncAndLog(ctx)
		}
	}
}

func (s *Service) SyncOnce(ctx context.Context) (domain.EntraSyncResult, error) {
	snapshot, err := s.client.Snapshot(ctx)
	if err != nil {
		return domain.EntraSyncResult{}, fmt.Errorf("fetch snapshot: %w", err)
	}

	result, err := s.store.ReconcileSnapshot(ctx, snapshot)
	if err != nil {
		return domain.EntraSyncResult{}, fmt.Errorf("reconcile snapshot: %w", err)
	}

	if err = s.store.UpdateAllMachineDesiredTargets(ctx); err != nil {
		return domain.EntraSyncResult{}, fmt.Errorf("sync machine desired rule targets: %w", err)
	}

	return result, nil
}

func (s *Service) syncAndLog(ctx context.Context) {
	start := time.Now()

	result, err := s.SyncOnce(ctx)
	if err != nil {
		s.logger.ErrorContext(
			ctx,
			"entra sync failed",
			"error", err,
			"duration", time.Since(start),
		)
		return
	}

	s.logger.InfoContext(
		ctx,
		"entra sync complete",
		"users", result.Users,
		"groups", result.Groups,
		"memberships", result.Memberships,
		"duration", time.Since(start),
	)
}
