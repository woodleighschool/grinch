package memberships

import (
	"context"

	"github.com/google/uuid"
)

// Repo defines persistence operations for group memberships.
type Repo interface {
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Membership, int64, error)
	ListByGroup(ctx context.Context, groupID uuid.UUID, limit, offset int) ([]Membership, int64, error)
	GroupIDsForUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

// Service exposes membership related domain operations.
type Service struct {
	repo Repo
}

// NewService constructs a Service backed by the provided repository.
func NewService(repo Repo) Service {
	return Service{repo: repo}
}

// ListByUser returns memberships associated with the given user.
func (s Service) ListByUser(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
) ([]Membership, int64, error) {
	return s.repo.ListByUser(ctx, userID, limit, offset)
}

// ListByGroup returns memberships associated with the given group.
func (s Service) ListByGroup(
	ctx context.Context,
	groupID uuid.UUID,
	limit, offset int,
) ([]Membership, int64, error) {
	return s.repo.ListByGroup(ctx, groupID, limit, offset)
}

// GroupIDsForUser returns the IDs of all groups the user belongs to.
func (s Service) GroupIDsForUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	return s.repo.GroupIDsForUser(ctx, userID)
}
