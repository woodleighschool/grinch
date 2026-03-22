package postgres

import (
	"context"
	"time"
)

func (store *Store) DeleteEventsBefore(ctx context.Context, createdAt time.Time) (int64, error) {
	deletedExecution, executionErr := store.Queries().DeleteExecutionEventsBefore(ctx, createdAt)
	if executionErr != nil {
		return 0, executionErr
	}

	deletedFileAccess, fileAccessErr := store.Queries().DeleteFileAccessEventsBefore(ctx, createdAt)
	if fileAccessErr != nil {
		return 0, fileAccessErr
	}

	return deletedExecution + deletedFileAccess, nil
}
