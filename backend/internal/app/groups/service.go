package groups

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain"
)

type WriteInput struct {
	Name        string
	Description string
}

type Store interface {
	ListGroups(context.Context, domain.GroupListOptions) ([]domain.Group, int32, error)
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

func (service *Service) ListGroups(
	ctx context.Context,
	options domain.GroupListOptions,
) ([]domain.Group, int32, error) {
	return service.store.ListGroups(ctx, options)
}

func (service *Service) GetGroup(ctx context.Context, id uuid.UUID) (domain.Group, error) {
	return service.store.GetGroup(ctx, id)
}

func (service *Service) CreateGroup(ctx context.Context, input WriteInput) (domain.Group, error) {
	normalized := normalizeInput(input)
	if validationErr := validateInput(normalized); validationErr != nil {
		return domain.Group{}, validationErr
	}

	return service.store.CreateLocalGroup(ctx, normalized.Name, normalized.Description)
}

func (service *Service) UpdateGroup(ctx context.Context, id uuid.UUID, input WriteInput) (domain.Group, error) {
	normalized := normalizeInput(input)
	if validationErr := validateInput(normalized); validationErr != nil {
		return domain.Group{}, validationErr
	}

	return service.store.UpdateGroup(ctx, id, normalized.Name, normalized.Description)
}

func (service *Service) DeleteGroup(ctx context.Context, id uuid.UUID) error {
	return service.store.DeleteGroup(ctx, id)
}

func normalizeInput(input WriteInput) WriteInput {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	return input
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
