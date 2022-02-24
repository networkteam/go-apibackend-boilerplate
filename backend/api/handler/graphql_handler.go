package handler

import (
	"context"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	graphql_handler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/apollotracing"

	"myvendor.mytld/myproject/backend/api"
	"myvendor.mytld/myproject/backend/api/graph"
	"myvendor.mytld/myproject/backend/api/graph/generated"
	graphql_middleware "myvendor.mytld/myproject/backend/api/graph/middleware"
	http_middleware "myvendor.mytld/myproject/backend/api/http/middleware"
)

type HandlerConfig struct {
	EnableTracing  bool
	EnableLogging  bool
	DisableRecover bool
}

func NewGraphqlHandler(
	deps api.ResolverDependencies,
	handlerConfig HandlerConfig,
) http.Handler {
	config := generated.Config{
		Resolvers: &graph.Resolver{
			ResolverDependencies: deps,
		},
		Directives: generated.DirectiveRoot{
			// No op implementation, will be checked in middleware
			BypassAuthentication: func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
				return next(ctx)
			},
		},
	}
	exec := generated.NewExecutableSchema(config)
	srv := graphql_handler.NewDefaultServer(exec)
	srv.SetErrorPresenter(ErrorPresenter)

	if handlerConfig.EnableLogging {
		srv.AroundFields(graphql_middleware.LoggerFieldMiddleware)
	}

	srv.AroundFields(graphql_middleware.RequireAuthenticationFieldMiddleware)
	srv.AroundFields(graphql_middleware.SentryGraphqlMiddleware)

	if handlerConfig.EnableTracing {
		srv.Use(apollotracing.Tracer{})
	}
	if !handlerConfig.DisableRecover {
		srv.SetRecoverFunc(sentryRecoverFunc)
	}
	// else: DefaultRecover from gqlgen is okay for tests, it dumps a stacktrace to the console

	return http_middleware.RequestAndResponseWriterMiddleware(
		srv,
	)
}
