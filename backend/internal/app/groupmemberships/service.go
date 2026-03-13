package groupmemberships

import (
	"context"

	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain"
)

type CreateInput struct {
	GroupID    uuid.UUID
	MemberKind domain.MemberKind
	MemberID   uuid.UUID
}

type Store interface {
	ListGroupMemberships(
		context.Context,
		domain.GroupMembershipListOptions,
	) ([]domain.GroupMembership, int32, error)
	GetGroupMembership(context.Context, uuid.UUID) (domain.GroupMembership, error)
	CreateGroupMembership(
		context.Context,
		uuid.UUID,
		domain.MemberKind,
		uuid.UUID,
		domain.GroupMembershipOrigin,
	) (domain.GroupMembership, error)
	DeleteGroupMembership(context.Context, uuid.UUID) error
	GetGroup(context.Context, uuid.UUID) (domain.Group, error)
}

type Service struct {
	store Store
}

func New(store Store) *Service {
	return &Service{store: store}
}

func (service *Service) ListGroupMemberships(
	ctx context.Context,
	options domain.GroupMembershipListOptions,
) ([]domain.GroupMembership, int32, error) {
	return service.store.ListGroupMemberships(ctx, options)
}

func (service *Service) GetGroupMembership(ctx context.Context, id uuid.UUID) (domain.GroupMembership, error) {
	return service.store.GetGroupMembership(ctx, id)
}

func (service *Service) CreateGroupMembership(ctx context.Context, input CreateInput) (domain.GroupMembership, error) {
	group, err := service.store.GetGroup(ctx, input.GroupID)
	if err != nil {
		return domain.GroupMembership{}, err
	}
	if group.Source == domain.PrincipalSourceEntra {
		return domain.GroupMembership{}, domain.ErrGroupReadOnly
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
		return domain.GroupMembership{}, validationErr
	}

	return service.store.CreateGroupMembership(
		ctx,
		input.GroupID,
		input.MemberKind,
		input.MemberID,
		domain.GroupMembershipOriginExplicit,
	)
}

func (service *Service) DeleteGroupMembership(ctx context.Context, id uuid.UUID) error {
	membership, err := service.store.GetGroupMembership(ctx, id)
	if err != nil {
		return err
	}

	if membership.Group.Source == domain.PrincipalSourceEntra {
		return domain.ErrGroupReadOnly
	}

	return service.store.DeleteGroupMembership(ctx, id)
}
