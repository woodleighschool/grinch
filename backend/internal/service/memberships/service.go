package memberships

import (
	"context"

	"github.com/google/uuid"

	corememberships "github.com/woodleighschool/grinch/internal/core/memberships"
)

// MembershipStore describes persistence for group memberships.
type MembershipStore interface {
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]corememberships.Membership, int64, error)
	ListByGroup(ctx context.Context, groupID uuid.UUID, limit, offset int) ([]corememberships.Membership, int64, error)
	GroupIDsForUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

// MembershipService exposes membership operations.
type MembershipService struct {
	store MembershipStore
}

// NewMembershipService constructs a MembershipService.
func NewMembershipService(store MembershipStore) *MembershipService {
	return &MembershipService{store: store}
}

// ListByUser returns memberships associated with the given user.
func (s *MembershipService) ListByUser(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
) ([]corememberships.Membership, int64, error) {
	return s.store.ListByUser(ctx, userID, limit, offset)
}

// ListByGroup returns memberships associated with the given group.
func (s *MembershipService) ListByGroup(
	ctx context.Context,
	groupID uuid.UUID,
	limit, offset int,
) ([]corememberships.Membership, int64, error) {
	return s.store.ListByGroup(ctx, groupID, limit, offset)
}

// GroupIDsForUser returns the IDs of all groups the user belongs to.
func (s *MembershipService) GroupIDsForUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	return s.store.GroupIDsForUser(ctx, userID)
}
