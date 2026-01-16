package users

import (
	"context"

	"github.com/google/uuid"

	coreusers "github.com/woodleighschool/grinch/internal/core/users"
	"github.com/woodleighschool/grinch/internal/listing"
)

// UserStore describes the persistence contract for users.
type UserStore interface {
	Get(ctx context.Context, id uuid.UUID) (coreusers.User, error)
	GetByUPN(ctx context.Context, upn string) (uuid.UUID, error)
	Upsert(ctx context.Context, user coreusers.User) error
	List(ctx context.Context, query listing.Query) ([]coreusers.User, listing.Page, error)
}

// UserService owns user level operations and validation.
type UserService struct {
	store UserStore
}

// NewUserService constructs a UserService.
func NewUserService(store UserStore) *UserService {
	return &UserService{store: store}
}

// Get resolves a user by ID.
func (s *UserService) Get(ctx context.Context, id uuid.UUID) (coreusers.User, error) {
	return s.store.Get(ctx, id)
}

// List returns users that match the provided listing query.
func (s *UserService) List(ctx context.Context, query listing.Query) ([]coreusers.User, listing.Page, error) {
	return s.store.List(ctx, query)
}

// Upsert validates and persists a user.
func (s *UserService) Upsert(ctx context.Context, user coreusers.User) error {
	return s.store.Upsert(ctx, user)
}

// ResolveIDByUPN returns a user ID for a UPN, or uuid.Nil if not known.
func (s *UserService) ResolveIDByUPN(ctx context.Context, upn string) (uuid.UUID, error) {
	return s.store.GetByUPN(ctx, upn)
}
