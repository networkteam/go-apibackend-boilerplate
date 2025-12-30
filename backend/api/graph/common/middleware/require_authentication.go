package middleware

import (
	"context"

	"github.com/99designs/gqlgen/graphql"

	"myvendor.mytld/myproject/backend/api"
	"myvendor.mytld/myproject/backend/security/authentication"
)

const (
	bypassAuthenticationDirectiveName = "bypassAuthentication"
	schemaQuery                       = "__schema"
)

func RequireAuthenticationFieldMiddleware(ctx context.Context, next graphql.Resolver) (res any, err error) {
	authCtx := authentication.GetAuthContext(ctx)

	resolverCtx := graphql.GetFieldContext(ctx)

	// Check directives
	if resolverCtx.Field.Definition != nil {
		if resolverCtx.Field.Definition.Name == schemaQuery {
			authCtx.IgnoreAuthenticationState = true
			ctx = authentication.WithAuthContext(ctx, authCtx)
		} else {
			for _, directive := range resolverCtx.Field.Definition.Directives {
				if directive != nil && directive.Name == bypassAuthenticationDirectiveName {
					authCtx.IgnoreAuthenticationState = true
					ctx = authentication.WithAuthContext(ctx, authCtx)
					break
				}
			}
		}
	}

	// Proceed if auth context is authenticated or the bypass directive was present
	if authCtx.Authenticated || authCtx.IgnoreAuthenticationState {
		return next(ctx)
	}

	// If specific error is set in auth context, return it
	if authCtx.Error != nil {
		return nil, authCtx.Error
	}

	// Otherwise authentication is required
	return nil, api.ErrAuthenticationRequired
}
