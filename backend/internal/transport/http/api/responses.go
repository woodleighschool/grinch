package apihttp

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/woodleighschool/grinch/internal/domain"
)

func writeJSON(w http.ResponseWriter, statusCode int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(v)
}

func writeNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// writeValidationErrors writes a React Admin compatible validation error body.
// Shape: {"errors": {"root": {"serverError": "..."}, "field": "message", ...}}.
func writeValidationErrors(w http.ResponseWriter, root string, fields []domain.FieldError) {
	errs := make(map[string]any, len(fields)+1)
	if root != "" {
		errs["root"] = map[string]string{"serverError": root}
	}

	for _, fieldErr := range fields {
		errs[fieldErr.Field] = fieldErr.Message
	}

	writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs})
}

func writeError(w http.ResponseWriter, err error) {
	var validationErr *domain.ValidationError
	var badReqErr badRequestError

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		w.WriteHeader(http.StatusNotFound)
		return
	case errors.Is(err, domain.ErrGroupReadOnly):
		w.WriteHeader(http.StatusForbidden)
		return
	case errors.Is(err, domain.ErrInvalidSort), errors.As(err, &badReqErr):
		w.WriteHeader(http.StatusBadRequest)
		return
	case errors.As(err, &validationErr):
		writeValidationErrors(w, validationErr.Detail, validationErr.FieldErrors)
		return
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			writeValidationErrors(w, "Resource already exists.", nil)
			return
		case pgerrcode.ForeignKeyViolation:
			writeValidationErrors(w, "Referenced resource does not exist.", nil)
			return
		}
	}

	w.WriteHeader(http.StatusInternalServerError)
}
