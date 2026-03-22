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
	ListMemberships(
		context.Context,
		domain.MembershipListOptions,
	) ([]domain.MembershipListItem, int32, error)
	GetMembership(context.Context, uuid.UUID) (domain.Membership, error)
	CreateMembership(
		context.Context,
		uuid.UUID,
		domain.MemberKind,
		uuid.UUID,
		domain.MembershipOrigin,
	) (domain.Membership, error)
	DeleteMembership(context.Context, uuid.UUID) error
	GetGroup(context.Context, uuid.UUID) (domain.Group, error)
	SyncMachineDesiredRuleTargets(context.Context, uuid.UUID) error
	SyncMachineDesiredRuleTargetsByPrimaryUserID(context.Context, uuid.UUID) error
}

type Service struct {
	store Store
}

func New(store Store) *Service {
	return &Service{store: store}
}

func (service *Service) ListMemberships(
	ctx context.Context,
	options domain.MembershipListOptions,
) ([]domain.MembershipListItem, int32, error) {
	return service.store.ListMemberships(ctx, options)
}

func (service *Service) GetMembership(ctx context.Context, id uuid.UUID) (domain.Membership, error) {
	return service.store.GetMembership(ctx, id)
}

func (service *Service) CreateMembership(ctx context.Context, input CreateInput) (domain.Membership, error) {
	group, err := service.store.GetGroup(ctx, input.GroupID)
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

	membership, err := service.store.CreateMembership(
		ctx,
		input.GroupID,
		input.MemberKind,
		input.MemberID,
		domain.MembershipOriginExplicit,
	)
	if err != nil {
		return domain.Membership{}, err
	}
	syncErr := syncMembershipMachineRuleTargets(ctx, service.store, input.MemberKind, input.MemberID)
	if syncErr != nil {
		return domain.Membership{}, syncErr
	}

	return membership, nil
}

func (service *Service) DeleteMembership(ctx context.Context, id uuid.UUID) error {
	membership, err := service.store.GetMembership(ctx, id)
	if err != nil {
		return err
	}

	if membership.Group.Source == domain.PrincipalSourceEntra {
		return domain.ErrGroupReadOnly
	}

	deleteErr := service.store.DeleteMembership(ctx, id)
	if deleteErr != nil {
		return deleteErr
	}

	return syncMembershipMachineRuleTargets(ctx, service.store, membership.Member.Kind, membership.Member.ID)
}

func syncMembershipMachineRuleTargets(
	ctx context.Context,
	store Store,
	memberKind domain.MemberKind,
	memberID uuid.UUID,
) error {
	switch memberKind {
	case domain.MemberKindMachine:
		return store.SyncMachineDesiredRuleTargets(ctx, memberID)
	case domain.MemberKindUser:
		return store.SyncMachineDesiredRuleTargetsByPrimaryUserID(ctx, memberID)
	default:
		return fmt.Errorf("unsupported member kind %q", memberKind)
	}
}
