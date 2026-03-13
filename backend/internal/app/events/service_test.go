package events_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/woodleighschool/grinch/internal/app/events"
)

type fakeStore struct {
	deletedEvents int64
	deleteCutoff  time.Time
	deleteErr     error
}

func (store *fakeStore) DeleteEventsBefore(_ context.Context, cutoff time.Time) (int64, error) {
	store.deleteCutoff = cutoff
	return store.deletedEvents, store.deleteErr
}

func TestCleanupExpiredEvents_DeletesEventsBeforeRetentionCutoff(t *testing.T) {
	store := &fakeStore{deletedEvents: 7}
	service := events.New(testLogger(), store, 30)

	start := time.Now().UTC()
	deleted, err := service.CleanupExpiredEvents(context.Background())
	end := time.Now().UTC()
	if err != nil {
		t.Fatalf("CleanupExpiredEvents() error = %v", err)
	}
	if deleted != 7 {
		t.Fatalf("CleanupExpiredEvents() deleted = %d, want 7", deleted)
	}

	lower := start.AddDate(0, 0, -30)
	upper := end.AddDate(0, 0, -30)
	if store.deleteCutoff.Before(lower) || store.deleteCutoff.After(upper) {
		t.Fatalf("cutoff out of expected range: got=%s lower=%s upper=%s", store.deleteCutoff, lower, upper)
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
