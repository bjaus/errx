package errx

import (
	"errors"
	"fmt"
	"runtime"
	"slices"
)

// stackSkipDepth is the number of stack frames to skip when capturing the stack trace.
// This skips: runtime.Callers (1) + captureStackTrace (2) + newError (3) + public function (4)
// so that the stack trace starts at the actual caller of New/Newf/Wrap/Wrapf.
const stackSkipDepth = 4

// New creates a new Error with the given code and message.
// The message should be safe to expose to clients.
func New(code Code, message string) *Error {
	return newError(code, message, nil)
}

// Newf creates a new Error with a formatted message.
// The message should be safe to expose to clients.
func Newf(code Code, format string, args ...any) *Error {
	return newError(code, fmt.Sprintf(format, args...), nil)
}

// Wrap wraps an existing error with additional context and an error code.
// The message should be safe to expose to clients.
func Wrap(err error, code Code, message string) *Error {
	if err == nil {
		return nil
	}
	return newError(code, message, err)
}

// Wrapf wraps an existing error with a formatted message.
// The message should be safe to expose to clients.
func Wrapf(err error, code Code, format string, args ...any) *Error {
	if err == nil {
		return nil
	}
	return newError(code, fmt.Sprintf(format, args...), err)
}

// Ensure guarantees the returned error is an *Error.
// If err is nil, returns nil.
// If err is already an *Error (or wraps one), returns the existing *Error unchanged.
// Otherwise, wraps err as the cause with the given code and message.
func Ensure(err error, code Code, message string) *Error {
	if err == nil {
		return nil
	}
	if e, ok := As(err); ok {
		return e
	}
	return newError(code, message, err)
}

// Ensuref is like Ensure but with a formatted message for the fallback case.
func Ensuref(err error, code Code, format string, args ...any) *Error {
	if err == nil {
		return nil
	}
	if e, ok := As(err); ok {
		return e
	}
	return newError(code, fmt.Sprintf(format, args...), err)
}

// Is checks if err is or wraps an *Error.
func Is(err error) bool {
	return IsType[*Error](err)
}

// As finds the first *Error in err's tree.
// Returns the error value and true if found, or nil and false otherwise.
func As(err error) (*Error, bool) {
	return AsType[*Error](err)
}

// IsType checks if err is or wraps the specified error type E.
func IsType[E error](err error) bool {
	_, ok := AsType[E](err)
	return ok
}

// AsType finds the first error in err's tree that matches the type E.
// Returns the error value and true if found, or zero value and false otherwise.
func AsType[E error](err error) (E, bool) {
	if err == nil {
		var zero E
		return zero, false
	}
	var pe E
	valid := errors.As(err, &pe)
	return pe, valid
}

// CodeOf extracts the error code from an error.
// Returns CodeUnknown if the error is not an *Error.
func CodeOf(err error) Code {
	if e, ok := As(err); ok {
		return e.code
	}
	return CodeUnknown
}

// CodeIs checks if an error has a specific error code.
// It unwraps the error chain to find an *Error.
func CodeIs(err error, code Code) bool {
	if e, ok := As(err); ok {
		return e.code == code
	}
	return false
}

// CodeIn checks if an error has a code matching any of the provided codes.
// It unwraps the error chain to find an *Error.
func CodeIn(err error, codes ...Code) bool {
	e, ok := As(err)
	if !ok {
		return false
	}
	return slices.Contains(codes, e.code)
}

// IsRetryable checks if an error indicates a retryable operation.
// Returns false if the error is not an *Error.
func IsRetryable(err error) bool {
	e, ok := As(err)
	if !ok {
		return false
	}
	return e.IsRetryable()
}

// newError is an internal helper that creates an Error with the given parameters.
func newError(code Code, message string, cause error) *Error {
	return &Error{
		code:       code,
		message:    message,
		cause:      cause,
		details:    make(map[string]any),
		metadata:   make(map[string]any),
		stackTrace: captureStackTrace(stackSkipDepth),
	}
}

// captureStackTrace captures the current stack trace.
func captureStackTrace(skip int) []uintptr {
	const maxDepth = 32
	pcs := make([]uintptr, maxDepth)
	n := runtime.Callers(skip, pcs)
	return pcs[:n]
}
