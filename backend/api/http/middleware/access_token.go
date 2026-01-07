package middleware

import (
	"net/http"

	api_types "myvendor.mytld/myproject/backend/api/types"
	"myvendor.mytld/myproject/backend/security/authentication"
)

// AccessTokenMiddleware adds the access token from a HTTP request
// to context for access in a GraphQL resolver / middleware
func AccessTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken, skipCsrfCheck := authentication.GetAccessTokenAndSkipCsrfCheckFromRequest(r)

		ctx := api_types.WithAccessToken(r.Context(), accessToken)
		ctx = api_types.WithSkipCsrfCheck(ctx, skipCsrfCheck)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
