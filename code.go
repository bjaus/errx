//go:generate go tool github.com/abice/go-enum -t=./helper.tmpl --names --values --no-iota --noparse --nocomments

package errx

// Code represents a standardized error code.
// These codes provide a consistent error classification system that is transport-agnostic.
// Transport layers (HTTP, gRPC, etc.) can map these codes to their specific status codes.
//
//	  ENUM(
//		  unknown,              // Unknown error (default/zero value)
//			canceled,             // Operation canceled by caller
//			invalid_argument,     // Request is invalid, regardless of system state
//			deadline_exceeded,    // Deadline expired before operation completed
//			not_found,            // Requested resource cannot be found
//			already_exists,       // Resource already exists
//			permission_denied,    // Caller isn't authorized
//			resource_exhausted,   // Resource exhausted (quota, storage, etc.)
//			failed_precondition,  // System isn't in required state
//			aborted,              // Operation aborted (concurrency issue)
//			out_of_range,         // Operation attempted past valid range
//			unimplemented,        // Operation not implemented/supported
//			internal,             // Internal error (invariant broken)
//			unavailable,          // Service temporarily unavailable
//			data_loss,            // Unrecoverable data loss or corruption
//			unauthenticated,      // Valid authentication credentials required
//		)
type Code uint8
