// Package rest provides shared HTTP helpers for the react-admin simple-rest data provider.
package rest

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain/errx"
	"github.com/woodleighschool/grinch/internal/logging"
)

// Logger is the logging interface used by this package.
type Logger = logging.Logger

// WriteJSON writes v as a JSON response with the given status code.
func WriteJSON(ctx context.Context, w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		logging.FromContext(ctx).WarnContext(ctx, "encode json", "error", err)
	}
}

// WriteError writes an error response in the format expected by react-admin.
func WriteError(ctx context.Context, w http.ResponseWriter, log Logger, err error, msg string) {
	status := errx.Status(err)

	if status >= http.StatusInternalServerError {
		log.ErrorContext(ctx, msg, "error", err)
	} else {
		log.WarnContext(ctx, msg, "error", err)
	}

	resp := struct {
		Errors map[string]string `json:"errors"`
	}{
		Errors: make(map[string]string),
	}

	if derr, ok := errx.As(err); ok {
		switch {
		case len(derr.Fields) > 0:
			resp.Errors = derr.Fields
		case derr.Message != "":
			resp.Errors["root"] = derr.Message
		default:
			resp.Errors["root"] = http.StatusText(status)
		}
	} else {
		resp.Errors["root"] = http.StatusText(status)
	}

	WriteJSON(ctx, w, status, resp)
}

// ParseUUID returns a UUID URL parameter or a not found error.
func ParseUUID(r *http.Request, param string) (uuid.UUID, error) {
	raw := chi.URLParam(r, param)
	if raw == "" {
		return uuid.Nil, errx.NotFound("not found")
	}

	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errx.NotFound("not found")
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
		return v, errx.WithStatus(err, http.StatusBadRequest)
	}
	return v, nil
}
