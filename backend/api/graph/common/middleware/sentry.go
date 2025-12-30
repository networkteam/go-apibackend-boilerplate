package middleware

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	"github.com/getsentry/sentry-go"
	"github.com/networkteam/slogutils"
	sloghttp "github.com/samber/slog-http"

	"myvendor.mytld/myproject/backend/domain/types"
)

func SentryGraphqlMiddleware(ctx context.Context, next graphql.Resolver) (res any, err error) {
	fieldCtx := graphql.GetFieldContext(ctx)

	res, err = next(ctx)
	if err != nil {
		// Skip field resolvable errors, since these are expected to occur
		var fieldErr types.FieldResolvableError
		if errors.As(err, &fieldErr) {
			return nil, err
		}

		// Skip error if ctx is cancelled
		if errors.Is(err, context.Canceled) {
			// Check if ctx was cancelled to avoid ignoring errors that were cancelled inside the resolver
			select {
			case <-ctx.Done():
				return nil, err
			default:
			}
		}

		logger := slogutils.FromContext(ctx)

		hub := sentry.GetHubFromContext(ctx)
		if hub == nil {
			hub = sentry.CurrentHub()
		}

		hub.WithScope(func(scope *sentry.Scope) {
			scope.SetTag("section", "graphql")
			scope.SetExtras(map[string]any{
				"Field":      fieldCtx.Field.Name,
				"Type":       fieldCtx.Object,
				"Request ID": sloghttp.GetRequestIDFromContext(ctx),
			})

			var eventID string
			eID := hub.CaptureException(err)
			if eID != nil {
				eventID = string(*eID)
				logger.InfoContext(ctx, "Captured error with Sentry", "sentryEventId", eventID)
			}
		})
	}

	return res, err
}
