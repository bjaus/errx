package errx_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/bjaus/errx"
)

type errxSuite struct {
	suite.Suite
}

type customError struct {
	message string
	code    int
}

func (e *customError) Error() string {
	return e.message
}

func TestErrxSuite(t *testing.T) {
	suite.Run(t, new(errxSuite))
}

func (s *errxSuite) TestNew() {
	err := errx.New(errx.CodeNotFound, "user not found")

	s.Equal(errx.CodeNotFound, err.Code())
	s.Equal("user not found", err.Error())
	s.Equal("user not found", err.Error())
	s.NotEmpty(err.StackTrace())
}

func (s *errxSuite) TestNewf() {
	err := errx.Newf(errx.CodeInvalidArgument, "invalid user ID: %d", 123)

	s.Equal(errx.CodeInvalidArgument, err.Code())
	s.Equal("invalid user ID: 123", err.Error())
}

func (s *errxSuite) TestWrap() {
	originalErr := errors.New("database connection failed")
	err := errx.Wrap(originalErr, errx.CodeInternal, "failed to fetch user")

	s.Equal(errx.CodeInternal, err.Code())
	s.Equal("failed to fetch user", err.Error())
	s.Equal("failed to fetch user", err.Error())
	s.Equal(originalErr, err.Unwrap())
}

func (s *errxSuite) TestWrapNil() {
	err := errx.Wrap(nil, errx.CodeInternal, "should be nil")
	s.Nil(err)
}

func (s *errxSuite) TestWrapf() {
	originalErr := errors.New("network timeout")
	err := errx.Wrapf(originalErr, errx.CodeUnavailable, "service %s unavailable", "user-service")

	s.Equal("service user-service unavailable", err.Error())
}

func (s *errxSuite) TestWrapfNil() {
	err := errx.Wrapf(nil, errx.CodeInternal, "should be nil")
	s.Nil(err)
}

func (s *errxSuite) TestEnsure_NilError() {
	result := errx.Ensure(nil, errx.CodeInternal, "fallback")
	s.Nil(result)
}

func (s *errxSuite) TestEnsure_AlreadyErrxError() {
	original := errx.New(errx.CodeNotFound, "not found")
	result := errx.Ensure(original, errx.CodeInternal, "fallback")

	// Should return the same pointer, preserving the original code
	s.Same(original, result)
	s.Equal(errx.CodeNotFound, result.Code())
	s.Equal("not found", result.Error())
}

func (s *errxSuite) TestEnsure_StandardError() {
	stdErr := errors.New("something broke")
	result := errx.Ensure(stdErr, errx.CodeInternal, "internal error")

	s.Require().NotNil(result)
	s.Equal(errx.CodeInternal, result.Code())
	s.Equal("internal error", result.Error())
	s.Equal(stdErr, result.Unwrap())
}

func (s *errxSuite) TestEnsure_WrappedErrxError() {
	original := errx.New(errx.CodeNotFound, "not found")
	wrapped := fmt.Errorf("outer: %w", original)

	result := errx.Ensure(wrapped, errx.CodeInternal, "fallback")

	// Should extract and return the inner *errx.Error
	s.Same(original, result)
	s.Equal(errx.CodeNotFound, result.Code())
}

func (s *errxSuite) TestEnsuref_NilError() {
	result := errx.Ensuref(nil, errx.CodeInternal, "fallback %s", "msg")
	s.Nil(result)
}

func (s *errxSuite) TestEnsuref_AlreadyErrxError() {
	original := errx.New(errx.CodeNotFound, "not found")
	result := errx.Ensuref(original, errx.CodeInternal, "fallback %s", "msg")

	s.Same(original, result)
	s.Equal(errx.CodeNotFound, result.Code())
}

func (s *errxSuite) TestEnsuref_StandardError() {
	stdErr := errors.New("something broke")
	result := errx.Ensuref(stdErr, errx.CodeInternal, "internal error: %s", "details")

	s.Require().NotNil(result)
	s.Equal(errx.CodeInternal, result.Code())
	s.Equal("internal error: details", result.Error())
	s.Equal(stdErr, result.Unwrap())
}

func (s *errxSuite) TestCodeIs() {
	err := errx.New(errx.CodeNotFound, "not found")

	s.True(errx.CodeIs(err, errx.CodeNotFound))
	s.False(errx.CodeIs(err, errx.CodeInternal))

	// Test with wrapped error
	wrappedErr := fmt.Errorf("outer: %w", err)
	s.True(errx.CodeIs(wrappedErr, errx.CodeNotFound))

	// Test with non-errx error
	stdErr := errors.New("standard error")
	s.False(errx.CodeIs(stdErr, errx.CodeNotFound))
}

func (s *errxSuite) TestCodeOf() {
	err := errx.New(errx.CodePermissionDenied, "access denied")
	s.Equal(errx.CodePermissionDenied, errx.CodeOf(err))

	// Test with standard error
	stdErr := errors.New("standard error")
	s.Equal(errx.CodeUnknown, errx.CodeOf(stdErr))

	// Test with wrapped error
	wrappedErr := fmt.Errorf("wrapped: %w", err)
	s.Equal(errx.CodePermissionDenied, errx.CodeOf(wrappedErr))
}

func (s *errxSuite) TestCodeIn() {
	err := errx.New(errx.CodeNotFound, "not found")

	// Test single code match
	s.True(errx.CodeIn(err, errx.CodeNotFound))

	// Test multiple codes with match
	s.True(errx.CodeIn(err, errx.CodeInternal, errx.CodeNotFound, errx.CodeUnavailable))

	// Test multiple codes without match
	s.False(errx.CodeIn(err, errx.CodeInternal, errx.CodeUnavailable))

	// Test with no codes provided
	s.False(errx.CodeIn(err))

	// Test with standard error
	stdErr := errors.New("standard error")
	s.False(errx.CodeIn(stdErr, errx.CodeNotFound))

	// Test with wrapped error
	wrappedErr := fmt.Errorf("outer: %w", err)
	s.True(errx.CodeIn(wrappedErr, errx.CodeNotFound, errx.CodeInternal))
}

func (s *errxSuite) TestIs() {
	// Test with errx.Error
	err := errx.New(errx.CodeNotFound, "not found")
	s.True(errx.Is(err))

	// Test with standard error
	stdErr := errors.New("standard error")
	s.False(errx.Is(stdErr))

	// Test with wrapped errx.Error
	wrappedErr := fmt.Errorf("outer: %w", err)
	s.True(errx.Is(wrappedErr))

	// Test with nil
	s.False(errx.Is(nil))
}

func (s *errxSuite) TestAs() {
	// Test with errx.Error
	err := errx.New(errx.CodeNotFound, "not found")
	e, ok := errx.As(err)
	s.True(ok)
	s.Equal(errx.CodeNotFound, e.Code())
	s.Equal("not found", e.Error())

	// Test with standard error
	stdErr := errors.New("standard error")
	e, ok = errx.As(stdErr)
	s.False(ok)
	s.Nil(e)

	// Test with wrapped errx.Error
	wrappedErr := fmt.Errorf("outer: %w", err)
	e, ok = errx.As(wrappedErr)
	s.True(ok)
	s.Equal(errx.CodeNotFound, e.Code())

	// Test with nil
	e, ok = errx.As(nil)
	s.False(ok)
	s.Nil(e)

	// Test with wrapped errx.Error with metadata
	richErr := errx.New(errx.CodePermissionDenied, "access denied").
		WithSource("auth-service").
		WithMeta("user_id", "123")
	wrappedRichErr := fmt.Errorf("auth failed: %w", richErr)
	e, ok = errx.As(wrappedRichErr)
	s.True(ok)
	s.Equal("auth-service", e.Source())
	s.Equal("123", e.Metadata()["user_id"])
}

func (s *errxSuite) TestIsType() {
	customErr := &customError{message: "custom error"}

	// Test with matching custom error type
	s.True(errx.IsType[*customError](customErr))

	// Test with wrapped custom error
	wrappedCustom := fmt.Errorf("wrapper: %w", customErr)
	s.True(errx.IsType[*customError](wrappedCustom))

	// Test with non-matching error type
	stdErr := errors.New("standard error")
	s.False(errx.IsType[*customError](stdErr))

	// Test with nil
	s.False(errx.IsType[*customError](nil))

	// Test with errx.Error
	errxErr := errx.New(errx.CodeNotFound, "not found")
	s.True(errx.IsType[*errx.Error](errxErr))

	// Test with wrapped errx.Error
	wrappedErrx := fmt.Errorf("outer: %w", errxErr)
	s.True(errx.IsType[*errx.Error](wrappedErrx))

	// Test with standard error (should not match *errx.Error)
	s.False(errx.IsType[*errx.Error](stdErr))
}

func (s *errxSuite) TestAsType() {
	customErr := &customError{message: "custom error", code: 42}

	// Test with matching custom error type
	e, ok := errx.AsType[*customError](customErr)
	s.True(ok)
	s.NotNil(e)
	s.Equal("custom error", e.message)
	s.Equal(42, e.code)

	// Test with wrapped custom error
	wrappedCustom := fmt.Errorf("wrapper: %w", customErr)
	e, ok = errx.AsType[*customError](wrappedCustom)
	s.True(ok)
	s.NotNil(e)
	s.Equal("custom error", e.message)
	s.Equal(42, e.code)

	// Test with non-matching error type
	stdErr := errors.New("standard error")
	e, ok = errx.AsType[*customError](stdErr)
	s.False(ok)
	s.Nil(e)

	// Test with nil
	e, ok = errx.AsType[*customError](nil)
	s.False(ok)
	s.Nil(e)

	// Test with errx.Error
	errxErr := errx.New(errx.CodeNotFound, "not found").
		WithSource("test-service").
		WithMeta("test_id", "123")

	ee, ok := errx.AsType[*errx.Error](errxErr)
	s.True(ok)
	s.NotNil(ee)
	s.Equal(errx.CodeNotFound, ee.Code())
	s.Equal("not found", ee.Error())
	s.Equal("test-service", ee.Source())
	s.Equal("123", ee.Metadata()["test_id"])

	// Test with wrapped errx.Error
	wrappedErrx := fmt.Errorf("outer: %w", errxErr)
	ee, ok = errx.AsType[*errx.Error](wrappedErrx)
	s.True(ok)
	s.NotNil(ee)
	s.Equal(errx.CodeNotFound, ee.Code())
	s.Equal("test-service", ee.Source())

	// Test with standard error (should not match *errx.Error)
	ee, ok = errx.AsType[*errx.Error](stdErr)
	s.False(ok)
	s.Nil(ee)

	// Test with multiple wrapped levels
	multiWrapped := fmt.Errorf("level2: %w", fmt.Errorf("level1: %w", customErr))
	e, ok = errx.AsType[*customError](multiWrapped)
	s.True(ok)
	s.NotNil(e)
	s.Equal("custom error", e.message)
	s.Equal(42, e.code)
}

func (s *errxSuite) TestIsRetryable() {
	// Test with retryable errx.Error
	retryableErr := errx.New(errx.CodeUnavailable, "service unavailable").WithRetryable()
	s.True(errx.IsRetryable(retryableErr))

	// Test with non-retryable errx.Error
	nonRetryableErr := errx.New(errx.CodeInvalidArgument, "invalid input")
	s.False(errx.IsRetryable(nonRetryableErr))

	// Test with wrapped retryable errx.Error
	wrappedRetryable := fmt.Errorf("operation failed: %w", retryableErr)
	s.True(errx.IsRetryable(wrappedRetryable))

	// Test with wrapped non-retryable errx.Error
	wrappedNonRetryable := fmt.Errorf("validation failed: %w", nonRetryableErr)
	s.False(errx.IsRetryable(wrappedNonRetryable))

	// Test with standard error
	stdErr := errors.New("standard error")
	s.False(errx.IsRetryable(stdErr))

	// Test with nil
	s.False(errx.IsRetryable(nil))

	// Test with multiple levels of wrapping
	deeplyWrapped := fmt.Errorf("level3: %w", fmt.Errorf("level2: %w", retryableErr))
	s.True(errx.IsRetryable(deeplyWrapped))
}
