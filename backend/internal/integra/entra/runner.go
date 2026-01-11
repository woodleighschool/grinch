package entra

import (
	"context"
	"log/slog"
	"time"
)

// Runner executes periodic synchronisation against Entra.
type Runner struct {
	syncer   *Syncer
	log      *slog.Logger
	interval time.Duration
}

// NewRunner constructs a Runner with the given syncer and interval.
func NewRunner(syncer *Syncer, interval time.Duration, log *slog.Logger) *Runner {
	return &Runner{
		syncer:   syncer,
		log:      log.With("component", "entra_runner"),
		interval: interval,
	}
}

// Start runs the sync loop until the context is canceled.
func (r *Runner) Start(ctx context.Context) {
	r.log.InfoContext(ctx, "entra runner starting", "interval", r.interval)

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		start := time.Now()

		if err := r.syncer.Sync(ctx); err != nil {
			r.log.ErrorContext(
				ctx,
				"entra sync failed",
				"error", err,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		} else {
			r.log.InfoContext(
				ctx,
				"entra sync completed",
				"duration_ms", time.Since(start).Milliseconds(),
			)
		}

		select {
		case <-ctx.Done():
			r.log.InfoContext(ctx, "entra runner stopping")
			return
		case <-ticker.C:
		}
	}
}
