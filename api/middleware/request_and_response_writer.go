package middleware

import (
	"net/http"

	"myvendor/myproject/backend/api"
)

// RequestAndResponseWriterMiddleware adds the HTTP ResponseWriter to
// context for access in a GraphQL root
func RequestAndResponseWriterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := api.WithHTTPRequest(r.Context(), r)
		ctx = api.WithHTTPResponse(ctx, w)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
