package events_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/woodleighschool/grinch/internal/app/events"
)

type testStore struct {
	deletedEvents int64
	deleteCutoff  time.Time
	deleteErr     error
}

func (s *testStore) DeleteEventsBefore(_ context.Context, cutoff time.Time) (int64, error) {
	s.deleteCutoff = cutoff
	return s.deletedEvents, s.deleteErr
}

func newTestService(store *testStore) *events.Service {
	return events.New(newTestLogger(), store, 30)
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestCleanupExpiredEvents_DeletesEventsBeforeRetentionCutoff(t *testing.T) {
	store := &testStore{deletedEvents: 7}
	service := newTestService(store)

	before := time.Now().UTC()
	deleted, err := service.CleanupExpiredEvents(context.Background())
	after := time.Now().UTC()
	if err != nil {
		t.Fatalf("CleanupExpiredEvents() error = %v", err)
	}

	if deleted != 7 {
		t.Fatalf("deleted = %d, want 7", deleted)
	}

	lower := before.AddDate(0, 0, -30)
	upper := after.AddDate(0, 0, -30)
	if store.deleteCutoff.Before(lower) || store.deleteCutoff.After(upper) {
		t.Fatalf("DeleteCutoff out of expected range: %s", store.deleteCutoff)
	}
}
