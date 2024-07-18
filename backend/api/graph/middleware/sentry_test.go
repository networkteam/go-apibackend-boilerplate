package middleware_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"
	"github.com/vektah/gqlparser/v2/ast"

	"myvendor.mytld/myproject/backend/api/graph/middleware"
	test_sentry "myvendor.mytld/myproject/backend/test/sentry"
)

type mockResolver struct {
	returnErr error
}

func (m *mockResolver) Resolve(ctx context.Context) (any, error) {
	return nil, m.returnErr
}

func TestSentryGraphqlMiddleware_ContextCanceled(t *testing.T) {
	ctx := context.Background()

	mockHub, mockTransport := test_sentry.NewHubMock()
	_ = mockTransport
	ctx = sentry.SetHubOnContext(ctx, mockHub)
	fieldCtx := &graphql.FieldContext{
		Field: graphql.CollectedField{
			Field: &ast.Field{
				Name: "foo",
			},
		},
	}
	ctx = graphql.WithFieldContext(ctx, fieldCtx)

	ctx, cancel := context.WithCancel(ctx)
	cancel()

	testErr := fmt.Errorf("some operation failed: %w", context.Canceled)
	resolver := &mockResolver{
		returnErr: testErr,
	}

	_, err := middleware.SentryGraphqlMiddleware(ctx, resolver.Resolve)

	assert.Equal(t, err, testErr, "resolver error should be returned as is")
	assert.Empty(t, mockTransport.Events(), "no event should be sent to Sentry")
}

func TestSentryGraphqlMiddleware_CanceledErr_ResolverNotCanceled(t *testing.T) {
	ctx := context.Background()

	mockHub, mockTransport := test_sentry.NewHubMock()
	_ = mockTransport
	ctx = sentry.SetHubOnContext(ctx, mockHub)
	fieldCtx := &graphql.FieldContext{
		Field: graphql.CollectedField{
			Field: &ast.Field{
				Name: "foo",
			},
		},
	}
	ctx = graphql.WithFieldContext(ctx, fieldCtx)

	testErr := fmt.Errorf("some operation failed: %w", context.Canceled)
	resolver := &mockResolver{
		returnErr: testErr,
	}

	_, err := middleware.SentryGraphqlMiddleware(ctx, resolver.Resolve)

	assert.Equal(t, err, testErr, "resolver error should be returned as is")
	assert.NotEmpty(t, mockTransport.Events(), "event should be sent to Sentry")
}
