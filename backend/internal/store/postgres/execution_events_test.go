package postgres //nolint:testpackage // exercises unexported transaction planning and retry classification.

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/woodleighschool/grinch/internal/santa/model"
)

func TestOrderedDistinctExecutablesUsesCanonicalIdentityOrder(t *testing.T) {
	events := []model.ExecutionEventWrite{
		{Executable: model.ExecutableWrite{FileSHA256: "sha-b", FileName: "B"}},
		{Executable: model.ExecutableWrite{FileSHA256: "sha-a", FileName: "Z"}},
		{Executable: model.ExecutableWrite{FileSHA256: "sha-a", FileName: "A"}},
	}

	got := orderedDistinctExecutables(events)

	if len(got) != 3 {
		t.Fatalf("orderedDistinctExecutables() returned %d executables, want 3", len(got))
	}

	want := [][2]string{
		{"sha-a", "A"},
		{"sha-a", "Z"},
		{"sha-b", "B"},
	}
	for i, executable := range got {
		identity := [2]string{executable.FileSHA256, executable.FileName}
		if identity != want[i] {
			t.Fatalf("orderedDistinctExecutables()[%d] identity = %q, want %q", i, identity, want[i])
		}
	}
}

func TestOrderedDistinctExecutablesKeepsFirstPayloadPerIdentity(t *testing.T) {
	events := []model.ExecutionEventWrite{
		{
			Executable: model.ExecutableWrite{
				FileSHA256: "same-sha",
				FileName:   "same-name",
				SigningID:  "first",
			},
		},
		{
			Executable: model.ExecutableWrite{
				FileSHA256: "same-sha",
				FileName:   "same-name",
				SigningID:  "later",
			},
		},
	}

	got := orderedDistinctExecutables(events)

	if len(got) != 1 {
		t.Fatalf("orderedDistinctExecutables() returned %d executables, want 1", len(got))
	}
	if got[0].SigningID != "first" {
		t.Fatalf("orderedDistinctExecutables()[0].SigningID = %q, want first payload preserved", got[0].SigningID)
	}
}

func TestIsRetryableEventIngestError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "deadlock",
			err:  &pgconn.PgError{Code: pgErrDeadlockDetected},
			want: true,
		},
		{
			name: "serialization failure",
			err:  &pgconn.PgError{Code: pgErrSerializationFailure},
			want: true,
		},
		{
			name: "wrapped retryable",
			err:  fmt.Errorf("ingest events: %w", &pgconn.PgError{Code: pgErrDeadlockDetected}),
			want: true,
		},
		{
			name: "invalid byte sequence",
			err:  &pgconn.PgError{Code: "22021"},
		},
		{
			name: "non postgres error",
			err:  errors.New("plain error"),
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
