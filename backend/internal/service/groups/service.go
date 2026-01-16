package groups

import (
	"context"

	"github.com/google/uuid"

	coreerrors "github.com/woodleighschool/grinch/internal/core/errors"
	coregroups "github.com/woodleighschool/grinch/internal/core/groups"
	"github.com/woodleighschool/grinch/internal/listing"
)

// AssignmentRefresher recomputes derived policy state after membership changes.
type AssignmentRefresher interface {
	RefreshAssignments(ctx context.Context) error
}

// GroupStore describes the persistence contract for groups.
type GroupStore interface {
	Upsert(ctx context.Context, g coregroups.Group) error
	ReplaceMemberships(ctx context.Context, groupID uuid.UUID, userIDs []uuid.UUID) error
	Get(ctx context.Context, id uuid.UUID) (coregroups.Group, error)
	List(ctx context.Context, query listing.Query) ([]coregroups.Group, listing.Page, error)
}

// GroupService provides group operations and refreshes policy assignments when memberships change.
type GroupService struct {
	store     GroupStore
	refresher AssignmentRefresher
}

// NewGroupService constructs a GroupService.
func NewGroupService(store GroupStore, refresher AssignmentRefresher) *GroupService {
	return &GroupService{store: store, refresher: refresher}
}

// Get returns a group by ID.
func (s *GroupService) Get(ctx context.Context, id uuid.UUID) (coregroups.Group, error) {
	return s.store.Get(ctx, id)
}

// List returns groups matching the query.
func (s *GroupService) List(ctx context.Context, query listing.Query) ([]coregroups.Group, listing.Page, error) {
	return s.store.List(ctx, query)
}

// Upsert validates and persists a group.
func (s *GroupService) Upsert(ctx context.Context, g coregroups.Group) error {
	return s.store.Upsert(ctx, g)
}

// ReplaceMemberships replaces a group's memberships and refreshes derived policy state.
func (s *GroupService) ReplaceMemberships(ctx context.Context, groupID uuid.UUID, userIDs []uuid.UUID) error {
	if groupID == uuid.Nil {
		return &coreerrors.Error{
			Code:    coreerrors.CodeInvalid,
			Message: "Validation failed",
			Fields:  map[string]string{"group_id": "Group ID is required"},
		}
	}

	if err := s.store.ReplaceMemberships(ctx, groupID, userIDs); err != nil {
		return err
	}

	if s.refresher != nil {
		return s.refresher.RefreshAssignments(ctx)
	}
	return nil
}
