package machines

import (
	"context"

	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/listing"
)

// UserLookup resolves user IDs from UPNs.
type UserLookup interface {
	GetByUPN(ctx context.Context, upn string) (uuid.UUID, error)
}

// Repo defines persistence operations for machines.
type Repo interface {
	Upsert(ctx context.Context, mc Machine) (Machine, error)
	Get(ctx context.Context, id uuid.UUID) (Machine, error)
	List(ctx context.Context, query listing.Query) ([]ListItem, listing.Page, error)
	UpdatePolicyState(ctx context.Context, id uuid.UUID, policyID *uuid.UUID, status PolicyStatus) error
}

// Service provides machine operations.
type Service struct {
	repo  Repo
	users UserLookup
}

// NewService constructs a Service.
func NewService(repo Repo, users UserLookup) Service {
	return Service{repo: repo, users: users}
}

// Get returns a machine by ID.
func (s Service) Get(ctx context.Context, id uuid.UUID) (Machine, error) {
	return s.repo.Get(ctx, id)
}

// List returns machines matching the query.
func (s Service) List(ctx context.Context, query listing.Query) ([]ListItem, listing.Page, error) {
	return s.repo.List(ctx, query)
}

// Upsert creates or updates a machine after normalising fields and resolving user references.
func (s Service) Upsert(ctx context.Context, mc Machine) (Machine, error) {
	mc = s.resolveUserID(ctx, mc)
	return s.repo.Upsert(ctx, mc)
}

// UpdatePolicyState updates policy assignment metadata for a machine.
func (s Service) UpdatePolicyState(ctx context.Context, id uuid.UUID, policyID *uuid.UUID, status PolicyStatus) error {
	return s.repo.UpdatePolicyState(ctx, id, policyID, status)
}

func (s Service) resolveUserID(ctx context.Context, mc Machine) Machine {
	if mc.PrimaryUser == nil || *mc.PrimaryUser == "" {
		mc.UserID = nil
		return mc
	}
	if mc.UserID != nil {
		return mc
	}
	if s.users == nil {
		return mc
	}

	userID, err := s.users.GetByUPN(ctx, *mc.PrimaryUser)
	if err != nil || userID == uuid.Nil {
		return mc
	}

	mc.UserID = &userID
	return mc
}
