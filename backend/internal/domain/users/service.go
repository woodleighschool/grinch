package users

import (
	"context"

	"github.com/google/uuid"
	"github.com/woodleighschool/grinch/internal/domain/errx"
	"github.com/woodleighschool/grinch/internal/listing"
)

// Repo defines persistence operations for users.
type Repo interface {
	Get(ctx context.Context, id uuid.UUID) (User, error)
	GetByUPN(ctx context.Context, upn string) (uuid.UUID, error)
	Upsert(ctx context.Context, user User) error
	List(ctx context.Context, query listing.Query) ([]User, listing.Page, error)
}

// Service provides user related domain operations.
type Service struct {
	repo Repo
}

// NewService constructs a Service backed by the provided repository.
func NewService(repo Repo) Service {
	return Service{repo: repo}
}

// Get returns the user identified by id.
func (s Service) Get(ctx context.Context, id uuid.UUID) (User, error) {
	return s.repo.Get(ctx, id)
}

// GetByUPN resolves a user ID from a UPN.
func (s Service) GetByUPN(ctx context.Context, upn string) (uuid.UUID, error) {
	return s.repo.GetByUPN(ctx, upn)
}

// Upsert validates and persists the user.
func (s Service) Upsert(ctx context.Context, user User) error {
	if err := errx.ValidateStruct(user); err != nil {
		return err
	}
	return s.repo.Upsert(ctx, user)
}

// List returns users matching the supplied query.
func (s Service) List(ctx context.Context, query listing.Query) ([]User, listing.Page, error) {
	return s.repo.List(ctx, query)
}
