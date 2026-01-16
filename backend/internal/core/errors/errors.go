package coreerrors

import "errors"

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

// Error represents the canonical domain error used across transports.
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
