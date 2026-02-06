package errx_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/bjaus/errx"
)

type contextSuite struct {
	suite.Suite
}

func TestContextSuite(t *testing.T) {
	suite.Run(t, new(contextSuite))
}

func (s *contextSuite) TestWithMetaContext_StoresMetadata() {
	ctx := errx.WithMetaContext(context.Background(), "user_id", 123, "action", "delete")

	err := errx.New(errx.CodeInternal, "something failed").WithMetaFromContext(ctx)

	s.Equal(123, err.Metadata()["user_id"])
	s.Equal("delete", err.Metadata()["action"])
}

func (s *contextSuite) TestWithMetaContext_AccumulatesMetadata() {
	ctx := errx.WithMetaContext(context.Background(), "user_id", 123)
	ctx = errx.WithMetaContext(ctx, "action", "delete")

	err := errx.New(errx.CodeInternal, "something failed").WithMetaFromContext(ctx)

	s.Equal(123, err.Metadata()["user_id"])
	s.Equal("delete", err.Metadata()["action"])
}

func (s *contextSuite) TestWithMetaContext_OverwritesSameKey() {
	ctx := errx.WithMetaContext(context.Background(), "user_id", 123)
	ctx = errx.WithMetaContext(ctx, "user_id", 456)

	err := errx.New(errx.CodeInternal, "something failed").WithMetaFromContext(ctx)

	s.Equal(456, err.Metadata()["user_id"])
}

func (s *contextSuite) TestWithMetaFromContext_NilContext() {
	err := errx.New(errx.CodeInternal, "something failed").WithMetaFromContext(nil) //nolint:staticcheck // testing nil context

	s.NotNil(err)
	s.Empty(err.Metadata())
}

func (s *contextSuite) TestWithMetaFromContext_EmptyContext() {
	err := errx.New(errx.CodeInternal, "something failed").WithMetaFromContext(context.Background())

	s.NotNil(err)
	s.Empty(err.Metadata())
}

func (s *contextSuite) TestWithMetaFromContext_NilError() {
	ctx := errx.WithMetaContext(context.Background(), "user_id", 123)

	var e *errx.Error
	result := e.WithMetaFromContext(ctx)

	s.Nil(result)
}

func (s *contextSuite) TestWithMetaFromContext_OverwritesExistingMeta() {
	ctx := errx.WithMetaContext(context.Background(), "user_id", 123, "source", "context")

	err := errx.New(errx.CodeInternal, "something failed").
		WithMeta("user_id", 999).
		WithMetaFromContext(ctx)

	// Context metadata overwrites existing keys (last-write-wins)
	s.Equal(123, err.Metadata()["user_id"])
	// New key from context should be added
	s.Equal("context", err.Metadata()["source"])
}

func (s *contextSuite) TestWithMeta_OverwritesContextMeta() {
	ctx := errx.WithMetaContext(context.Background(), "user_id", 123)

	err := errx.New(errx.CodeInternal, "something failed").
		WithMetaFromContext(ctx).
		WithMeta("user_id", 999)

	// WithMeta after WithMetaFromContext overwrites the context value
	s.Equal(999, err.Metadata()["user_id"])
}

func (s *contextSuite) TestWithMetaContext_SkipsNonStringKey() {
	ctx := errx.WithMetaContext(context.Background(), 123, "value", "valid_key", "valid_value")

	err := errx.New(errx.CodeInternal, "something failed").WithMetaFromContext(ctx)

	s.Nil(err.Metadata()["value"], "non-string key pair should be skipped")
	s.Equal("valid_value", err.Metadata()["valid_key"])
}

func (s *contextSuite) TestWithMetaContext_DropsTrailingKey() {
	ctx := errx.WithMetaContext(context.Background(), "key1", "val1", "key2")

	err := errx.New(errx.CodeInternal, "something failed").WithMetaFromContext(ctx)

	s.Equal("val1", err.Metadata()["key1"])
	s.Nil(err.Metadata()["key2"], "trailing key with no value should be dropped")
}
