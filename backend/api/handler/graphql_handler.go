package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	graphql_handler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/apollotracing"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gorilla/websocket"

	"myvendor.mytld/myproject/backend/api"
	"myvendor.mytld/myproject/backend/api/graph"
	"myvendor.mytld/myproject/backend/api/graph/generated"
	graphql_middleware "myvendor.mytld/myproject/backend/api/graph/middleware"
	http_middleware "myvendor.mytld/myproject/backend/api/http/middleware"
)

type Config struct {
	EnableTracing        bool
	EnableLogging        bool
	DisableRecover       bool
	WebsocketAllowOrigin string
	// Constant time duration for sensitive operations (e.g. login / request password reset / perform password reset / registration)
	SensitiveOperationConstantTime time.Duration
}

func NewGraphqlHandler(deps api.ResolverDependencies, handlerConfig Config) http.Handler {
	config := generated.Config{
		Resolvers: &graph.Resolver{
			ResolverDependencies: deps,
			ResolverConfig: api.ResolverConfig{
				SensitiveOperationConstantTime: handlerConfig.SensitiveOperationConstantTime,
			},
		},
		Directives: generated.DirectiveRoot{
			// No op implementation, will be checked in middleware
			BypassAuthentication: func(ctx context.Context, obj any, next graphql.Resolver) (res any, err error) {
				return next(ctx)
			},
		},
	}
	exec := generated.NewExecutableSchema(config)
	srv := newDefaultServer(exec, handlerConfig)
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

// Copied from graphql_handler.NewDefaultServer to include the CheckOrigin function for the Websocket transport
func newDefaultServer(es graphql.ExecutableSchema, handlerConfig Config) *graphql_handler.Server {
	srv := graphql_handler.New(es)

	websocketTransport := transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
	}

	// If an allowed origin for websockets is configured, use it - otherwise the default check (origin == request host) is used.
	if handlerConfig.WebsocketAllowOrigin != "" {
		websocketTransport.Upgrader = websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return r.Header.Get("Origin") == handlerConfig.WebsocketAllowOrigin
			},
		}
	}

	srv.AddTransport(websocketTransport)
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	srv.SetQueryCache(lru.New(1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New(100),
	})

	return srv
}
