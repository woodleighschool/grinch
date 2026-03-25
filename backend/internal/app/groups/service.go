package groups

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain"
)

type WriteInput struct {
	Name        string
	Description string
}

type Store interface {
	ListGroups(context.Context, domain.ListOptions) ([]domain.Group, int32, error)
	GetGroup(context.Context, uuid.UUID) (domain.Group, error)
	CreateLocalGroup(context.Context, string, string) (domain.Group, error)
	UpdateGroup(context.Context, uuid.UUID, string, string) (domain.Group, error)
	DeleteGroup(context.Context, uuid.UUID) error
	ListMemberships(context.Context, domain.MembershipListOptions) ([]domain.Membership, int32, error)
	CreateMembership(
		context.Context,
		uuid.UUID,
		domain.MemberKind,
		uuid.UUID,
		domain.MembershipOrigin,
	) (domain.Membership, error)
	DeleteMembership(context.Context, uuid.UUID, domain.MemberKind) error
	UpdateMachineDesiredTargets(context.Context, uuid.UUID) error
	UpdateMachineDesiredTargetsByPrimaryUserID(context.Context, uuid.UUID) error
}

type Service struct {
	store Store
}

func New(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) ListGroups(ctx context.Context, opts domain.ListOptions) ([]domain.Group, int32, error) {
	return s.store.ListGroups(ctx, opts)
}

func (s *Service) GetGroup(ctx context.Context, id uuid.UUID) (domain.Group, error) {
	return s.store.GetGroup(ctx, id)
}

func (s *Service) CreateGroup(ctx context.Context, input WriteInput) (domain.Group, error) {
	if err := validateInput(input); err != nil {
		return domain.Group{}, err
	}

	return s.store.CreateLocalGroup(ctx, input.Name, input.Description)
}

func (s *Service) UpdateGroup(ctx context.Context, id uuid.UUID, input WriteInput) (domain.Group, error) {
	if err := validateInput(input); err != nil {
		return domain.Group{}, err
	}

	return s.store.UpdateGroup(ctx, id, input.Name, input.Description)
}

func (s *Service) DeleteGroup(ctx context.Context, id uuid.UUID) error {
	return s.store.DeleteGroup(ctx, id)
}

func (s *Service) AddUser(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) error {
	return s.addMember(ctx, groupID, domain.MemberKindUser, userID)
}

func (s *Service) RemoveUser(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) error {
	return s.removeMember(ctx, groupID, domain.MemberKindUser, userID)
}

func (s *Service) AddMachine(ctx context.Context, groupID uuid.UUID, machineID uuid.UUID) error {
	return s.addMember(ctx, groupID, domain.MemberKindMachine, machineID)
}

func (s *Service) RemoveMachine(ctx context.Context, groupID uuid.UUID, machineID uuid.UUID) error {
	return s.removeMember(ctx, groupID, domain.MemberKindMachine, machineID)
}

func validateInput(input WriteInput) *domain.ValidationError {
	err := &domain.ValidationError{
		Code:   "validation_error",
		Detail: "Group is invalid.",
	}

	if input.Name == "" {
		err.Add("name", "must not be empty", "required")
	}

	if !err.HasFieldErrors() {
		return nil
	}
	return err
}

func (s *Service) addMember(
	ctx context.Context,
	groupID uuid.UUID,
	memberKind domain.MemberKind,
	memberID uuid.UUID,
) error {
	if err := s.validateMutableGroup(ctx, groupID); err != nil {
		return err
	}

	opts, err := membershipLookupOptions(groupID, memberKind, memberID)
	if err != nil {
		return err
	}

	memberships, _, err := s.store.ListMemberships(ctx, opts)
	if err != nil {
		return err
	}
	if len(memberships) > 0 {
		return nil
	}

	if _, err = s.store.CreateMembership(
		ctx,
		groupID,
		memberKind,
		memberID,
		domain.MembershipOriginExplicit,
	); err != nil {
		return err
	}

	return syncMembershipMachineRuleTargets(ctx, s.store, memberKind, memberID)
}

func (s *Service) removeMember(
	ctx context.Context,
	groupID uuid.UUID,
	memberKind domain.MemberKind,
	memberID uuid.UUID,
) error {
	if err := s.validateMutableGroup(ctx, groupID); err != nil {
		return err
	}

	opts, err := membershipLookupOptions(groupID, memberKind, memberID)
	if err != nil {
		return err
	}

	memberships, _, err := s.store.ListMemberships(ctx, opts)
	if err != nil {
		return err
	}
	if len(memberships) == 0 {
		return nil
	}

	if err = s.store.DeleteMembership(ctx, memberships[0].ID, memberKind); err != nil {
		return err
	}

	return syncMembershipMachineRuleTargets(ctx, s.store, memberKind, memberID)
}

func (s *Service) validateMutableGroup(ctx context.Context, groupID uuid.UUID) error {
	group, err := s.store.GetGroup(ctx, groupID)
	if err != nil {
		return err
	}
	if group.Source == domain.PrincipalSourceEntra {
		return domain.ErrGroupReadOnly
	}

	return nil
}

func membershipLookupOptions(
	groupID uuid.UUID,
	memberKind domain.MemberKind,
	memberID uuid.UUID,
) (domain.MembershipListOptions, error) {
	opts := domain.MembershipListOptions{
		ListOptions: domain.ListOptions{},
		GroupID:     &groupID,
	}

	switch memberKind {
	case domain.MemberKindUser:
		opts.UserID = &memberID
	case domain.MemberKindMachine:
		opts.MachineID = &memberID
	default:
		return domain.MembershipListOptions{}, fmt.Errorf("unsupported member kind %q", memberKind)
	}

	return opts, nil
}

func syncMembershipMachineRuleTargets(
	ctx context.Context,
	store Store,
	memberKind domain.MemberKind,
	memberID uuid.UUID,
) error {
	switch memberKind {
	case domain.MemberKindMachine:
		return store.UpdateMachineDesiredTargets(ctx, memberID)
	case domain.MemberKindUser:
		return store.UpdateMachineDesiredTargetsByPrimaryUserID(ctx, memberID)
	default:
		return fmt.Errorf("unsupported member kind %q", memberKind)
	}
}
