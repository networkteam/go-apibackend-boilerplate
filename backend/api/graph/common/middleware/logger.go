package middleware

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/networkteam/slogutils"

	"myvendor.mytld/myproject/backend/api"
	"myvendor.mytld/myproject/backend/security/authorization"
)

func LoggerFieldMiddleware(ctx context.Context, next graphql.Resolver) (res any, err error) {
	fieldCtx := graphql.GetFieldContext(ctx)

	shouldLogResolver := (fieldCtx.Parent == nil || fieldCtx.Parent.Parent == nil) && fieldCtx.Field.Name != "__schema"

	start := time.Now()

	res, err = next(ctx)

	if shouldLogResolver {
		logger := slogutils.FromContext(ctx).
			With("component", "graphql")

		logEntry := logger.With(
			"field", fieldCtx.Field.Name,
			"type", fieldCtx.Object,
			"duration", time.Since(start),
		)

		logLevel := decideLogLevel(ctx, err, fieldCtx)

		// Include error details in log if present
		if err != nil {
			logEntry = logEntry.With(slogutils.Err(err))
		}

		// Log GraphQL event
		// Note: logLevel slog.LevelError also sends an event to sentry
		logEntry.Log(ctx, logLevel, fieldCtx.Object, "field", fieldCtx.Field.Name)
	}

	return res, err
}

func decideLogLevel(ctx context.Context, err error, fieldCtx *graphql.FieldContext) slog.Level {
	logLevel := slog.LevelDebug

	// Log mutations with info level
	if fieldCtx.Object == "Mutation" {
		logLevel = slog.LevelInfo
	}

	if err == nil {
		return logLevel
	}

	// expected typed errors
	var typedError api.TypedError
	if errors.As(err, &typedError) {
		return slog.LevelWarn
	}

	// expected fields errors
	if fieldsError := api.FieldsErrorFromErr(err); fieldsError != nil {
		return slog.LevelWarn
	}

	// expected auth errors
	var authError authorization.Error
	if errors.As(err, &authError) {
		return slog.LevelWarn
	}

	// expected context cancelled errors
	if errors.Is(err, context.Canceled) {
		// Check if ctx was cancelled to avoid ignoring errors that were cancelled inside the resolver
		select {
		case <-ctx.Done():
			return slog.LevelDebug
		default:
		}
	}

	return slog.LevelError
}
