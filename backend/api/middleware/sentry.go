package middleware

import (
	"context"
	"net/http"

	sentry "github.com/getsentry/sentry-go"

	"github.com/99designs/gqlgen/graphql"
	sentryhttp "github.com/getsentry/sentry-go/http"
)

func SentryMiddleware(next http.Handler) http.Handler {
	sentryHandler := sentryhttp.New(sentryhttp.Options{})

	return sentryHandler.HandleFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	}))
}

func SentryGraphqlMiddleware(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
	resolverCtx := graphql.GetResolverContext(ctx)

	res, err = next(ctx)

	if err != nil {
		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetTag("section", "graphql")
			scope.SetExtras(map[string]interface{}{
				"Field":      resolverCtx.Field.Name,
				"Type":       resolverCtx.Object,
				"Request ID": GetReqID(ctx),
			})

			sentry.CaptureException(err)
		})
	}

	return
}
