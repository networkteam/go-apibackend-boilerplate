package middleware

import (
	"net/http"

	"myvendor.mytld/myproject/backend/api"
	"myvendor.mytld/myproject/backend/security/authentication"
)

// AuthTokenMiddleware adds the auth token from a HTTP request
// to context for access in a GraphQL resolver / middleware
func AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authToken, skipCsrfCheck := authentication.GetAuthTokenAndSkipCsrfCheckFromRequest(r)

		ctx := api.WithAuthToken(r.Context(), authToken)
		ctx = api.WithSkipCsrfCheck(ctx, skipCsrfCheck)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
