//go:generate go run github.com/99designs/gqlgen generate
package admin

import (
	"context"

	"github.com/99designs/gqlgen/graphql"

	"myvendor.mytld/myproject/backend/api"
	"myvendor.mytld/myproject/backend/api/graph/admin/generated"
	"myvendor.mytld/myproject/backend/api/handler"
)

func BuildExecutableSchema(deps api.ResolverDependencies, handlerConfig handler.Config) graphql.ExecutableSchema {
	config := generated.Config{
		Resolvers: NewResolver(deps, api.ResolverConfig{
			SensitiveOperationConstantTime: handlerConfig.SensitiveOperationConstantTime,
		}),
		Directives: generated.DirectiveRoot{
			// No op implementation, will be checked in middleware
			BypassAuthentication: func(ctx context.Context, _ any, next graphql.Resolver) (res any, err error) {
				return next(ctx)
			},
		},
	}
	return generated.NewExecutableSchema(config)
}
