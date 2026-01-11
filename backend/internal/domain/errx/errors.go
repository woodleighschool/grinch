// Package errx defines domain level error types and maps them to HTTP semantics.
package errx

import (
	"context"
	"errors"
	"net/http"
)

// Code is a stable, machine readable domain error classification.
type Code string

const (
	CodeNotFound     Code = "not_found"
	CodeConflict     Code = "conflict"
	CodeInvalid      Code = "invalid"
	CodeUnauthorized Code = "unauthorized"
	CodeForbidden    Code = "forbidden"
	CodeRateLimited  Code = "rate_limited"
	CodeUnavailable  Code = "unavailable"
	CodeTimeout      Code = "timeout"
	CodeInternal     Code = "internal"
)

// Error represents the canonical domain error.
type Error struct {
	Code    Code
	Message string
	Fields  map[string]string
	Cause   error
}

// Error returns a human readable error message.
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Code != "" {
		return string(e.Code)
	}
	return "error"
}

// Unwrap returns the underlying cause, if any.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// StatusCode returns the HTTP status associated with this error.
func (e *Error) StatusCode() int {
	if e == nil {
		return http.StatusOK
	}
	return StatusForCode(e.Code)
}

// StatusForCode maps a domain error code to an HTTP status.
func StatusForCode(code Code) int {
	switch code {
	case CodeInvalid:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	case CodeRateLimited:
		return http.StatusTooManyRequests
	case CodeTimeout:
		return http.StatusGatewayTimeout
	case CodeUnavailable:
		return http.StatusServiceUnavailable
	case CodeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// StatusCoder is implemented by errors that can report an HTTP status code.
type StatusCoder interface {
	StatusCode() int
}

// Status resolves the HTTP status code for any error.
func Status(err error) int {
	if err == nil {
		return http.StatusOK
	}

	var sc StatusCoder
	if errors.As(err, &sc) && sc.StatusCode() != 0 {
		return sc.StatusCode()
	}

	if derr, ok := As(err); ok {
		return StatusForCode(derr.Code)
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return http.StatusGatewayTimeout
	}
	if errors.Is(err, context.Canceled) {
		return http.StatusRequestTimeout
	}

	return http.StatusInternalServerError
}

// WithStatus wraps an error with an explicit HTTP status override.
func WithStatus(err error, status int) error {
	if err == nil {
		return nil
	}
	return &statusError{err: err, status: status}
}

type statusError struct {
	err    error
	status int
}

// Error returns the wrapped error message.
func (e *statusError) Error() string { return e.err.Error() }

// Unwrap returns the wrapped error.
func (e *statusError) Unwrap() error { return e.err }

// StatusCode returns the overridden HTTP status.
func (e *statusError) StatusCode() int { return e.status }

// NotFound constructs a not_found domain error.
func NotFound(message string) *Error {
	return &Error{Code: CodeNotFound, Message: message}
}

// Conflict constructs a conflict domain error.
func Conflict(message string) *Error {
	return &Error{Code: CodeConflict, Message: message}
}

// Internal constructs an internal error wrapping a cause.
func Internal(message string, cause error) *Error {
	return &Error{Code: CodeInternal, Message: message, Cause: cause}
}

// As extracts a domain Error from an arbitrary error value.
func As(err error) (*Error, bool) {
	var derr *Error
	if errors.As(err, &derr) {
		return derr, true
	}
	return nil, false
}

// CodeOf returns the domain error code for an error, if present.
func CodeOf(err error) Code {
	if derr, ok := As(err); ok {
		return derr.Code
	}
	return ""
}

// IsCode reports whether an error has the given domain error code.
func IsCode(err error, code Code) bool {
	return CodeOf(err) == code
}
