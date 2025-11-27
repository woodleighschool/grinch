package admin

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// parseInt32 best-effort parses query params while providing defaults.
func parseInt32(value string, fallback int32) int32 {
	if value == "" {
		return fallback
	}
	if n, err := strconv.ParseInt(value, 10, 32); err == nil {
		return int32(n)
	}
	return fallback
}

// parseUUIDParam pulls a path parameter and parses it as a UUID.
func parseUUIDParam(r *http.Request, key string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, key))
}
