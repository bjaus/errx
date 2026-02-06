package errx

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
)

// Compile-time interface assertions
//
//nolint:errcheck // These are compile-time interface checks, not error returns
var (
	_ error          = (*Error)(nil)
	_ slog.LogValuer = (*Error)(nil)
)

// Error represents a rich error with code, context, and debugging information.
// It implements the standard error interface and supports error wrapping.
type Error struct {
	code         Code
	message      string         // Client-safe message
	debugMessage string         // Internal debug message
	cause        error          // Wrapped error
	source       string         // Source (service/package/component) where error occurred
	tags         []string       // Tags for categorization
	details      map[string]any // Client-safe key-value details
	metadata     map[string]any // Internal debug metadata
	stackTrace   []uintptr      // Stack trace
	retryable    bool           // Whether the error indicates a retryable operation
}

// Code returns the error code.
// Returns CodeUnknown if the error is nil.
func (e *Error) Code() Code {
	if e == nil {
		return CodeUnknown
	}
	return e.code
}

// Error implements the error interface.
// It returns the client-safe message.
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.message
}

// Unwrap returns the wrapped error, supporting errors.Unwrap.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// DebugMessage returns a detailed debug message with all context.
// This should only be logged or shown to system maintainers, never to clients.
func (e *Error) DebugMessage() string {
	if e == nil {
		return ""
	}

	var parts []string

	// Add code and message
	parts = append(parts, fmt.Sprintf("[%s] %s", e.code.String(), e.message))

	// Add source if present
	if e.source != "" {
		parts = append(parts, fmt.Sprintf("source=%s", e.source))
	}

	// Add tags if present
	if len(e.tags) > 0 {
		parts = append(parts, fmt.Sprintf("tags=%v", e.tags))
	}

	// Add details if present
	if len(e.details) > 0 {
		parts = append(parts, fmt.Sprintf("details=%v", e.details))
	}

	// Add metadata if present
	if len(e.metadata) > 0 {
		parts = append(parts, fmt.Sprintf("metadata=%v", e.metadata))
	}

	// Add retryable status if true
	if e.retryable {
		parts = append(parts, "retryable=true")
	}

	// Add debug message if different from message
	if e.debugMessage != "" && e.debugMessage != e.message {
		parts = append(parts, fmt.Sprintf("debug=%s", e.debugMessage))
	}

	// Add wrapped error
	if e.cause != nil {
		parts = append(parts, fmt.Sprintf("cause=%v", e.cause))
	}

	return strings.Join(parts, " | ")
}

// WithDetail adds a client-safe key-value detail to the error.
// These details are safe to expose to clients and are typically included in error responses.
func (e *Error) WithDetail(key string, value any) *Error {
	if e == nil {
		return nil
	}
	e.details[key] = value
	return e
}

// WithMetaFromContext pulls metadata stored via [WithMetaContext] from the context and merges it
// into the error's internal metadata map. Context values overwrite existing keys
// (last-write-wins), so call ordering determines precedence:
//
//	// Context wins — post_id will be 42:
//	errx.New(code, msg).WithMeta("post_id", 99).WithMetaFromContext(ctx)
//
//	// WithMeta wins — post_id will be 99:
//	errx.New(code, msg).WithMetaFromContext(ctx).WithMeta("post_id", 99)
func (e *Error) WithMetaFromContext(ctx context.Context) *Error {
	if e == nil {
		return nil
	}
	for k, v := range getCtxMeta(ctx) {
		e.metadata[k] = v
	}
	return e
}

// WithMeta adds a key-value pair to the error's internal metadata.
// This metadata is included in debug messages but NOT exposed to clients.
func (e *Error) WithMeta(key string, value any) *Error {
	if e == nil {
		return nil
	}
	e.metadata[key] = value
	return e
}

// WithDebug sets an internal debug message with additional implementation details.
// This is only shown in debug messages, never to clients.
func (e *Error) WithDebug(message string) *Error {
	if e == nil {
		return nil
	}
	e.debugMessage = message
	return e
}

// WithDebugf sets a formatted internal debug message.
func (e *Error) WithDebugf(format string, args ...any) *Error {
	return e.WithDebug(fmt.Sprintf(format, args...))
}

// WithSource sets the source (service/package/component) where the error occurred.
func (e *Error) WithSource(source string) *Error {
	if e == nil {
		return nil
	}
	e.source = source
	return e
}

// WithTags adds one or more tags to categorize the error.
func (e *Error) WithTags(tags ...string) *Error {
	if e == nil {
		return nil
	}
	e.tags = append(e.tags, tags...)
	return e
}

// WithRetryable marks the error as representing a retryable operation.
// This indicates that the same request can be retried and may succeed.
func (e *Error) WithRetryable() *Error {
	if e == nil {
		return nil
	}
	e.retryable = true
	return e
}

// Source returns the source (service/package/component) where the error occurred.
func (e *Error) Source() string {
	if e == nil {
		return ""
	}
	return e.source
}

// Tags returns the error's tags.
func (e *Error) Tags() []string {
	if e == nil {
		return nil
	}
	return e.tags
}

// Details returns the client-safe error details.
func (e *Error) Details() map[string]any {
	if e == nil {
		return nil
	}
	return e.details
}

// Metadata returns the error's internal debug metadata.
func (e *Error) Metadata() map[string]any {
	if e == nil {
		return nil
	}
	return e.metadata
}

// StackTrace returns the captured stack trace.
func (e *Error) StackTrace() []uintptr {
	if e == nil {
		return nil
	}
	return e.stackTrace
}

// IsRetryable returns whether the error indicates a retryable operation.
func (e *Error) IsRetryable() bool {
	if e == nil {
		return false
	}
	return e.retryable
}

// FormatStackTrace returns a human-readable stack trace.
func (e *Error) FormatStackTrace() string {
	if e == nil || len(e.stackTrace) == 0 {
		return ""
	}

	frames := runtime.CallersFrames(e.stackTrace)
	var lines []string

	for {
		frame, more := frames.Next()
		lines = append(lines, fmt.Sprintf("%s\n\t%s:%d", frame.Function, frame.File, frame.Line))
		if !more {
			break
		}
	}

	return strings.Join(lines, "\n")
}

// Is supports error comparison with errors.Is.
// Two errors are considered equal if they have the same code.
func (e *Error) Is(target error) bool {
	if e == nil {
		return target == nil
	}

	t, ok := target.(*Error)
	if !ok {
		return false
	}

	return e.code == t.code
}

// LogValue implements slog.LogValuer for structured logging integration.
// Returns a slog.GroupValue containing all error fields for debugging.
func (e *Error) LogValue() slog.Value {
	if e == nil {
		return slog.Value{}
	}

	attrs := []slog.Attr{
		slog.String("code", e.code.String()),
		slog.String("message", e.message),
	}

	if e.source != "" {
		attrs = append(attrs, slog.String("source", e.source))
	}

	if len(e.tags) > 0 {
		attrs = append(attrs, slog.Any("tags", e.tags))
	}

	if len(e.details) > 0 {
		attrs = append(attrs, slog.Any("details", e.details))
	}

	if len(e.metadata) > 0 {
		attrs = append(attrs, slog.Any("metadata", e.metadata))
	}

	if e.retryable {
		attrs = append(attrs, slog.Bool("retryable", true))
	}

	if e.debugMessage != "" && e.debugMessage != e.message {
		attrs = append(attrs, slog.String("debug", e.debugMessage))
	}

	if e.cause != nil {
		attrs = append(attrs, slog.Any("cause", e.cause))
	}

	return slog.GroupValue(attrs...)
}
