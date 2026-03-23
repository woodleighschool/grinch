package memberships

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain"
)

type CreateInput struct {
	GroupID    uuid.UUID
	MemberKind domain.MemberKind
	MemberID   uuid.UUID
}

type Store interface {
	ListMemberships(context.Context, domain.MembershipListOptions) ([]domain.Membership, int32, error)
	GetMembership(context.Context, uuid.UUID) (domain.Membership, error)
	CreateMembership(
		context.Context,
		uuid.UUID,
		domain.MemberKind,
		uuid.UUID,
		domain.MembershipOrigin,
	) (domain.Membership, error)
	DeleteMembership(context.Context, uuid.UUID, domain.MemberKind) error
	GetGroup(context.Context, uuid.UUID) (domain.Group, error)
	UpdateMachineDesiredTargets(context.Context, uuid.UUID) error
	UpdateMachineDesiredTargetsByPrimaryUserID(context.Context, uuid.UUID) error
}

type Service struct {
	store Store
}

func New(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) ListMemberships(
	ctx context.Context,
	opts domain.MembershipListOptions,
) ([]domain.Membership, int32, error) {
	return s.store.ListMemberships(ctx, opts)
}

func (s *Service) GetMembership(ctx context.Context, id uuid.UUID) (domain.Membership, error) {
	return s.store.GetMembership(ctx, id)
}

func (s *Service) CreateMembership(ctx context.Context, input CreateInput) (domain.Membership, error) {
	group, err := s.store.GetGroup(ctx, input.GroupID)
	if err != nil {
		return domain.Membership{}, err
	}
	if group.Source == domain.PrincipalSourceEntra {
		return domain.Membership{}, domain.ErrGroupReadOnly
	}

	validationErr := &domain.ValidationError{
		Code:   "validation_error",
		Detail: "Group membership is invalid.",
	}

	switch input.MemberKind {
	case domain.MemberKindUser, domain.MemberKindMachine:
	default:
		validationErr.Add("member_kind", "must be user or machine", "invalid")
	}

	if validationErr.HasFieldErrors() {
		return domain.Membership{}, validationErr
	}

	membership, err := s.store.CreateMembership(
		ctx,
		input.GroupID,
		input.MemberKind,
		input.MemberID,
		domain.MembershipOriginExplicit,
	)
	if err != nil {
		return domain.Membership{}, err
	}

	if err = syncMembershipMachineRuleTargets(ctx, s.store, input.MemberKind, input.MemberID); err != nil {
		return domain.Membership{}, err
	}

	return membership, nil
}

func (s *Service) DeleteMembership(ctx context.Context, id uuid.UUID) error {
	membership, err := s.store.GetMembership(ctx, id)
	if err != nil {
		return err
	}

	if membership.Group.Source == domain.PrincipalSourceEntra {
		return domain.ErrGroupReadOnly
	}

	if err = s.store.DeleteMembership(ctx, id, membership.Member.Kind); err != nil {
		return err
	}

	return syncMembershipMachineRuleTargets(ctx, s.store, membership.Member.Kind, membership.Member.ID)
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
