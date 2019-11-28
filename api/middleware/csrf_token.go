package middleware

import (
	"net/http"

	"myvendor/myproject/backend/api"
)

// CsrfTokenMiddleware adds the CSRF token from a HTTP request
// to context for access in a GraphQL root
func CsrfTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		csrfToken := r.Header.Get("X-CSRF-Token")

		ctx := api.WithCsrfToken(r.Context(), csrfToken)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
