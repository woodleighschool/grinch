package events

import (
	"context"
	"time"

	"github.com/google/uuid"

	coreevents "github.com/woodleighschool/grinch/internal/core/events"
	"github.com/woodleighschool/grinch/internal/listing"
)

// EventStore defines persistence operations for event records.
type EventStore interface {
	InsertBatch(ctx context.Context, items []coreevents.Event) error
	Get(ctx context.Context, id uuid.UUID) (coreevents.Event, error)
	List(ctx context.Context, query listing.Query) ([]coreevents.EventListItem, listing.Page, error)
	PruneBefore(ctx context.Context, before time.Time) (int64, error)
}

// EventService provides event related operations.
type EventService struct {
	store EventStore
}

// NewEventService constructs an EventService.
func NewEventService(store EventStore) *EventService {
	return &EventService{store: store}
}

// Get returns the event identified by id.
func (s *EventService) Get(ctx context.Context, id uuid.UUID) (coreevents.Event, error) {
	return s.store.Get(ctx, id)
}

// List returns events matching the supplied listing query.
func (s *EventService) List(
	ctx context.Context,
	query listing.Query,
) ([]coreevents.EventListItem, listing.Page, error) {
	return s.store.List(ctx, query)
}

// InsertBatch persists multiple events in a single operation.
func (s *EventService) InsertBatch(ctx context.Context, items []coreevents.Event) error {
	if len(items) == 0 {
		return nil
	}
	return s.store.InsertBatch(ctx, items)
}

// Prune removes events older than the given timestamp.
func (s *EventService) Prune(ctx context.Context, before time.Time) (int64, error) {
	return s.store.PruneBefore(ctx, before)
}
