package errx_test

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/bjaus/errx"
)

// ExampleNew demonstrates creating a simple error with a code and message.
func ExampleNew() {
	err := errx.New(errx.CodeNotFound, "user not found")

	fmt.Println("Code:", err.Code())
	fmt.Println("Message:", err.Error())
	fmt.Println("Error:", err.Error())

	// Output:
	// Code: not_found
	// Message: user not found
	// Error: user not found
}

// ExampleNewf demonstrates creating an error with formatted message.
func ExampleNewf() {
	userID := 12345
	err := errx.Newf(errx.CodeInvalidArgument, "invalid user ID: %d", userID)

	fmt.Println(err.Error())

	// Output:
	// invalid user ID: 12345
}

// ExampleWrap demonstrates wrapping an existing error with errx context.
func ExampleWrap() {
	// Simulate a database error
	dbErr := errors.New("connection timeout")

	// Wrap it with errx error
	err := errx.Wrap(dbErr, errx.CodeUnavailable, "database unavailable")

	fmt.Println(err.Error())
	fmt.Println("Code:", err.Code())
	fmt.Println("Original error found:", errors.Is(err, dbErr))

	// Output:
	// database unavailable
	// Code: unavailable
	// Original error found: true
}

// ExampleError_WithDetail demonstrates adding client-safe details.
func ExampleError_WithDetail() {
	err := errx.New(errx.CodeInvalidArgument, "validation failed").
		WithDetail("field", "email").
		WithDetail("reason", "invalid format")

	fmt.Println("Details:", err.Details())

	// These details are safe to expose to clients in the error response

	// Output:
	// Details: map[field:email reason:invalid format]
}

// ExampleError_WithMeta demonstrates adding internal metadata for debugging.
func ExampleError_WithMeta() {
	err := errx.New(errx.CodeInternal, "operation failed").
		WithMeta("user_id", 12345).
		WithMeta("operation", "update_profile").
		WithMeta("retry_count", 3)

	// Print metadata fields individually for deterministic output
	metadata := err.Metadata()
	fmt.Println("user_id:", metadata["user_id"])
	fmt.Println("operation:", metadata["operation"])
	fmt.Println("retry_count:", metadata["retry_count"])

	// This metadata is NOT exposed to clients, only used in logs/debugging

	// Output:
	// user_id: 12345
	// operation: update_profile
	// retry_count: 3
}

// ExampleError_WithDebug demonstrates separating client messages from debug details.
func ExampleError_WithDebug() {
	err := errx.New(errx.CodeInternal, "service temporarily unavailable").
		WithDebug("failed to connect to postgres.internal.company.com:5432")

	fmt.Println("Client sees:", err.Error())
	fmt.Println("Logs contain:", err.DebugMessage())

	// Output:
	// Client sees: service temporarily unavailable
	// Logs contain: [internal] service temporarily unavailable | debug=failed to connect to postgres.internal.company.com:5432
}

// ExampleError_WithSource demonstrates tagging errors with source (service/package/component).
func ExampleError_WithSource() {
	err := errx.New(errx.CodeUnavailable, "service unavailable").
		WithSource("payment-service")

	fmt.Println("Source:", err.Source())

	// Useful for identifying which service/component generated the error

	// Output:
	// Source: payment-service
}

// ExampleError_WithTags demonstrates categorizing errors with tags.
func ExampleError_WithTags() {
	err := errx.New(errx.CodeInternal, "operation failed").
		WithTags("database", "postgres").
		WithTags("critical")

	fmt.Println("Tags:", err.Tags())

	// Tags can be used for filtering, alerting, or categorizing errors

	// Output:
	// Tags: [database postgres critical]
}

// ExampleError_layeredArchitecture demonstrates error handling through application layers.
func ExampleError_layeredArchitecture() {
	// 1. Database layer error
	dbErr := errors.New("connection refused")

	// 2. Repository layer wraps and adds context
	repoErr := errx.Wrap(dbErr, errx.CodeUnavailable, "database query failed").
		WithSource("user-repository").
		WithTags("database").
		WithMeta("query", "SELECT * FROM users WHERE id = ?").
		WithDebug("postgres connection pool exhausted")

	// 3. Service layer wraps again with business context
	serviceErr := errx.Wrap(repoErr, errx.CodeNotFound, "user not found").
		WithSource("user-service").
		WithDetail("user_id", "12345")

	// Each layer sees its own context
	fmt.Println("Service layer sees:", serviceErr.Error())
	fmt.Println("Source:", serviceErr.Source())

	// Can check for any error in the chain
	fmt.Println("Contains DB error:", errors.Is(serviceErr, dbErr))

	// Output:
	// Service layer sees: user not found
	// Source: user-service
	// Contains DB error: true
}

// ExampleCodeIs demonstrates checking error codes.
func ExampleCodeIs() {
	err := errx.New(errx.CodeNotFound, "resource not found")

	// Check if error has specific code
	if errx.CodeIs(err, errx.CodeNotFound) {
		fmt.Println("Resource not found")
	}

	// Works with wrapped errors too
	wrappedErr := fmt.Errorf("operation failed: %w", err)
	if errx.CodeIs(wrappedErr, errx.CodeNotFound) {
		fmt.Println("Still found through wrapper")
	}

	// Output:
	// Resource not found
	// Still found through wrapper
}

// ExampleCodeIn demonstrates checking if error matches any of several codes.
func ExampleCodeIn() {
	err := errx.New(errx.CodeNotFound, "resource not found")

	// Check if error matches any of the provided codes
	if errx.CodeIn(err, errx.CodeNotFound, errx.CodeUnauthenticated) {
		fmt.Println("Client error occurred")
	}

	// Output:
	// Client error occurred
}

// ExampleCodeOf demonstrates extracting error code from any error.
func ExampleCodeOf() {
	// From errx error
	err1 := errx.New(errx.CodeInvalidArgument, "bad input")
	fmt.Println("errx error code:", errx.CodeOf(err1))

	// From standard error (returns CodeUnknown)
	err2 := errors.New("standard error")
	fmt.Println("standard error code:", errx.CodeOf(err2))

	// From wrapped errx error
	err3 := fmt.Errorf("wrapped: %w", err1)
	fmt.Println("wrapped error code:", errx.CodeOf(err3))

	// Output:
	// errx error code: invalid_argument
	// standard error code: unknown
	// wrapped error code: invalid_argument
}

// ExampleError_slogJSON demonstrates structured logging output with nested errors.
func ExampleError_slogJSON() {
	// Create a logger that outputs JSON
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			// Remove timestamp for consistent output
			if a.Key == "time" {
				return slog.Attr{}
			}
			return a
		},
	}))

	// Scenario: Inner error from repository
	innerErr := errx.New(errx.CodeInternal, "query execution failed").
		WithSource("user-repository").
		WithTags("database", "postgres").
		WithMeta("query", "SELECT * FROM users").
		WithDebug("deadlock detected")

	// Outer error from service
	outerErr := errx.Wrap(innerErr, errx.CodeNotFound, "user not found").
		WithSource("user-service").
		WithDetail("user_id", "12345").
		WithMeta("request_id", "req-789")

	// Log it
	logger.Error("operation failed", "error", outerErr)

	// Output:
	// {"level":"ERROR","msg":"operation failed","error":{"code":"not_found","message":"user not found","source":"user-service","details":{"user_id":"12345"},"metadata":{"request_id":"req-789"},"cause":{"code":"internal","message":"query execution failed","source":"user-repository","tags":["database","postgres"],"metadata":{"query":"SELECT * FROM users"},"debug":"deadlock detected"}}}
}

// ExampleError_slogJSON_threeLevels demonstrates three-level nested error logging.
func ExampleError_slogJSON_threeLevels() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == "time" {
				return slog.Attr{}
			}
			return a
		},
	}))

	// Level 1: Database driver
	dbErr := errx.New(errx.CodeInternal, "connection refused").
		WithSource("postgres-driver").
		WithMeta("port", 5432)

	// Level 2: Repository
	repoErr := errx.Wrap(dbErr, errx.CodeUnavailable, "database unavailable").
		WithSource("repository").
		WithMeta("operation", "findByID")

	// Level 3: Service
	serviceErr := errx.Wrap(repoErr, errx.CodeNotFound, "resource not found").
		WithSource("service").
		WithMeta("resource_type", "user")

	logger.Error("request failed", "error", serviceErr)

	// Note how each layer is nested in the "cause" field, preserving all context

	// Output:
	// {"level":"ERROR","msg":"request failed","error":{"code":"not_found","message":"resource not found","source":"service","metadata":{"resource_type":"user"},"cause":{"code":"unavailable","message":"database unavailable","source":"repository","metadata":{"operation":"findByID"},"cause":{"code":"internal","message":"connection refused","source":"postgres-driver","metadata":{"port":5432}}}}}
}

// ExampleError_clientVsInternalData demonstrates the separation between client-safe
// and internal debugging data.
func ExampleError_clientVsInternalData() {
	// Create an error with both client-safe and internal data
	err := errx.New(errx.CodeInternal, "service temporarily unavailable").
		WithDetail("retry_after", "30s").                        // Client sees this
		WithSource("payment-service").                           // Internal only
		WithTags("external", "payment-provider").                // Internal only
		WithMeta("transaction_id", "txn_123456").                // Internal only
		WithMeta("retry_count", 3).                              // Internal only
		WithDebug("payment provider returned 500 after retries") // Internal only

	fmt.Println("=== Client-Safe Data (can be exposed in API responses) ===")
	fmt.Println("Message:", err.Error())
	// Print details field individually for deterministic output
	details := err.Details()
	fmt.Println("retry_after:", details["retry_after"])

	fmt.Println("\n=== Internal Data (for logs/debugging only) ===")
	fmt.Println("Source:", err.Source())
	fmt.Println("Tags:", err.Tags())
	// Print metadata fields individually for deterministic output
	metadata := err.Metadata()
	fmt.Println("transaction_id:", metadata["transaction_id"])
	fmt.Println("retry_count:", metadata["retry_count"])

	// Output:
	// === Client-Safe Data (can be exposed in API responses) ===
	// Message: service temporarily unavailable
	// retry_after: 30s
	//
	// === Internal Data (for logs/debugging only) ===
	// Source: payment-service
	// Tags: [external payment-provider]
	// transaction_id: txn_123456
	// retry_count: 3
}
