package helpers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	coreerrors "github.com/woodleighschool/grinch/internal/core/errors"
	httpstatus "github.com/woodleighschool/grinch/internal/transport/http/status"
)

// ParseUUID returns a UUID URL parameter or a not found error.
func ParseUUID(r *http.Request, param string) (uuid.UUID, error) {
	raw := chi.URLParam(r, param)
	if raw == "" {
		return uuid.Nil, coreerrors.NotFound("not found")
	}

	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, coreerrors.NotFound("not found")
	}

	return id, nil
}

// ParseUUIDValue parses a UUID from v and returns uuid.Nil on failure.
func ParseUUIDValue(v any) uuid.UUID {
	s, ok := v.(string)
	if !ok || s == "" {
		return uuid.Nil
	}

	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}

	return id
}

// DecodeJSON decodes a JSON request body into T.
func DecodeJSON[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, httpstatus.WithStatus(err, http.StatusBadRequest)
	}
	return v, nil
}
