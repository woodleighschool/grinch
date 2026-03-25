package postgres

import (
	"context"
	"time"
)

func (s *Store) DeleteEventsBefore(ctx context.Context, createdAt time.Time) (int64, error) {
	deletedExecution, executionErr := s.Queries().DeleteExecutionEventsBefore(ctx, createdAt)
	if executionErr != nil {
		return 0, executionErr
	}

	deletedFileAccess, fileAccessErr := s.Queries().DeleteFileAccessEventsBefore(ctx, createdAt)
	if fileAccessErr != nil {
		return 0, fileAccessErr
	}

	return deletedExecution + deletedFileAccess, nil
}
