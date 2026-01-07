package middleware

import (
	"net/http"

	api_types "myvendor.mytld/myproject/backend/api/types"
	"myvendor.mytld/myproject/backend/security/authentication"
)

// CsrfTokenMiddleware adds the CSRF token from a HTTP request
// to context for access in a GraphQL root
func CsrfTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		csrfToken := authentication.GetCsrfTokenFromHeader(r)

		ctx := api_types.WithCsrfToken(r.Context(), csrfToken)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
