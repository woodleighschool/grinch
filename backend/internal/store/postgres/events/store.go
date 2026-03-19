package events

import (
	"context"
	"time"

	"github.com/woodleighschool/grinch/internal/store/postgres"
)

type Store struct {
	store *postgres.Store
}

func New(store *postgres.Store) *Store {
	return &Store{store: store}
}

func (store *Store) DeleteEventsBefore(ctx context.Context, createdAt time.Time) (int64, error) {
	deletedExecution, executionErr := store.store.Queries().DeleteExecutionEventsBefore(ctx, createdAt)
	if executionErr != nil {
		return 0, executionErr
	}

	deletedFileAccess, fileAccessErr := store.store.Queries().DeleteFileAccessEventsBefore(ctx, createdAt)
	if fileAccessErr != nil {
		return 0, fileAccessErr
	}

	return deletedExecution + deletedFileAccess, nil
}
