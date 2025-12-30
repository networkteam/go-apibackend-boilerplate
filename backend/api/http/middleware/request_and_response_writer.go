package middleware

import (
	"net/http"

	api_types "myvendor.mytld/myproject/backend/api/types"
)

// RequestAndResponseWriterMiddleware adds the HTTP ResponseWriter to
// context for access in a GraphQL root
func RequestAndResponseWriterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := api_types.WithHTTPRequest(r.Context(), r)
		ctx = api_types.WithHTTPResponse(ctx, w)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
