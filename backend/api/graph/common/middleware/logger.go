package middleware

import (
	"context"
	"time"

	"github.com/99designs/gqlgen/graphql"
	logger "github.com/apex/log"

	"myvendor.mytld/myproject/backend/api"
)

func LoggerFieldMiddleware(ctx context.Context, next graphql.Resolver) (res any, err error) {
	fieldCtx := graphql.GetFieldContext(ctx)

	shouldLogResolver := (fieldCtx.Parent == nil || fieldCtx.Parent.Parent == nil) && fieldCtx.Field.Name != "__schema"

	start := time.Now()

	res, err = next(ctx)

	if shouldLogResolver {
		log := logger.FromContext(ctx).
			WithField("component", "graphql")

		logEntry := log.WithField("field", fieldCtx.Field.Name).
			WithField("type", fieldCtx.Object).
			WithDuration(time.Since(start))

		logFunc := logEntry.Debugf
		// Log mutations with info level
		if fieldCtx.Object == "Mutation" {
			logFunc = logEntry.Infof
		}

		if err != nil {
			logEntry = logEntry.WithError(err)
			// Only warn if this is a expected domain error
			if fieldsErr := api.FieldsErrorFromErr(err); fieldsErr != nil {
				logFunc = logEntry.Warnf
			} else {
				logFunc = logEntry.Errorf
			}
		}

		logFunc("%s %s", fieldCtx.Object, fieldCtx.Field.Name)
	}

	return res, err
}
