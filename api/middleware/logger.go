package middleware

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/apex/log"
)

func LoggerFieldMiddleware(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
	resolverCtx := graphql.GetResolverContext(ctx)

	shouldLogResolver := (resolverCtx.Parent == nil || resolverCtx.Parent.Parent == nil) && resolverCtx.Field.Name != "__schema"
	if shouldLogResolver {
		reqID := GetReqID(ctx)
		defer log.WithField("field", resolverCtx.Field.Name).
			WithField("type", resolverCtx.Object).
			WithField("requestID", reqID).
			Trace(fmt.Sprintf("GraphQL %s %s", resolverCtx.Object, resolverCtx.Field.Name)).Stop(&err)
	}

	res, err = next(ctx)

	return
}
