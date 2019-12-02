package middleware

import (
	"context"
	"fmt"
	"net/http"

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

	resolverCtx := graphql.GetResolverContext(ctx)

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

func RequireAuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		authCtx := authentication.GetAuthContext(ctx)

		if authCtx.Authenticated || authCtx.IgnoreAuthenticationState {
			// Proceed if auth context is authenticated or authentication state should be ignored
			next.ServeHTTP(w, r)
		} else if authCtx.Error != nil {
			// If specific error is set in auth context, send it
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintf(w, "Authentication failure: %v", authCtx.Error)
		} else {
			// Otherwise authentication is required
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprint(w, "Authentication required")
		}
	})
}
