package errx_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/bjaus/errx"
)

type codeEnumHelpersSuite struct {
	suite.Suite
}

func TestCodeEnumHelpersSuite(t *testing.T) {
	suite.Run(t, new(codeEnumHelpersSuite))
}

// Function signatures for helper variants
type (
	newFn     func(string) *errx.Error
	newfFn    func(string, ...any) *errx.Error
	wrapFn    func(error, string) *errx.Error
	wrapfFn   func(error, string, ...any) *errx.Error
	ensureFn  func(error, string) *errx.Error
	ensurefFn func(error, string, ...any) *errx.Error
)

func (s *codeEnumHelpersSuite) Test_All_New_Wrap_Wrapf() {
	tests := map[string]struct {
		code    errx.Code
		new     newFn
		newf    newfFn
		wrap    wrapFn
		wrapf   wrapfFn
		ensure  ensureFn
		ensuref ensurefFn
	}{
		"unknown":             {code: errx.CodeUnknown, new: errx.NewUnknown, newf: errx.NewfUnknown, wrap: errx.WrapUnknown, wrapf: errx.WrapfUnknown, ensure: errx.EnsureUnknown, ensuref: errx.EnsurefUnknown},
		"canceled":            {code: errx.CodeCanceled, new: errx.NewCanceled, newf: errx.NewfCanceled, wrap: errx.WrapCanceled, wrapf: errx.WrapfCanceled, ensure: errx.EnsureCanceled, ensuref: errx.EnsurefCanceled},
		"invalid_argument":    {code: errx.CodeInvalidArgument, new: errx.NewInvalidArgument, newf: errx.NewfInvalidArgument, wrap: errx.WrapInvalidArgument, wrapf: errx.WrapfInvalidArgument, ensure: errx.EnsureInvalidArgument, ensuref: errx.EnsurefInvalidArgument},
		"deadline_exceeded":   {code: errx.CodeDeadlineExceeded, new: errx.NewDeadlineExceeded, newf: errx.NewfDeadlineExceeded, wrap: errx.WrapDeadlineExceeded, wrapf: errx.WrapfDeadlineExceeded, ensure: errx.EnsureDeadlineExceeded, ensuref: errx.EnsurefDeadlineExceeded},
		"not_found":           {code: errx.CodeNotFound, new: errx.NewNotFound, newf: errx.NewfNotFound, wrap: errx.WrapNotFound, wrapf: errx.WrapfNotFound, ensure: errx.EnsureNotFound, ensuref: errx.EnsurefNotFound},
		"already_exists":      {code: errx.CodeAlreadyExists, new: errx.NewAlreadyExists, newf: errx.NewfAlreadyExists, wrap: errx.WrapAlreadyExists, wrapf: errx.WrapfAlreadyExists, ensure: errx.EnsureAlreadyExists, ensuref: errx.EnsurefAlreadyExists},
		"permission_denied":   {code: errx.CodePermissionDenied, new: errx.NewPermissionDenied, newf: errx.NewfPermissionDenied, wrap: errx.WrapPermissionDenied, wrapf: errx.WrapfPermissionDenied, ensure: errx.EnsurePermissionDenied, ensuref: errx.EnsurefPermissionDenied},
		"resource_exhausted":  {code: errx.CodeResourceExhausted, new: errx.NewResourceExhausted, newf: errx.NewfResourceExhausted, wrap: errx.WrapResourceExhausted, wrapf: errx.WrapfResourceExhausted, ensure: errx.EnsureResourceExhausted, ensuref: errx.EnsurefResourceExhausted},
		"failed_precondition": {code: errx.CodeFailedPrecondition, new: errx.NewFailedPrecondition, newf: errx.NewfFailedPrecondition, wrap: errx.WrapFailedPrecondition, wrapf: errx.WrapfFailedPrecondition, ensure: errx.EnsureFailedPrecondition, ensuref: errx.EnsurefFailedPrecondition},
		"aborted":             {code: errx.CodeAborted, new: errx.NewAborted, newf: errx.NewfAborted, wrap: errx.WrapAborted, wrapf: errx.WrapfAborted, ensure: errx.EnsureAborted, ensuref: errx.EnsurefAborted},
		"out_of_range":        {code: errx.CodeOutOfRange, new: errx.NewOutOfRange, newf: errx.NewfOutOfRange, wrap: errx.WrapOutOfRange, wrapf: errx.WrapfOutOfRange, ensure: errx.EnsureOutOfRange, ensuref: errx.EnsurefOutOfRange},
		"unimplemented":       {code: errx.CodeUnimplemented, new: errx.NewUnimplemented, newf: errx.NewfUnimplemented, wrap: errx.WrapUnimplemented, wrapf: errx.WrapfUnimplemented, ensure: errx.EnsureUnimplemented, ensuref: errx.EnsurefUnimplemented},
		"internal":            {code: errx.CodeInternal, new: errx.NewInternal, newf: errx.NewfInternal, wrap: errx.WrapInternal, wrapf: errx.WrapfInternal, ensure: errx.EnsureInternal, ensuref: errx.EnsurefInternal},
		"unavailable":         {code: errx.CodeUnavailable, new: errx.NewUnavailable, newf: errx.NewfUnavailable, wrap: errx.WrapUnavailable, wrapf: errx.WrapfUnavailable, ensure: errx.EnsureUnavailable, ensuref: errx.EnsurefUnavailable},
		"data_loss":           {code: errx.CodeDataLoss, new: errx.NewDataLoss, newf: errx.NewfDataLoss, wrap: errx.WrapDataLoss, wrapf: errx.WrapfDataLoss, ensure: errx.EnsureDataLoss, ensuref: errx.EnsurefDataLoss},
		"unauthenticated":     {code: errx.CodeUnauthenticated, new: errx.NewUnauthenticated, newf: errx.NewfUnauthenticated, wrap: errx.WrapUnauthenticated, wrapf: errx.WrapfUnauthenticated, ensure: errx.EnsureUnauthenticated, ensuref: errx.EnsurefUnauthenticated},
	}

	// Ensure we are testing all defined codes
	s.Require().Equal(16, len(tests))

	for name, tt := range tests {
		s.Run(name, func() {
			// New
			msg := "test message"
			eNew := tt.new(msg)
			s.Require().NotNil(eNew)
			s.Equal(tt.code, eNew.Code())
			s.Equal(msg, eNew.Error())
			s.Nil(eNew.Unwrap())
			s.Greater(len(eNew.StackTrace()), 0)

			// Newf
			format := "hello %s %d"
			formatted := fmt.Sprintf(format, "world", 42)
			eNewf := tt.newf(format, "world", 42)
			s.Require().NotNil(eNewf)
			s.Equal(tt.code, eNewf.Code())
			s.Equal(formatted, eNewf.Error())
			s.Nil(eNewf.Unwrap())
			s.Greater(len(eNewf.StackTrace()), 0)

			// Wrap
			cause := errors.New("root cause")
			eWrap := tt.wrap(cause, msg)
			s.Require().NotNil(eWrap)
			s.Equal(tt.code, eWrap.Code())
			s.Equal(msg, eWrap.Error())
			s.Equal(cause, eWrap.Unwrap())
			s.Greater(len(eWrap.StackTrace()), 0)

			// Wrapf
			eWrapf := tt.wrapf(cause, format, "world", 42)
			s.Require().NotNil(eWrapf)
			s.Equal(tt.code, eWrapf.Code())
			s.Equal(formatted, eWrapf.Error())
			s.Equal(cause, eWrapf.Unwrap())
			s.Greater(len(eWrapf.StackTrace()), 0)

			// Ensure with standard error (should wrap with this code)
			stdErr := errors.New("standard error")
			eEnsure := tt.ensure(stdErr, msg)
			s.Require().NotNil(eEnsure)
			s.Equal(tt.code, eEnsure.Code())
			s.Equal(msg, eEnsure.Error())
			s.Equal(stdErr, eEnsure.Unwrap())

			// Ensure with existing errx error (should passthrough)
			existingErr := errx.New(errx.CodeNotFound, "existing")
			eEnsurePass := tt.ensure(existingErr, msg)
			s.Same(existingErr, eEnsurePass)

			// Ensuref with standard error
			eEnsuref := tt.ensuref(stdErr, format, "world", 42)
			s.Require().NotNil(eEnsuref)
			s.Equal(tt.code, eEnsuref.Code())
			s.Equal(formatted, eEnsuref.Error())
			s.Equal(stdErr, eEnsuref.Unwrap())

			// Ensuref with existing errx error (should passthrough)
			eEnsurefPass := tt.ensuref(existingErr, format, "world", 42)
			s.Same(existingErr, eEnsurefPass)

			// Ensure with nil error (should return nil)
			s.Nil(tt.ensure(nil, msg))

			// Ensuref with nil error (should return nil)
			s.Nil(tt.ensuref(nil, format, "world", 42))
		})
	}
}
