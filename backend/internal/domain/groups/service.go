package groups

import (
	"context"

	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain/errx"
	"github.com/woodleighschool/grinch/internal/listing"
)

// PolicyRefresher refreshes derived policy state after membership changes.
type PolicyRefresher interface {
	RefreshAll(ctx context.Context) error
}

// Repo defines persistence operations for groups.
type Repo interface {
	Upsert(ctx context.Context, g Group) error
	ReplaceMemberships(ctx context.Context, groupID uuid.UUID, userIDs []uuid.UUID) error
	Get(ctx context.Context, id uuid.UUID) (Group, error)
	List(ctx context.Context, query listing.Query) ([]Group, listing.Page, error)
}

// Service provides group operations.
type Service struct {
	repo      Repo
	refresher PolicyRefresher
}

// NewService constructs a Service.
func NewService(repo Repo, refresher PolicyRefresher) Service {
	return Service{repo: repo, refresher: refresher}
}

// Get returns a group by ID.
func (s Service) Get(ctx context.Context, id uuid.UUID) (Group, error) {
	return s.repo.Get(ctx, id)
}

// List returns groups matching the query.
func (s Service) List(ctx context.Context, query listing.Query) ([]Group, listing.Page, error) {
	return s.repo.List(ctx, query)
}

// Upsert validates and persists a group.
func (s Service) Upsert(ctx context.Context, g Group) error {
	if err := errx.ValidateStruct(g); err != nil {
		return err
	}
	return s.repo.Upsert(ctx, g)
}

// ReplaceMemberships replaces a group's memberships and refreshes derived policy state.
func (s Service) ReplaceMemberships(ctx context.Context, groupID uuid.UUID, userIDs []uuid.UUID) error {
	if groupID == uuid.Nil {
		return &errx.Error{
			Code:    errx.CodeInvalid,
			Message: "Validation failed",
			Fields:  map[string]string{"group_id": "Group ID is required"},
		}
	}

	if err := s.repo.ReplaceMemberships(ctx, groupID, userIDs); err != nil {
		return err
	}

	return s.refreshPolicies(ctx)
}

func (s Service) refreshPolicies(ctx context.Context) error {
	if s.refresher == nil {
		return nil
	}
	return s.refresher.RefreshAll(ctx)
}
