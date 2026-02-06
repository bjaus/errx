package errx

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type internalSuite struct {
	suite.Suite
}

func TestInternalSuite(t *testing.T) {
	suite.Run(t, new(internalSuite))
}

// TestCodeNamesMap verifies the internal codeNames map is complete and correct
func (s *internalSuite) TestCodeNamesMap() {
	expectedMappings := map[Code]string{
		CodeUnknown:            "unknown",
		CodeCanceled:           "canceled",
		CodeInvalidArgument:    "invalid_argument",
		CodeDeadlineExceeded:   "deadline_exceeded",
		CodeNotFound:           "not_found",
		CodeAlreadyExists:      "already_exists",
		CodePermissionDenied:   "permission_denied",
		CodeResourceExhausted:  "resource_exhausted",
		CodeFailedPrecondition: "failed_precondition",
		CodeAborted:            "aborted",
		CodeOutOfRange:         "out_of_range",
		CodeUnimplemented:      "unimplemented",
		CodeInternal:           "internal",
		CodeUnavailable:        "unavailable",
		CodeDataLoss:           "data_loss",
		CodeUnauthenticated:    "unauthenticated",
	}

	// Verify the internal map has all expected entries
	s.Equal(len(expectedMappings), len(_CodeMap), fmt.Sprintf("codeNames map should have exactly %d entries", len(expectedMappings)))

	for code, expectedName := range expectedMappings {
		actualName, exists := _CodeMap[code]
		s.True(exists, "Code %d should exist in codeNames map", code)
		s.Equal(expectedName, actualName, "Code %d should map to %s", code, expectedName)
	}
}

// TestCaptureStackTrace verifies stack trace capture works
func (s *internalSuite) TestCaptureStackTrace() {
	trace := captureStackTrace(1)

	s.NotEmpty(trace, "Stack trace should not be empty")
	s.Greater(len(trace), 0, "Stack trace should contain frames")
}

// TestErrorInternals verifies internal error structure
func (s *internalSuite) TestErrorInternals() {
	err := New(CodeInternal, "test error")

	// Verify internal fields are set correctly
	s.Equal(CodeInternal, err.code)
	s.Equal("test error", err.message)
	s.Empty(err.debugMessage)
	s.Nil(err.cause)
	s.Empty(err.source)
	s.Empty(err.tags)
	s.NotNil(err.details)
	s.NotNil(err.metadata)
	s.NotEmpty(err.stackTrace)
}

// TestErrorWithAllFields verifies all internal fields work together
func (s *internalSuite) TestErrorWithAllFields() {
	cause := errors.New("database error")
	err := Wrap(cause, CodeUnavailable, "service down").
		WithSource("payment-service").
		WithTags("critical", "payment").
		WithDetail("user_id", "user-123").
		WithMeta("transaction_id", "tx-123").
		WithDebug("database connection pool exhausted")

	s.Equal(CodeUnavailable, err.code)
	s.Equal("service down", err.message)
	s.Equal("database connection pool exhausted", err.debugMessage)
	s.Equal(cause, err.cause)
	s.Equal("payment-service", err.source)
	s.Equal([]string{"critical", "payment"}, err.tags)
	s.Equal("user-123", err.details["user_id"])
	s.Equal("tx-123", err.metadata["transaction_id"])
	s.NotEmpty(err.stackTrace)
}
