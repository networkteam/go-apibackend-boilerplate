package middleware

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"

	"myvendor.mytld/myproject/backend/logger"
	"myvendor.mytld/myproject/backend/security/authentication"
)

func LoggerFieldMiddleware(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
	resolverCtx := graphql.GetResolverContext(ctx)

	shouldLogResolver := (resolverCtx.Parent == nil || resolverCtx.Parent.Parent == nil) && resolverCtx.Field.Name != "__schema"
	if shouldLogResolver {
		log := logger.GetLogger(ctx)

		authCtx := authentication.GetAuthContext(ctx)
		log.
			WithFields(authCtx).
			Debug("Authentication context")

		defer log.WithField("field", resolverCtx.Field.Name).
			WithField("type", resolverCtx.Object).
			Trace(fmt.Sprintf("GraphQL %s %s", resolverCtx.Object, resolverCtx.Field.Name)).Stop(&err)
	}

	res, err = next(ctx)

	return
}
