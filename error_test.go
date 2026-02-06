package errx_test

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/bjaus/errx"
)

type errorSuite struct {
	suite.Suite
}

func TestErrorSuite(t *testing.T) {
	suite.Run(t, new(errorSuite))
}

func (s *errorSuite) TestErrorBuilder() {
	err := errx.New(errx.CodePermissionDenied, "access denied").
		WithSource("auth-service").
		WithTags("security", "rbac").
		WithDetail("resource", "admin-panel").
		WithMeta("user_id", "123").
		WithDebug("user missing admin role")

	s.Equal(errx.CodePermissionDenied, err.Code())
	s.Equal("access denied", err.Error())
	s.Equal("auth-service", err.Source())
	s.Equal([]string{"security", "rbac"}, err.Tags())
	s.Equal("admin-panel", err.Details()["resource"])
	s.Equal("123", err.Metadata()["user_id"])

	debugMsg := err.DebugMessage()
	s.Contains(debugMsg, "permission_denied")
	s.Contains(debugMsg, "access denied")
	s.Contains(debugMsg, "source=auth-service")
	s.Contains(debugMsg, "tags=[security rbac]")
	s.Contains(debugMsg, "details=map[resource:admin-panel]")
	s.Contains(debugMsg, "metadata=map[user_id:123]")
	s.Contains(debugMsg, "user missing admin role")
}

func (s *errorSuite) TestErrorsIs() {
	err1 := errx.New(errx.CodeNotFound, "not found")
	err2 := errx.New(errx.CodeNotFound, "different message")
	err3 := errx.New(errx.CodeInternal, "internal")

	s.True(errors.Is(err1, err2), "errors with same code should match")
	s.False(errors.Is(err1, err3), "errors with different codes should not match")
}

func (s *errorSuite) TestErrorsAs() {
	baseErr := errors.New("base error")
	wrappedErr := errx.Wrap(baseErr, errx.CodeInternal, "wrapped")

	var errxErr *errx.Error
	s.Require().True(errors.As(wrappedErr, &errxErr))
	s.Equal(errx.CodeInternal, errxErr.Code())
}

func (s *errorSuite) TestStackTrace() {
	err := errx.New(errx.CodeInternal, "internal error")

	stackTrace := err.StackTrace()
	s.NotEmpty(stackTrace)

	formatted := err.FormatStackTrace()
	s.NotEmpty(formatted)
	s.Contains(formatted, "TestStackTrace")
}

func (s *errorSuite) TestNilError() {
	var err *errx.Error

	s.Equal(errx.CodeUnknown, err.Code())
	s.Equal("", err.Error())
	s.Equal("", err.Error())
	s.Equal("", err.DebugMessage())
	s.Equal("", err.Source())
	s.Nil(err.Tags())
	s.Nil(err.Details())
	s.Nil(err.Metadata())
	s.Nil(err.StackTrace())
	s.Equal("", err.FormatStackTrace())

	// Test Unwrap on nil error
	s.Nil(err.Unwrap())

	// Test Is on nil error
	s.True(err.Is(nil))
	s.False(err.Is(errors.New("other")))

	// Builder methods should handle nil gracefully
	s.Nil(err.WithDetail("key", "value"))
	s.Nil(err.WithMeta("key", "value"))
	s.Nil(err.WithDebug("debug"))
	s.Nil(err.WithSource("source"))
	s.Nil(err.WithTags("tag"))
	s.Nil(err.WithRetryable())

	// IsRetryable should return false for nil error
	s.False(err.IsRetryable())
}

func (s *errorSuite) TestDebugMessageVsClientMessage() {
	err := errx.New(errx.CodeInternal, "Something went wrong").
		WithDebug("Failed to connect to database: timeout after 30s").
		WithMeta("db_host", "postgres.internal").
		WithMeta("query", "SELECT * FROM users WHERE id = 123")

	// Client message should be safe
	clientMsg := err.Error()
	s.Equal("Something went wrong", clientMsg)
	s.NotContains(clientMsg, "postgres.internal")
	s.NotContains(clientMsg, "SELECT")

	// Debug message should contain details
	debugMsg := err.DebugMessage()
	s.Contains(debugMsg, "Something went wrong")
	s.Contains(debugMsg, "Failed to connect to database: timeout after 30s")
	s.Contains(debugMsg, "postgres.internal")
	s.Contains(debugMsg, "SELECT * FROM users WHERE id = 123")
}

func (s *errorSuite) TestErrorChaining() {
	// Simulate a chain of errors through different layers
	dbErr := errors.New("connection refused")

	repoErr := errx.Wrap(dbErr, errx.CodeUnavailable, "database unavailable").
		WithSource("user-repository").
		WithTags("database").
		WithMeta("operation", "find_by_id")

	serviceErr := errx.Wrap(repoErr, errx.CodeNotFound, "user not found").
		WithSource("user-service").
		WithTags("business-logic").
		WithMeta("user_id", 123)

	// Verify error chain - Error() only returns the message, not the cause
	s.Equal("user not found", serviceErr.Error())

	// Verify we can unwrap to original error
	s.True(errors.Is(serviceErr, dbErr))

	// Verify source-specific info is preserved
	s.Equal("user-service", serviceErr.Source())

	// Verify we can find the repository error in the chain
	var repoErrCheck *errx.Error
	s.Require().True(errors.As(serviceErr, &repoErrCheck))

	// The first *Error in the chain is serviceErr itself
	s.Equal("user-service", repoErrCheck.Source())
}

func (s *errorSuite) TestWithDebugf() {
	err := errx.New(errx.CodeInternal, "operation failed").
		WithDebugf("failed to process item %d in batch %s", 42, "batch-123")

	debugMsg := err.DebugMessage()
	s.Contains(debugMsg, "failed to process item 42 in batch batch-123")
}

func (s *errorSuite) TestMultipleTags() {
	err := errx.New(errx.CodeInternal, "error").WithTags("tag1").WithTags("tag2", "tag3")

	s.Equal([]string{"tag1", "tag2", "tag3"}, err.Tags())
}

func (s *errorSuite) TestEmptySourceAndTags() {
	err := errx.New(errx.CodeInternal, "error")

	debugMsg := err.DebugMessage()
	s.NotContains(debugMsg, "source=")
	s.NotContains(debugMsg, "tags=")
}

func (s *errorSuite) TestMetadataIsolation() {
	err1 := errx.New(errx.CodeInternal, "error 1").WithMeta("key", "value1")
	err2 := errx.New(errx.CodeInternal, "error 2").WithMeta("key", "value2")

	s.Equal("value1", err1.Metadata()["key"])
	s.Equal("value2", err2.Metadata()["key"])
}

func (s *errorSuite) TestWithRetryable() {
	// Test default (not retryable)
	err := errx.New(errx.CodeInternal, "internal error")
	s.False(err.IsRetryable())

	// Test marking as retryable
	err = errx.New(errx.CodeUnavailable, "service unavailable").WithRetryable()
	s.True(err.IsRetryable())

	// Test chaining with other methods
	err = errx.New(errx.CodeDeadlineExceeded, "request timeout").
		WithRetryable().
		WithSource("api-gateway").
		WithMeta("attempt", 1)
	s.True(err.IsRetryable())
	s.Equal("api-gateway", err.Source())
	s.Equal(1, err.Metadata()["attempt"])

	// Test with nil error
	var nilErr *errx.Error
	s.Nil(nilErr.WithRetryable())
	s.False(nilErr.IsRetryable())
}

func (s *errorSuite) TestRetryableInDebugMessage() {
	// Test that retryable appears in debug message
	err := errx.New(errx.CodeUnavailable, "service down").
		WithRetryable().
		WithSource("payment-service")

	debugMsg := err.DebugMessage()
	s.Contains(debugMsg, "retryable=true")
	s.Contains(debugMsg, "service down")
	s.Contains(debugMsg, "payment-service")

	// Test that non-retryable doesn't include retryable in debug message
	nonRetryableErr := errx.New(errx.CodeInternal, "internal error")
	s.NotContains(nonRetryableErr.DebugMessage(), "retryable")
}

func (s *errorSuite) TestSlogIntegration() {
	cause := errors.New("database error")
	err := errx.Wrap(cause, errx.CodePermissionDenied, "access denied").
		WithSource("auth-service").
		WithTags("security", "rbac").
		WithDetail("resource", "admin-panel").
		WithMeta("user_id", "123").
		WithDebug("user missing admin role")

	// Test that LogValue returns a group value
	logValue := err.LogValue()
	s.Equal("Group", logValue.Kind().String())

	// Test retryable in LogValue
	retryableErr := errx.New(errx.CodeUnavailable, "service down").WithRetryable()
	retryableLogValue := retryableErr.LogValue()
	s.Equal("Group", retryableLogValue.Kind().String())

	// Test nil error
	var nilErr *errx.Error
	nilLogValue := nilErr.LogValue()
	// Nil error should return an empty/zero value
	s.NotEqual("Group", nilLogValue.Kind().String())
}

func (s *errorSuite) TestSlogJSONOutput() {
	var buf strings.Builder
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	s.Run("errx wrapping regular error", func() {
		buf.Reset()

		regularErr := errors.New("connection timeout")
		err := errx.Wrap(regularErr, errx.CodeUnavailable, "service unavailable").
			WithSource("payment-service").
			WithTags("external", "timeout").
			WithDetail("service", "payment").
			WithMeta("retry_count", 3).
			WithDebug("payment gateway connection timeout")

		logger.Error("request failed", "error", err)

		output := buf.String()
		s.T().Logf("JSON Output:\n%s", output)

		// Verify structure
		s.Contains(output, `"error":`)
		s.Contains(output, `"code":"unavailable"`)
		s.Contains(output, `"message":"service unavailable"`)
		s.Contains(output, `"source":"payment-service"`)
		s.Contains(output, `"tags":["external","timeout"]`)
		s.Contains(output, `"details":{"service":"payment"}`)
		s.Contains(output, `"metadata":{"retry_count":3}`)
		s.Contains(output, `"debug":"payment gateway connection timeout"`)
		s.Contains(output, `"cause":"connection timeout"`)
	})

	s.Run("errx wrapping errx", func() {
		buf.Reset()

		// Inner error from repository layer
		innerErr := errx.New(errx.CodeInternal, "query execution failed").
			WithSource("user-repository").
			WithTags("database", "postgres").
			WithDetail("table", "users").
			WithMeta("query", "SELECT * FROM users WHERE id = ?").
			WithMeta("db_host", "postgres.internal").
			WithDebug("deadlock detected on users table")

		// Outer error from service layer
		outerErr := errx.Wrap(innerErr, errx.CodeNotFound, "user not found").
			WithSource("user-service").
			WithTags("business-logic").
			WithDetail("user_id", "12345").
			WithMeta("request_id", "req-789").
			WithDebug("failed to retrieve user from database")

		logger.Error("operation failed", "error", outerErr)

		output := buf.String()
		s.T().Logf("JSON Output:\n%s", output)

		// Verify outer error structure
		s.Contains(output, `"code":"not_found"`)
		s.Contains(output, `"message":"user not found"`)
		s.Contains(output, `"source":"user-service"`)
		s.Contains(output, `"tags":["business-logic"]`)
		s.Contains(output, `"details":{"user_id":"12345"}`)
		s.Contains(output, `"metadata":{"request_id":"req-789"}`)
		s.Contains(output, `"debug":"failed to retrieve user from database"`)

		// Verify nested inner error in cause field
		s.Contains(output, `"cause":`)
		s.Contains(output, `"code":"internal"`)
		s.Contains(output, `"message":"query execution failed"`)
		s.Contains(output, `"source":"user-repository"`)
		s.Contains(output, `"tags":["database","postgres"]`)
		s.Contains(output, `"table":"users"`)
		s.Contains(output, `"db_host":"postgres.internal"`)
		s.Contains(output, `"deadlock detected on users table"`)
	})

	s.Run("three level errx wrapping", func() {
		buf.Reset()

		// Level 1: Database driver error
		dbErr := errx.New(errx.CodeInternal, "connection refused").
			WithSource("postgres-driver").
			WithMeta("port", 5432)

		// Level 2: Repository layer error
		repoErr := errx.Wrap(dbErr, errx.CodeUnavailable, "database unavailable").
			WithSource("repository").
			WithMeta("operation", "findByID")

		// Level 3: Service layer error
		serviceErr := errx.Wrap(repoErr, errx.CodeNotFound, "resource not found").
			WithSource("service").
			WithMeta("resource_type", "user")

		logger.Error("request failed", "error", serviceErr)

		output := buf.String()
		s.T().Logf("JSON Output (3 levels):\n%s", output)

		// All three layers should be present in nested structure
		s.Contains(output, `"source":"service"`)
		s.Contains(output, `"source":"repository"`)
		s.Contains(output, `"source":"postgres-driver"`)
		s.Contains(output, `"resource_type":"user"`)
		s.Contains(output, `"operation":"findByID"`)
		s.Contains(output, `"port":5432`)
	})
}

func (s *errorSuite) TestErrxWrappingErrx() {
	// Create an inner errx.Error (e.g., from a repository layer)
	innerErr := errx.New(errx.CodeUnavailable, "database connection failed").
		WithSource("user-repository").
		WithTags("database", "postgres").
		WithDetail("retry_after", "30s").
		WithMeta("db_host", "postgres.internal").
		WithMeta("connection_pool_size", 10).
		WithDebug("connection pool exhausted after 3 retries")

	// Create an outer errx.Error that wraps the inner one (e.g., from a service layer)
	outerErr := errx.Wrap(innerErr, errx.CodeNotFound, "user not found").
		WithSource("user-service").
		WithTags("business-logic").
		WithDetail("user_id", "123").
		WithMeta("request_id", "req-456").
		WithDebug("failed to fetch user from repository")

	// 1. Error() - only returns outer error's message
	s.Equal(
		"user not found",
		outerErr.Error(),
		"Error() should only return outer error's message")

	// 2. Message() - only returns outer error's message
	s.Equal(
		"user not found",
		outerErr.Error(),
		"Message() should only return outer error's message")

	// 3. Code() - only returns outer error's code
	s.Equal(
		errx.CodeNotFound,
		outerErr.Code(),
		"Code() should only return outer error's code")

	// 4. Source() - only returns outer error's source
	s.Equal(
		"user-service",
		outerErr.Source(),
		"Source() should only return outer error's source")

	// 5. Tags() - only returns outer error's tags
	s.Equal(
		[]string{"business-logic"},
		outerErr.Tags(),
		"Tags() should only return outer error's tags")

	// 6. Details() - only returns outer error's details
	s.Equal(
		"123",
		outerErr.Details()["user_id"],
		"Details() should only return outer error's details")
	s.Nil(
		outerErr.Details()["retry_after"],
		"Details() should not include inner error's details")

	// 7. Metadata() - only returns outer error's metadata
	s.Equal(
		"req-456",
		outerErr.Metadata()["request_id"],
		"Metadata() should only return outer error's metadata")
	s.Nil(
		outerErr.Metadata()["db_host"],
		"Metadata() should not include inner error's metadata")

	// 8. Unwrap() - returns the inner errx.Error
	unwrapped := outerErr.Unwrap()
	s.Require().NotNil(unwrapped)
	var innerErrUnwrapped *errx.Error
	s.Require().True(errors.As(unwrapped, &innerErrUnwrapped), "Unwrapped error should be *errx.Error")
	s.Equal(errx.CodeUnavailable, innerErrUnwrapped.Code())
	s.Equal("user-repository", innerErrUnwrapped.Source())

	// 9. DebugMessage() - shows outer error's context + cause as string
	debugMsg := outerErr.DebugMessage()
	s.Contains(debugMsg, "[not_found] user not found")
	s.Contains(debugMsg, "source=user-service")
	s.Contains(debugMsg, "tags=[business-logic]")
	s.Contains(debugMsg, "debug=failed to fetch user from repository")
	s.Contains(debugMsg, "cause=database connection failed")
	// Note: The cause shows the inner error's Error() string, not its full DebugMessage()

	// 10. LogValue() - recursively includes both errors with full context
	logValue := outerErr.LogValue()
	s.Equal("Group", logValue.Kind().String())
	// The slog.Any("cause", e.cause) will recursively call LogValue on the inner error
	// since it also implements slog.LogValuer

	// 11. errors.Is() and errors.As() work correctly
	s.True(errors.Is(outerErr, innerErr), "errors.Is should find inner error")
	s.True(errx.CodeIn(outerErr, errx.CodeNotFound, errx.CodeUnavailable))

	var extractedInner *errx.Error
	s.Require().True(errors.As(outerErr, &extractedInner))
	// errors.As finds the first matching type, which is outerErr itself
	s.Equal("user-service", extractedInner.Source())
}

func (s *errorSuite) TestMultiLevelErrxWrapping() {
	// Demonstrate 3 layers of errx errors
	dbErr := errx.New(errx.CodeInternal, "connection refused").
		WithSource("database")

	repoErr := errx.Wrap(dbErr, errx.CodeUnavailable, "query failed").
		WithSource("repository")

	serviceErr := errx.Wrap(repoErr, errx.CodeNotFound, "user not found").
		WithSource("service")

	// Error() only returns the message
	s.Equal(
		"user not found",
		serviceErr.Error())

	// Can unwrap to each layer
	s.Equal(repoErr, serviceErr.Unwrap())
	s.Equal(dbErr, repoErr.Unwrap())

	// errors.Is works through all layers
	s.True(errors.Is(serviceErr, dbErr))

	// Only the outermost error's fields are directly accessible
	s.Equal("service", serviceErr.Source())
	s.Equal(errx.CodeNotFound, serviceErr.Code())
}

// Example test to demonstrate usage patterns
func ExampleError_clientVsDebug() {
	// Simulating a database error in production
	dbError := errors.New("pq: connection refused on host db.internal.company.com:5432")

	// Wrap with safe client message
	err := errx.Wrap(dbError, errx.CodeUnavailable, "Service temporarily unavailable").
		WithSource("payment-service").
		WithTags("database", "postgres").
		WithDetail("service", "payment").
		WithMeta("db_host", "db.internal.company.com").
		WithMeta("retry_count", 3).
		WithDebug("PostgreSQL connection pool exhausted after 3 retries")

	// What the client sees
	fmt.Println("Client sees:", err.Error())

	// What gets logged for debugging
	fmt.Println("Logs contain:", strings.Contains(err.DebugMessage(), "db.internal.company.com"))

	// Output:
	// Client sees: Service temporarily unavailable
	// Logs contain: true
}
