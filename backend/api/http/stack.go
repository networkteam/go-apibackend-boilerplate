package http

import (
	"net/http"

	"github.com/gorilla/handlers"

	"myvendor.mytld/myproject/backend/api"
	http_middleware "myvendor.mytld/myproject/backend/api/http/middleware"
)

// MiddlewareStackWithAuth combines all necessary middlewares for request processing, logging and authentication
func MiddlewareStackWithAuth(deps api.ResolverDependencies, h http.Handler) http.Handler {
	return MiddlewareStackBasic(
		http_middleware.AuthTokenMiddleware(
			http_middleware.CsrfTokenMiddleware(
				http_middleware.AuthContextMiddleware(deps.DB, deps.TimeSource,
					http_middleware.RefreshTokensMiddleware(deps.DB, deps.TimeSource, h),
				),
			),
		),
	)
}

// MiddlewareStackBasic combines all necessary middlewares for request processing and logging (without authentication)
func MiddlewareStackBasic(h http.Handler) http.Handler {
	return handlers.ProxyHeaders(
		http_middleware.SentryMiddleware(
			h,
		),
	)
}
