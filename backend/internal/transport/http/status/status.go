package status

import (
	"context"
	"errors"
	"net/http"

	coreerrors "github.com/woodleighschool/grinch/internal/core/errors"
)

// Status resolves the HTTP status code for any error.
func Status(err error) int {
	if err == nil {
		return http.StatusOK
	}

	var se statusCoder
	if errors.As(err, &se) && se.StatusCode() != 0 {
		return se.StatusCode()
	}

	if derr, ok := coreerrors.As(err); ok {
		return statusForCode(derr.Code)
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return http.StatusGatewayTimeout
	}
	if errors.Is(err, context.Canceled) {
		return http.StatusRequestTimeout
	}

	return http.StatusInternalServerError
}

// WithStatus wraps err with an explicit HTTP status override.
func WithStatus(err error, status int) error {
	if err == nil {
		return nil
	}
	return &statusCodeError{err: err, status: status}
}

type statusCoder interface {
	StatusCode() int
}

type statusCodeError struct {
	err    error
	status int
}

func (e *statusCodeError) Error() string   { return e.err.Error() }
func (e *statusCodeError) Unwrap() error   { return e.err }
func (e *statusCodeError) StatusCode() int { return e.status }

func statusForCode(code coreerrors.Code) int {
	switch code {
	case coreerrors.CodeInvalid:
		return http.StatusBadRequest
	case coreerrors.CodeUnauthorized:
		return http.StatusUnauthorized
	case coreerrors.CodeForbidden:
		return http.StatusForbidden
	case coreerrors.CodeNotFound:
		return http.StatusNotFound
	case coreerrors.CodeConflict:
		return http.StatusConflict
	case coreerrors.CodeRateLimited:
		return http.StatusTooManyRequests
	case coreerrors.CodeTimeout:
		return http.StatusGatewayTimeout
	case coreerrors.CodeUnavailable:
		return http.StatusServiceUnavailable
	case coreerrors.CodeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
