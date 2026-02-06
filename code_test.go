package errx_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/bjaus/errx"
)

const expectedErrorCodeCount = 16

type codeSuite struct {
	suite.Suite
}

func TestCodeSuite(t *testing.T) {
	suite.Run(t, new(codeSuite))
}

func (s *codeSuite) TestCode() {
	// Test ALL error codes to ensure comprehensive coverage
	tests := map[string]struct {
		code errx.Code
	}{
		"unknown":             {code: errx.CodeUnknown},
		"canceled":            {code: errx.CodeCanceled},
		"invalid_argument":    {code: errx.CodeInvalidArgument},
		"deadline_exceeded":   {code: errx.CodeDeadlineExceeded},
		"not_found":           {code: errx.CodeNotFound},
		"already_exists":      {code: errx.CodeAlreadyExists},
		"permission_denied":   {code: errx.CodePermissionDenied},
		"resource_exhausted":  {code: errx.CodeResourceExhausted},
		"failed_precondition": {code: errx.CodeFailedPrecondition},
		"aborted":             {code: errx.CodeAborted},
		"out_of_range":        {code: errx.CodeOutOfRange},
		"unimplemented":       {code: errx.CodeUnimplemented},
		"internal":            {code: errx.CodeInternal},
		"unavailable":         {code: errx.CodeUnavailable},
		"data_loss":           {code: errx.CodeDataLoss},
		"unauthenticated":     {code: errx.CodeUnauthenticated},
	}

	// This assertion ensures we test all defined codes
	// If someone adds/removes a code, this test will fail
	s.Equal(expectedErrorCodeCount, len(tests), "Expected %d defined error codes", expectedErrorCodeCount)

	for name, tt := range tests {
		s.Run(name, func() {
			err := errx.New(tt.code, "test error")
			s.Equal(tt.code, err.Code())
			s.Equal(name, tt.code.String())
		})
	}
}

func (s *codeSuite) TestCodeString() {
	// Test ALL error codes to ensure String() works correctly for every code
	tests := map[string]struct {
		code errx.Code
	}{
		"unknown":             {code: errx.CodeUnknown},
		"canceled":            {code: errx.CodeCanceled},
		"invalid_argument":    {code: errx.CodeInvalidArgument},
		"deadline_exceeded":   {code: errx.CodeDeadlineExceeded},
		"not_found":           {code: errx.CodeNotFound},
		"already_exists":      {code: errx.CodeAlreadyExists},
		"permission_denied":   {code: errx.CodePermissionDenied},
		"resource_exhausted":  {code: errx.CodeResourceExhausted},
		"failed_precondition": {code: errx.CodeFailedPrecondition},
		"aborted":             {code: errx.CodeAborted},
		"out_of_range":        {code: errx.CodeOutOfRange},
		"unimplemented":       {code: errx.CodeUnimplemented},
		"internal":            {code: errx.CodeInternal},
		"unavailable":         {code: errx.CodeUnavailable},
		"data_loss":           {code: errx.CodeDataLoss},
		"unauthenticated":     {code: errx.CodeUnauthenticated},
		"Code(255)":           {code: errx.Code(255)}, // Unknown codes display as Code(n)
	}

	// This assertion ensures we test all defined codes (plus 1 unknown)
	// If someone adds/removes a code, this test will fail
	s.Equal(expectedErrorCodeCount+1, len(tests), "Expected %d defined error codes + 1 unknown code test", expectedErrorCodeCount)

	for expected, tt := range tests {
		s.Run(expected, func() {
			s.Equal(expected, tt.code.String())
		})
	}
}

func (s *codeSuite) TestCodeValues() {
	// Test that all error codes have the expected numeric values
	// This ensures nobody accidentally changes the code numbering which could break compatibility
	tests := map[string]struct {
		code     errx.Code
		expected uint32
	}{
		"unknown":             {code: errx.CodeUnknown, expected: 0},
		"canceled":            {code: errx.CodeCanceled, expected: 1},
		"invalid_argument":    {code: errx.CodeInvalidArgument, expected: 2},
		"deadline_exceeded":   {code: errx.CodeDeadlineExceeded, expected: 3},
		"not_found":           {code: errx.CodeNotFound, expected: 4},
		"already_exists":      {code: errx.CodeAlreadyExists, expected: 5},
		"permission_denied":   {code: errx.CodePermissionDenied, expected: 6},
		"resource_exhausted":  {code: errx.CodeResourceExhausted, expected: 7},
		"failed_precondition": {code: errx.CodeFailedPrecondition, expected: 8},
		"aborted":             {code: errx.CodeAborted, expected: 9},
		"out_of_range":        {code: errx.CodeOutOfRange, expected: 10},
		"unimplemented":       {code: errx.CodeUnimplemented, expected: 11},
		"internal":            {code: errx.CodeInternal, expected: 12},
		"unavailable":         {code: errx.CodeUnavailable, expected: 13},
		"data_loss":           {code: errx.CodeDataLoss, expected: 14},
		"unauthenticated":     {code: errx.CodeUnauthenticated, expected: 15},
	}

	// This assertion ensures we test all defined codes
	// If someone adds/removes a code, this test will fail
	s.Equal(expectedErrorCodeCount, len(tests), "Expected %d defined error codes", expectedErrorCodeCount)

	for name, tt := range tests {
		s.Run(name, func() {
			s.Equal(tt.expected, uint32(tt.code), "Code %s should have value %d", name, tt.expected)
		})
	}
}
