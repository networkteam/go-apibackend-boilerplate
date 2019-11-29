package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/99designs/gqlgen-contrib/gqlapollotracing"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/handler"
	sentry "github.com/getsentry/sentry-go"
	"github.com/gorilla/handlers"
	"github.com/friendsofgo/errors"

	"myvendor.mytld/myproject/backend/api"
	"myvendor.mytld/myproject/backend/api/middleware"
	"myvendor.mytld/myproject/backend/api/root"
	"myvendor.mytld/myproject/backend/security/authentication"
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
	config := api.Config{
		Resolvers: &root.Resolver{
			ResolverDependencies: deps,
		},
		Directives: api.DirectiveRoot{
			// No op implementation
			BypassAuthentication: func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
				return next(ctx)
			},
		},
	}
	exec := api.NewExecutableSchema(config)
	var opts []handler.Option
	if handlerConfig.EnableLogging {
		opts = append(opts, handler.ResolverMiddleware(middleware.LoggerFieldMiddleware))
	}

	opts = append(opts, handler.ResolverMiddleware(middleware.RequireAuthenticationFieldMiddleware))
	opts = append(opts, handler.ResolverMiddleware(middleware.SentryGraphqlMiddleware))

	if handlerConfig.EnableTracing {
		opts = append(
			opts,
			handler.RequestMiddleware(gqlapollotracing.RequestMiddleware()),
			handler.Tracer(gqlapollotracing.NewTracer()),
		)
	}
	if !handlerConfig.DisableRecover {
		opts = append(opts, handler.RecoverFunc(sentryRecoverFunc))
	}
	graphqlHandler := handler.GraphQL(
		exec,
		opts...,
	)

	return middleware.RequestID(
		handlers.ProxyHeaders(
			middleware.SentryMiddleware(
				middleware.AuthTokenMiddleware(
					middleware.CsrfTokenMiddleware(
						middleware.AuthContextMiddleware(
							deps.Db,
							deps.TimeSource,
							middleware.RefreshTokensMiddleware(
								deps.Db,
								deps.TimeSource,
								middleware.RequestAndResponseWriterMiddleware(
									graphqlHandler,
								),
							),
						),
					),
				),
			),
		),
	)
}

func sentryRecoverFunc(ctx context.Context, err interface{}) error {
	req := api.GetHTTPRequest(ctx)

	parts := []string{""}
	if req.RemoteAddr != "" {
		parts = strings.Split(req.RemoteAddr, ":")
	}

	remoteAddr := parts[0]

	userInfo := authentication.GetAuthContext(ctx)
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{
			ID: userInfo.AccountID.String(),
		})

		scope.SetContext("Request", map[string]string{
			"Method":    req.Method,
			"URL":       req.RequestURI,
			"Client IP": remoteAddr,
		})
	})

	var newErr error
	if realErr, ok := err.(error); ok {
		newErr = realErr
	} else {
		newErr = errors.New(fmt.Sprintf("%s", err))
	}

	sentry.CaptureException(newErr)

	return newErr
}
