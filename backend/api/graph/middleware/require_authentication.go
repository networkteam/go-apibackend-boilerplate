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

func RequireAuthenticationFieldMiddleware(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
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

	if authCtx.Authenticated || authCtx.IgnoreAuthenticationState {
		// Proceed if auth context is authenticated or the bypass directive was present
		return next(ctx)
	} else if authCtx.Error != nil {
		// If specific error is set in auth context, return it
		return nil, authCtx.Error
	} else {
		// Otherwise authentication is required
		return nil, api.ErrAuthenticationRequired
	}
}
