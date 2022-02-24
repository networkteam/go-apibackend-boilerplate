package http

import (
	"net/http"

	"github.com/gorilla/handlers"

	"myvendor.mytld/myproject/backend/api"
	http_middleware "myvendor.mytld/myproject/backend/api/http/middleware"
)

// MiddlewareStack combines all necessary middlewares for request processing, authentication and logging
func MiddlewareStack(deps api.ResolverDependencies, h http.Handler) http.Handler {
	return handlers.ProxyHeaders(
		http_middleware.SentryMiddleware(
			http_middleware.AuthTokenMiddleware(
				http_middleware.CsrfTokenMiddleware(
					http_middleware.AuthContextMiddleware(
						deps.DB,
						deps.TimeSource,
						http_middleware.RefreshTokensMiddleware(
							deps.DB,
							deps.TimeSource,
							h,
						),
					),
				),
			),
		),
	)
}
