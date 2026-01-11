package events

import (
	"context"

	"github.com/google/uuid"
	"github.com/woodleighschool/grinch/internal/listing"
)

// Repo defines persistence operations for event records.
type Repo interface {
	InsertBatch(ctx context.Context, items []Event) error
	Get(ctx context.Context, id uuid.UUID) (Event, error)
	List(ctx context.Context, query listing.Query) ([]ListItem, listing.Page, error)
}

// Service provides event related domain operations.
type Service struct {
	repo Repo
}

// NewService constructs a Service backed by the given repository.
func NewService(repo Repo) Service {
	return Service{repo: repo}
}

// Get returns the event identified by id.
func (s Service) Get(ctx context.Context, id uuid.UUID) (Event, error) {
	return s.repo.Get(ctx, id)
}

// List returns events matching the supplied listing query.
func (s Service) List(ctx context.Context, query listing.Query) ([]ListItem, listing.Page, error) {
	return s.repo.List(ctx, query)
}

// InsertBatch persists multiple events in a single operation.
func (s Service) InsertBatch(ctx context.Context, items []Event) error {
	if len(items) == 0 {
		return nil
	}
	return s.repo.InsertBatch(ctx, items)
}
