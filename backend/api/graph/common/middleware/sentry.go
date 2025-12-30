package middleware

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	logger "github.com/apex/log"
	"github.com/getsentry/sentry-go"
	apexlogutils_middleware "github.com/networkteam/apexlogutils/middleware"

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

		log := logger.FromContext(ctx)

		hub := sentry.GetHubFromContext(ctx)
		if hub == nil {
			hub = sentry.CurrentHub()
		}

		hub.WithScope(func(scope *sentry.Scope) {
			scope.SetTag("section", "graphql")
			scope.SetExtras(map[string]any{
				"Field":      fieldCtx.Field.Name,
				"Type":       fieldCtx.Object,
				"Request ID": apexlogutils_middleware.GetReqID(ctx),
			})

			var eventID string
			eID := hub.CaptureException(err)
			if eID != nil {
				eventID = string(*eID)
				log.
					WithField("sentryEventId", eventID).
					Infof("Captured error with Sentry")
			}
		})
	}

	return res, err
}
