package groups

import (
	"context"

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
