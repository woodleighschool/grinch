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

func (s *Service) ListFileAccessEvents(
	ctx context.Context,
	opts domain.FileAccessEventListOptions,
) ([]domain.FileAccessEventSummary, int32, error) {
	return s.store.ListFileAccessEvents(ctx, opts)
}

func (s *Service) GetFileAccessEvent(ctx context.Context, id uuid.UUID) (domain.FileAccessEvent, error) {
	return s.store.GetFileAccessEvent(ctx, id)
}

func (s *Service) DeleteFileAccessEvent(ctx context.Context, id uuid.UUID) error {
	return s.store.DeleteFileAccessEvent(ctx, id)
}
