package middleware_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/networkteam/slogutils"
	"github.com/stretchr/testify/assert"
	"github.com/thejerf/slogassert"
	"github.com/vektah/gqlparser/v2/ast"

	"myvendor.mytld/myproject/backend/api"
	"myvendor.mytld/myproject/backend/api/graph/common/middleware"
	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/security/authorization"
)

type loggerMockResolver struct {
	returnValue any
	returnErr   error
}

func (m *loggerMockResolver) Resolve(ctx context.Context) (any, error) {
	return m.returnValue, m.returnErr
}

func setupLoggerTestContext(t *testing.T) (context.Context, *slogassert.Handler) {
	st := slogassert.NewDefault(t, slogassert.WithLeveler(slog.LevelDebug))
	logger := slog.New(st)
	ctx := slogutils.WithLogger(context.Background(), logger)
	return ctx, st
}

func setupGraphQLFieldContext(ctx context.Context, fieldName, objectType string, parent *graphql.FieldContext) context.Context {
	fieldCtx := &graphql.FieldContext{
		Field: graphql.CollectedField{
			Field: &ast.Field{
				Name: fieldName,
			},
		},
		Object: objectType,
		Parent: parent,
	}
	return graphql.WithFieldContext(ctx, fieldCtx)
}

type fieldConfig struct {
	name   string
	object string
	parent *fieldConfig
}

func setupNestedFieldContext(ctx context.Context, config *fieldConfig) context.Context {
	if config.parent != nil {
		ctx = setupNestedFieldContext(ctx, config.parent)
	}
	return setupGraphQLFieldContext(ctx, config.name, config.object, nil)
}

func TestLoggerFieldMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		field          *fieldConfig
		resolverValue  any
		resolverErr    error
		cancelContext  bool
		expectLogLevel slog.Level
		expectLogged   bool
	}{
		{
			name:           "query success - should log at debug level",
			field:          &fieldConfig{name: "users", object: "Query"},
			resolverValue:  "test result",
			resolverErr:    nil,
			expectLogLevel: slog.LevelDebug,
			expectLogged:   true,
		},
		{
			name:           "mutation success - should log at info level",
			field:          &fieldConfig{name: "createUser", object: "Mutation"},
			resolverValue:  "created",
			resolverErr:    nil,
			expectLogLevel: slog.LevelInfo,
			expectLogged:   true,
		},
		{
			name:           "typed error - should log at warn level",
			field:          &fieldConfig{name: "users", object: "Query"},
			resolverErr:    api.ErrAuthenticationFailed,
			expectLogLevel: slog.LevelWarn,
			expectLogged:   true,
		},
		{
			name:           "fields error - should log at warn level",
			field:          &fieldConfig{name: "createUser", object: "Mutation"},
			resolverErr:    types.FieldError{Field: "email", Code: "invalid"},
			expectLogLevel: slog.LevelWarn,
			expectLogged:   true,
		},
		{
			name:           "authorization error - should log at warn level",
			field:          &fieldConfig{name: "users", object: "Query"},
			resolverErr:    authorization.Fail("unauthorized"),
			expectLogLevel: slog.LevelWarn,
			expectLogged:   true,
		},
		{
			name:           "unexpected error - should log at error level",
			field:          &fieldConfig{name: "users", object: "Query"},
			resolverErr:    errors.New("database connection failed"),
			expectLogLevel: slog.LevelError,
			expectLogged:   true,
		},
		{
			name:         "__schema field - should not log",
			field:        &fieldConfig{name: "__schema", object: "Query"},
			expectLogged: false,
		},
		{
			name:           "context canceled - resolver context was canceled",
			field:          &fieldConfig{name: "users", object: "Query"},
			resolverErr:    context.Canceled,
			cancelContext:  true,
			expectLogLevel: slog.LevelDebug,
			expectLogged:   true,
		},
		{
			name:           "context canceled error - but resolver context not canceled",
			field:          &fieldConfig{name: "users", object: "Query"},
			resolverErr:    context.Canceled,
			cancelContext:  false,
			expectLogLevel: slog.LevelError,
			expectLogged:   true,
		},
		{
			name: "nested field - should not log",
			field: &fieldConfig{
				name:   "name",
				object: "String",
				parent: &fieldConfig{
					name:   "profile",
					object: "User",
					parent: &fieldConfig{
						name:   "users",
						object: "Query",
					},
				},
			},
			resolverValue: "John",
			expectLogged:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, st := setupLoggerTestContext(t)
			ctx = setupNestedFieldContext(ctx, tt.field)

			if tt.cancelContext {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			resolver := &loggerMockResolver{
				returnValue: tt.resolverValue,
				returnErr:   tt.resolverErr,
			}

			res, err := middleware.LoggerFieldMiddleware(ctx, resolver.Resolve)

			assert.Equal(t, tt.resolverErr, err, "error should be returned as-is")
			if tt.resolverErr == nil {
				assert.Equal(t, tt.resolverValue, res, "result should be returned as-is")
			}

			if tt.expectLogged {
				attrs := map[string]any{
					"component": "graphql",
					"field":     tt.field.name,
					"type":      tt.field.object,
				}

				st.AssertSomePrecise(slogassert.LogMessageMatch{
					Message: tt.field.object,
					Level:   tt.expectLogLevel,
					Attrs:   attrs,
				})
			} else {
				st.AssertEmpty()
			}
		})
	}
}
