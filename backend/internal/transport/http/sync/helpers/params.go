package helpers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func ParseMachineID(r *http.Request) (uuid.UUID, error) {
	raw := chi.URLParam(r, "machine_id")
	if raw == "" {
		return uuid.Nil, errors.New("machine_id parameter is required")
	}

	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("machine_id must be a valid UUID")
	}

	return id, nil
}

func ParseCursor(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	if value < 0 {
		return 0, errors.New("negative offset")
	}

	return value, nil
}
