package postgres

import (
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsRetryableEventIngestError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "deadlock",
			err:  &pgconn.PgError{Code: "40P01"},
			want: true,
		},
		{
			name: "serialization failure",
			err:  &pgconn.PgError{Code: "40001"},
			want: true,
		},
		{
			name: "wrapped retryable",
			err:  fmt.Errorf("ingest events: %w", &pgconn.PgError{Code: "40P01"}),
			want: true,
		},
		{
			name: "invalid byte sequence",
			err:  &pgconn.PgError{Code: "22021"},
		},
		{
			name: "non postgres error",
			err:  fmt.Errorf("plain error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRetryableEventIngestError(tt.err); got != tt.want {
				t.Fatalf("isRetryableEventIngestError() = %v, want %v", got, tt.want)
			}
		})
	}
}
