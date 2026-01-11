package errx

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// FromStore maps store layer errors into domain errors.
//
// constraints maps database constraint names to field names for validation errors.
func FromStore(err error, constraints map[string]string) error {
	if err == nil {
		return nil
	}

	if _, ok := As(err); ok {
		return err
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return NotFound("not found")
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return Internal("internal error", err)
	}

	switch pgErr.Code {
	case "23505": // unique_violation
		if field := constraintField(pgErr.ConstraintName, constraints); field != "" {
			return validation(field, "Must be unique")
		}
		return Conflict("conflict")

	case "23503": // foreign_key_violation
		if field := constraintField(pgErr.ConstraintName, constraints); field != "" {
			return validation(field, "Invalid reference")
		}
		return validation("", "Invalid reference")

	case "23514": // check_violation
		if field := constraintField(pgErr.ConstraintName, constraints); field != "" {
			return validation(field, "Invalid value")
		}
		return validation("", "Invalid value")

	default:
		return Internal("internal error", err)
	}
}

func constraintField(name string, constraints map[string]string) string {
	if name == "" || len(constraints) == 0 {
		return ""
	}
	return constraints[name]
}

func validation(field, message string) *Error {
	var fields map[string]string
	if field != "" {
		fields = map[string]string{field: message}
	}
	return &Error{
		Code:    CodeInvalid,
		Message: "Some fields are invalid",
		Fields:  fields,
	}
}
