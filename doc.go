// Package errx provides structured error handling with standardized error codes,
// rich context for debugging, and safe client messaging.
//
// errx is an error handling package that provides:
//
//   - Standardized Error Codes: Typed error codes for consistent error classification
//   - Three-Tier Error Messages:
//   - System components: Error codes for inter-service communication
//   - System maintainers: Rich debug information with stack traces, metadata, and implementation details
//   - Clients: Safe error messages without exposing internal details
//   - Transport Agnostic: Works with HTTP, gRPC, or any other transport layer
//   - Go Ecosystem Integration: Full support for errors.Is, errors.As, and errors.Unwrap
//   - Builder API: Fluent, chainable methods for error construction
//   - Automatic Stack Traces: Captures stack traces for debugging
//   - Error Wrapping: Chain errors while preserving context
//
// # Quick Start
//
// Creating Errors:
//
//	// Simple error
//	err := errx.New(errx.CodeNotFound, "user not found")
//
//	// Formatted error
//	err := errx.Newf(errx.CodeInvalidArgument, "invalid user ID: %d", userID)
//
//	// Wrapping existing errors
//	dbErr := errors.New("connection refused")
//	err := errx.Wrap(dbErr, errx.CodeUnavailable, "database unavailable")
//
// Adding Context:
//
//	err := errx.New(errx.CodePermissionDenied, "access denied").
//	    WithDetail("resource", "admin-panel").        // Client-safe details
//	    WithSource("auth-service").                   // Set source (service/package/component)
//	    WithTags("security", "rbac").                 // Add tags for categorization
//	    WithMeta("user_id", userID).                  // Internal metadata
//	    WithDebug("user missing admin role").         // Internal debug message
//	    WithRetryable()                               // Mark as retryable operation
//
// Using Errors:
//
//	// Get error code
//	errCode := err.Code()                 // Returns errx.Code
//	codeStr := err.Code().String()        // Returns "permission_denied"
//
//	// Get client-safe data
//	clientMsg := err.Error()              // "access denied"
//	details := err.Details()              // map[string]any{"resource": "admin-panel"}
//
//	// Get internal data
//	source := err.Source()                // "auth-service"
//	tags := err.Tags()                    // []string{"security", "rbac"}
//	metadata := err.Metadata()            // map[string]any{"user_id": 123}
//
//	// Get debug message (for logging/debugging)
//	debugMsg := err.DebugMessage()        // "[permission_denied] access denied | class=access-denied | ..."
//
//	// Get stack trace
//	stackTrace := err.FormatStackTrace()  // Human-readable stack trace
//
// # Error Codes
//
// The package provides strictly defined standardized error codes. All codes are
// constants to ensure consistency:
//
//	CodeUnknown              // Unknown error (default/zero value)
//	CodeCanceled             // Operation canceled by caller
//	CodeInvalidArgument      // Request is invalid, regardless of system state
//	CodeDeadlineExceeded     // Deadline expired before operation completed
//	CodeNotFound             // Requested resource cannot be found
//	CodeAlreadyExists        // Resource already exists
//	CodePermissionDenied     // Caller isn't authorized
//	CodeResourceExhausted    // Resource exhausted (quota, storage, etc.)
//	CodeFailedPrecondition   // System isn't in required state
//	CodeAborted              // Operation aborted (concurrency issue)
//	CodeOutOfRange           // Operation attempted past valid range
//	CodeUnimplemented        // Operation not implemented/supported
//	CodeInternal             // Internal error (invariant broken)
//	CodeUnavailable          // Service temporarily unavailable
//	CodeDataLoss             // Unrecoverable data loss or corruption
//	CodeUnauthenticated      // Valid authentication credentials required
//
// These error codes align with the Connect RPC protocol specification.
//
// # Context-Based Metadata
//
// Use WithMetaContext to store request-scoped metadata in a context, then attach it to errors
// with WithMetaFromContext:
//
//	ctx = errx.WithMetaContext(ctx, "user_id", 123, "action", "delete")
//
//	// Later, when creating an error:
//	err := errx.New(errx.CodeNotFound, "user not found").WithMetaFromContext(ctx)
//	// err.Metadata() == map[string]any{"user_id": 123, "action": "delete"}
//
// Metadata accumulates across calls:
//
//	ctx = errx.WithMetaContext(ctx, "user_id", 123)
//	ctx = errx.WithMetaContext(ctx, "action", "delete")
//	// Both "user_id" and "action" are present
//
// WithMetaFromContext uses last-write-wins: if the same key was set via WithMeta, the
// context value takes precedence. Reverse the call order to give WithMeta priority.
//
// # Ensure Functions
//
// Use Ensure and Ensuref to guarantee an error is an *Error without clobbering
// the original code. If the error is already an *Error (or wraps one), it is
// returned unchanged. Otherwise, it is wrapped with the given fallback code and message:
//
//	// At a service boundary — preserves not_found, only wraps unknown errors as internal
//	return errx.Ensure(err, errx.CodeInternal, "unexpected error")
//
//	// With formatting
//	return errx.Ensuref(err, errx.CodeInternal, "unexpected error in %s", "user-service")
//
// Convenience variants exist for each code:
//
//	return errx.EnsureInternal(err, "unexpected error")
//	return errx.EnsurefInternal(err, "unexpected error in %s", "user-service")
//
// # Convenience Functions
//
// For each error code, the package provides convenience constructors:
//
//	errx.New{Code}(msg)                    // e.g., errx.NewNotFound("user not found")
//	errx.Newf{Code}(format, args...)       // e.g., errx.NewfNotFound("user %d not found", id)
//	errx.Wrap{Code}(err, msg)              // e.g., errx.WrapInternal(err, "db failed")
//	errx.Wrapf{Code}(err, format, args...) // e.g., errx.WrapfInternal(err, "db %s failed", name)
//	errx.Ensure{Code}(err, msg)            // e.g., errx.EnsureInternal(err, "unexpected")
//	errx.Ensuref{Code}(err, format, args...) // e.g., errx.EnsurefInternal(err, "unexpected in %s", svc)
//
// # Usage Patterns
//
// Layered Error Handling:
//
//	// Data layer
//	func (r *UserRepository) FindByID(id int) (*User, error) {
//	    user, err := r.db.Query("SELECT * FROM users WHERE id = ?", id)
//	    if err != nil {
//	        return nil, errx.Wrap(err, errx.CodeInternal, "database query failed").
//	            WithSource("user-repository").
//	            WithTags("database").
//	            WithMeta("user_id", id).
//	            WithDebugf("query failed: %v", err)
//	    }
//
//	    if user == nil {
//	        return nil, errx.New(errx.CodeNotFound, "user not found").
//	            WithSource("user-repository").
//	            WithMeta("user_id", id)
//	    }
//
//	    return user, nil
//	}
//
//	// Service layer
//	func (s *UserService) GetUser(id int) (*User, error) {
//	    user, err := s.repo.FindByID(id)
//	    if err != nil {
//	        if e, ok := errx.As(err); ok && e.Code() != errx.CodeNotFound {
//	            return nil, err
//	        }
//	        user, err = s.repo.CreateDefaultUser(id)
//	        if err != nil {
//	            return nil, err
//	        }
//	    }
//	    return user, nil
//	}
//
// Working with Standard Errors:
//
//	// Check if error is an errx.Error
//	if errx.Is(err) {
//	    // It's an errx error
//	}
//
//	// Extract errx.Error from any error
//	if e, ok := errx.As(err); ok {
//	    code := e.Code()
//	    source := e.Source()
//	    metadata := e.Metadata()
//	}
//
//	// Check if error has a specific code
//	if errx.CodeIs(err, errx.CodeNotFound) {
//	    // Handle not found
//	}
//
//	// Check if error has any of multiple codes
//	if errx.CodeIn(err, errx.CodeNotFound, errx.CodeUnauthenticated) {
//	    // Handle client errors
//	}
//
//	// Extract error code
//	errCode := errx.CodeOf(err) // Returns CodeUnknown for non-errx errors
//
//	// Check if error is retryable
//	if errx.IsRetryable(err) {
//	    // Retry the operation
//	}
//
//	// Also compatible with standard errors package
//	err1 := errx.New(errx.CodeNotFound, "not found")
//	err2 := errx.New(errx.CodeNotFound, "different message")
//	errors.Is(err1, err2)  // true - same code
//
//	var errxErr *errx.Error
//	if errors.As(err, &errxErr) {
//	    code := errxErr.Code()
//	}
//
// # Generic Type Checking
//
// For most use cases, use the non-generic Is() and As() functions to work with *errx.Error.
// For checking other custom error types, use the generic IsType[E] and AsType[E] functions:
//
//	// Non-generic versions for errx.Error (recommended for common case)
//	if errx.Is(err) {
//	    // err is or wraps an *errx.Error
//	}
//
//	if e, ok := errx.As(err); ok {
//	    // e is an *errx.Error
//	    code := e.Code()
//	}
//
//	// Generic versions for custom error types
//	type MyCustomError struct {
//	    Reason string
//	}
//
//	func (e *MyCustomError) Error() string {
//	    return e.Reason
//	}
//
//	// Check if error is or wraps a custom type
//	if errx.IsType[*MyCustomError](err) {
//	    // err is or wraps a *MyCustomError
//	}
//
//	// Extract custom error type from error chain
//	if e, ok := errx.AsType[*MyCustomError](err); ok {
//	    // e is a *MyCustomError
//	    fmt.Println(e.Reason)
//	}
//
// # Marking Errors as Retryable
//
// Use WithRetryable() to indicate that an operation can be retried:
//
//	// Temporary service unavailability - can be retried
//	err := errx.New(errx.CodeUnavailable, "service temporarily unavailable").
//	    WithRetryable().
//	    WithSource("payment-service")
//
//	// Check if an error is retryable
//	if errx.IsRetryable(err) {
//	    // Implement retry logic (e.g., return to retry queue, exponential backoff, etc.)
//	}
//
//	// Works with wrapped errors too
//	wrappedErr := fmt.Errorf("payment failed: %w", err)
//	errx.IsRetryable(wrappedErr)  // true
//
// Common retryable scenarios:
//   - CodeUnavailable: Service temporarily down
//   - CodeDeadlineExceeded: Request timeout
//   - CodeResourceExhausted: Rate limit exceeded
//   - CodeAborted: Optimistic locking conflict
//
// Non-retryable errors typically include:
//   - CodeInvalidArgument: Bad request data
//   - CodeNotFound: Resource doesn't exist
//   - CodePermissionDenied: Authorization failure
//   - CodeUnimplemented: Feature not supported
//
// Protecting Client-Facing Messages:
//
//	// BAD: Exposing implementation details
//	err := errx.New(errx.CodeInternal, "failed to connect to postgres.internal.company.com:5432")
//
//	// GOOD: Safe client message + debug details
//	err := errx.New(errx.CodeInternal, "service temporarily unavailable").
//	    WithDebug("failed to connect to postgres.internal.company.com:5432").
//	    WithMeta("db_query", "SELECT * FROM users WHERE id = ?")
//
//	// Client sees: "service temporarily unavailable" (Error())
//	// Logs contain: full debug message with all details
//
// # Client vs. Internal Data
//
// The error package separates data into two categories:
//
// Client-Exposed Data (Safe to Return to End Users):
//
//   - Error(): Human-readable error message safe for clients (standard error interface)
//   - Details(): Key-value pairs safe to expose (e.g., {"resource": "admin-panel"})
//
// Internal-Only Data (For Debugging/Logging):
//
//   - Source(): Source (service/package/component) where error occurred (e.g., "user-service")
//   - Tags(): Categorization tags (e.g., ["database", "critical"])
//   - Metadata(): Debug key-value pairs (e.g., {"db_host": "postgres.internal", "user_id": 123})
//   - DebugMessage(): Full debug message with all context
//
// # Best Practices
//
// 1. Use Safe Client Messages: Never expose implementation details in the error message
//
//	// Bad
//	err := errx.New(errx.CodeInternal, "SQL: connection to db.internal failed")
//
//	// Good
//	err := errx.New(errx.CodeInternal, "service unavailable").
//	    WithDebug("SQL: connection to db.internal failed")
//
// 2. Add Context at Each Layer: Each layer should add relevant context
//
//	// Repository layer
//	err = errx.Wrap(err, code, msg).WithSource("user-repo").WithTags("database")
//
//	// Service layer
//	err = errx.Wrap(err, code, msg).WithSource("user-service").WithMeta("user_id", id)
//
// 3. Use Appropriate Error Codes: Choose codes that accurately represent the error
//
//	// Not found vs. permission denied
//	if user == nil {
//	    return errx.New(errx.CodeNotFound, "user not found")  // User doesn't exist
//	}
//	if !canAccess {
//	    return errx.New(errx.CodePermissionDenied, "access denied")  // User exists but no permission
//	}
//
// 4. Leverage Metadata for Debugging: Add useful debugging information
//
//	err.WithMeta("user_id", id).
//	    WithMeta("action", "delete").
//	    WithMeta("retry_count", retries)
//
// 5. Don't Create Empty Error Chains: Only wrap when you have context to add
//
//	// Bad - adds no value
//	return errx.Wrap(err, errx.CodeInternal, err.Error())
//
//	// Good - adds context
//	return errx.Wrap(err, errx.CodeInternal, "failed to process payment").
//	    WithMeta("payment_id", paymentID)
//
// # Testing
//
// Example test using errx errors:
//
//	func TestUserService_GetUser_NotFound(t *testing.T) {
//	    // ... setup
//
//	    user, err := service.GetUser(999)
//
//	    require.Error(t, err)
//	    assert.Nil(t, user)
//
//	    // Check if it's an errx error
//	    assert.True(t, errx.Is(err))
//
//	    // Extract and verify details
//	    e, ok := errx.As(err)
//	    require.True(t, ok)
//	    assert.Equal(t, errx.CodeNotFound, e.Code())
//	    assert.Equal(t, "user-service", e.Source())
//
//	    // Or use code helpers
//	    assert.True(t, errx.CodeIs(err, errx.CodeNotFound))
//	    assert.Equal(t, errx.CodeNotFound, errx.CodeOf(err))
//	    assert.True(t, errx.CodeIn(err, errx.CodeNotFound, errx.CodeUnauthenticated))
//	}
//
// # Educational Resources
//
// # Error Handling as Domain Design
//
// Failure is your Domain by Ben Johnson (https://web.archive.org/web/2020/https://middlemost.com/failure-is-your-domain/)
//
// This article advocates treating errors as part of your application's domain rather than
// as afterthoughts. It introduces a three-consumer model for errors:
//
// 1. Application: Needs machine-readable error codes for programmatic handling and inter-service communication
//
// 2. End User: Needs human-readable messages that are safe to display (without exposing implementation details)
//
// 3. Operator: Needs rich debugging context including logical stack traces, metadata, and internal details
//
// The errx package implements these principles with its standardized error codes (aligned with ConnectRPC),
// three-tier messaging system (Code/Error()/Details for clients, Source/Tags/Metadata/DebugMessage for operators),
// and automatic stack trace capture.
package errx
