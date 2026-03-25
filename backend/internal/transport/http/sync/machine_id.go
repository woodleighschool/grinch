package synchttp

import (
	"fmt"

	"github.com/google/uuid"

	appsanta "github.com/woodleighschool/grinch/internal/app/santa"
)

func parseMachineID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: parse machine_id %q: %w", appsanta.ErrInvalidSyncRequest, raw, err)
	}

	return id, nil
}
