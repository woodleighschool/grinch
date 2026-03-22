package fileaccessevents

import (
	"context"

	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain"
)

type Store interface {
	ListFileAccessEvents(
		context.Context,
		domain.FileAccessEventListOptions,
	) ([]domain.FileAccessEventSummary, int32, error)
	GetFileAccessEvent(context.Context, uuid.UUID) (domain.FileAccessEvent, error)
	DeleteFileAccessEvent(context.Context, uuid.UUID) error
}

type Service struct {
	store Store
}

func New(store Store) *Service {
	return &Service{store: store}
}

func (service *Service) ListFileAccessEvents(
	ctx context.Context,
	options domain.FileAccessEventListOptions,
) ([]domain.FileAccessEventSummary, int32, error) {
	return service.store.ListFileAccessEvents(ctx, options)
}

func (service *Service) GetFileAccessEvent(ctx context.Context, id uuid.UUID) (domain.FileAccessEvent, error) {
	return service.store.GetFileAccessEvent(ctx, id)
}

func (service *Service) DeleteFileAccessEvent(ctx context.Context, id uuid.UUID) error {
	return service.store.DeleteFileAccessEvent(ctx, id)
}
