# errx

[![Go Reference](https://pkg.go.dev/badge/github.com/bjaus/errx.svg)](https://pkg.go.dev/github.com/bjaus/errx)
[![Go Report Card](https://goreportcard.com/badge/github.com/bjaus/errx)](https://goreportcard.com/report/github.com/bjaus/errx)
[![CI](https://github.com/bjaus/errx/actions/workflows/ci.yml/badge.svg)](https://github.com/bjaus/errx/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/bjaus/errx/branch/main/graph/badge.svg)](https://codecov.io/gh/bjaus/errx)

Structured error handling for Go with standardized error codes, three-tier messaging, and seamless `slog` integration.

## Features

- **Standardized Error Codes** — 16 codes aligned with gRPC/ConnectRPC for consistent error classification
- **Three-Tier Messaging** — Separate client-safe messages from internal debug details
- **Builder API** — Fluent, chainable methods for error construction
- **Automatic Stack Traces** — Captures call stack at error creation for debugging
- **Context Metadata** — Attach request-scoped metadata via `context.Context`
- **Structured Logging** — Implements `slog.LogValuer` for rich JSON logs with nested cause chains
- **Go Ecosystem Integration** — Full support for `errors.Is`, `errors.As`, and `errors.Unwrap`
- **Zero Dependencies** — Only the standard library (plus testify for tests)

## Installation

```bash
go get github.com/bjaus/errx
```

Requires Go 1.24 or later.

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/bjaus/errx"
)

func main() {
    // Create a simple error
    err := errx.NewNotFound("user not found")
    fmt.Println(err.Code())  // not_found
    fmt.Println(err.Error()) // user not found

    // Create a rich error with context
    err = errx.New(errx.CodePermissionDenied, "access denied").
        WithSource("auth-service").
        WithDetail("resource", "admin-panel").
        WithMeta("user_id", 123).
        WithDebug("user lacks admin role")

    // Client sees only safe data
    fmt.Println(err.Error())   // access denied
    fmt.Println(err.Details()) // map[resource:admin-panel]

    // Logs contain full debug context
    fmt.Println(err.DebugMessage())
    // [permission_denied] access denied | source=auth-service | details=map[resource:admin-panel] | metadata=map[user_id:123] | debug=user lacks admin role
}
```

## Error Codes

| Code | Description |
|------|-------------|
| `CodeUnknown` | Unknown error (default/zero value) |
| `CodeCanceled` | Operation canceled by caller |
| `CodeInvalidArgument` | Request is invalid regardless of system state |
| `CodeDeadlineExceeded` | Deadline expired before operation completed |
| `CodeNotFound` | Requested resource cannot be found |
| `CodeAlreadyExists` | Resource already exists |
| `CodePermissionDenied` | Caller isn't authorized |
| `CodeResourceExhausted` | Resource exhausted (quota, storage, etc.) |
| `CodeFailedPrecondition` | System isn't in required state |
| `CodeAborted` | Operation aborted (concurrency issue) |
| `CodeOutOfRange` | Operation attempted past valid range |
| `CodeUnimplemented` | Operation not implemented/supported |
| `CodeInternal` | Internal error (invariant broken) |
| `CodeUnavailable` | Service temporarily unavailable |
| `CodeDataLoss` | Unrecoverable data loss or corruption |
| `CodeUnauthenticated` | Valid authentication credentials required |

## Usage

### Creating Errors

```go
// Simple error with code
err := errx.New(errx.CodeNotFound, "user not found")

// Formatted message
err := errx.Newf(errx.CodeInvalidArgument, "invalid ID: %d", id)

// Convenience constructors for each code
err := errx.NewNotFound("user not found")
err := errx.NewfInternal("db error: %v", dbErr)
```

### Wrapping Errors

```go
// Wrap an existing error
dbErr := sql.ErrNoRows
err := errx.Wrap(dbErr, errx.CodeNotFound, "user not found")

// Convenience wrappers
err := errx.WrapInternal(dbErr, "query failed")
err := errx.WrapfUnavailable(dbErr, "service %s down", svc)
```

### Adding Context

```go
err := errx.New(errx.CodeInternal, "operation failed").
    WithSource("payment-service").           // Service/component origin
    WithTags("database", "critical").        // Categorization tags
    WithDetail("order_id", "ord-123").       // Client-safe details
    WithMeta("query", "SELECT ...").         // Internal debug metadata
    WithDebug("connection pool exhausted").  // Internal debug message
    WithRetryable()                          // Mark as retryable
```

### Context-Based Metadata

Attach request-scoped metadata that automatically flows to errors:

```go
// In middleware
ctx = errx.WithMetaContext(ctx, "request_id", reqID, "user_id", userID)

// Later in your code
err := errx.NewInternal("operation failed").WithMetaFromContext(ctx)
// err.Metadata() contains request_id and user_id
```

### Checking Errors

```go
// Check if error is an errx.Error
if errx.Is(err) {
    // ...
}

// Extract errx.Error from chain
if e, ok := errx.As(err); ok {
    code := e.Code()
    source := e.Source()
}

// Check specific codes
if errx.CodeIs(err, errx.CodeNotFound) {
    // handle not found
}

// Check multiple codes
if errx.CodeIn(err, errx.CodeNotFound, errx.CodePermissionDenied) {
    // handle client error
}

// Extract code (returns CodeUnknown for non-errx errors)
code := errx.CodeOf(err)

// Check if retryable
if errx.IsRetryable(err) {
    // retry the operation
}
```

### Ensure Functions

Guarantee an error is an `*errx.Error` without clobbering existing codes:

```go
// At service boundaries - preserves original code if already errx
return errx.Ensure(err, errx.CodeInternal, "unexpected error")

// Convenience variants
return errx.EnsureInternal(err, "unexpected error")
```

### Structured Logging with slog

errx implements `slog.LogValuer` for rich structured logging:

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

innerErr := errx.NewInternal("query failed").
    WithSource("repository").
    WithMeta("table", "users")

outerErr := errx.Wrap(innerErr, errx.CodeNotFound, "user not found").
    WithSource("service").
    WithMeta("user_id", 123)

logger.Error("request failed", "error", outerErr)
```

Output:
```json
{
  "level": "ERROR",
  "msg": "request failed",
  "error": {
    "code": "not_found",
    "message": "user not found",
    "source": "service",
    "metadata": {"user_id": 123},
    "cause": {
      "code": "internal",
      "message": "query failed",
      "source": "repository",
      "metadata": {"table": "users"}
    }
  }
}
```

## Client vs Internal Data

errx separates data into client-safe and internal categories:

| Client-Safe | Internal-Only |
|-------------|---------------|
| `Error()` — Message | `Source()` — Origin service/component |
| `Details()` — Safe key-value pairs | `Tags()` — Categorization tags |
| `Code()` — Error code | `Metadata()` — Debug key-value pairs |
| | `DebugMessage()` — Full debug output |
| | `StackTrace()` — Call stack |

## Layered Architecture Example

```go
// Repository layer
func (r *UserRepo) FindByID(ctx context.Context, id int) (*User, error) {
    user, err := r.db.Get(ctx, id)
    if err != nil {
        return nil, errx.WrapInternal(err, "database query failed").
            WithSource("user-repository").
            WithMeta("user_id", id)
    }
    if user == nil {
        return nil, errx.NewNotFound("user not found").
            WithSource("user-repository").
            WithMeta("user_id", id)
    }
    return user, nil
}

// Service layer
func (s *UserService) GetUser(ctx context.Context, id int) (*User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, errx.Wrap(err, errx.CodeOf(err), "failed to get user").
            WithSource("user-service").
            WithMetaFromContext(ctx)
    }
    return user, nil
}

// Handler layer
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    user, err := h.svc.GetUser(r.Context(), id)
    if err != nil {
        if errx.CodeIs(err, errx.CodeNotFound) {
            http.Error(w, "User not found", http.StatusNotFound)
            return
        }
        slog.Error("request failed", "error", err)
        http.Error(w, "Internal error", http.StatusInternalServerError)
        return
    }
    // ...
}
```

## Testing

The package includes comprehensive tests with 100% coverage of core functionality:

```bash
go test -v ./...
```

## License

MIT License - see [LICENSE](LICENSE) for details.
