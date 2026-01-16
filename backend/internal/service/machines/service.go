package machines

import (
	"context"

	"github.com/google/uuid"

	coremachines "github.com/woodleighschool/grinch/internal/core/machines"
	"github.com/woodleighschool/grinch/internal/core/policies"
	"github.com/woodleighschool/grinch/internal/listing"
)

// UserResolver resolves user IDs from UPNs.
type UserResolver interface {
	ResolveIDByUPN(ctx context.Context, upn string) (uuid.UUID, error)
}

// MachineStore defines persistence operations for machines.
type MachineStore interface {
	Upsert(ctx context.Context, mc coremachines.Machine) (coremachines.Machine, error)
	Get(ctx context.Context, id uuid.UUID) (coremachines.Machine, error)
	List(ctx context.Context, query listing.Query) ([]coremachines.MachineListItem, listing.Page, error)
	UpdatePolicyState(ctx context.Context, id uuid.UUID, policyID *uuid.UUID, status policies.Status) error
}

// MachineService provides machine operations.
type MachineService struct {
	store MachineStore
	users UserResolver
}

// NewMachineService constructs a MachineService.
func NewMachineService(store MachineStore, users UserResolver) *MachineService {
	return &MachineService{store: store, users: users}
}

// Get returns a machine by ID.
func (s *MachineService) Get(ctx context.Context, id uuid.UUID) (coremachines.Machine, error) {
	return s.store.Get(ctx, id)
}

// List returns machines matching the query.
func (s *MachineService) List(
	ctx context.Context,
	query listing.Query,
) ([]coremachines.MachineListItem, listing.Page, error) {
	return s.store.List(ctx, query)
}

// Upsert creates or updates a machine after normalising fields and resolving user references.
func (s *MachineService) Upsert(ctx context.Context, mc coremachines.Machine) (coremachines.Machine, error) {
	mc = s.resolveUserID(ctx, mc)
	return s.store.Upsert(ctx, mc)
}

// UpdatePolicyState updates policy assignment metadata for a machine.
func (s *MachineService) UpdatePolicyState(
	ctx context.Context,
	id uuid.UUID,
	policyID *uuid.UUID,
	status policies.Status,
) error {
	return s.store.UpdatePolicyState(ctx, id, policyID, status)
}

func (s *MachineService) resolveUserID(ctx context.Context, mc coremachines.Machine) coremachines.Machine {
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

	userID, err := s.users.ResolveIDByUPN(ctx, *mc.PrimaryUser)
	if err != nil || userID == uuid.Nil {
		return mc
	}

	mc.UserID = &userID
	return mc
}
