package handler

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/playground"
)

// NewPlaygroundHandler creates a new Altair GraphQL playground handler with CSRF token support.
// It automatically injects the CSRF token from the csrf_token cookie into requests.
func NewPlaygroundHandler(title, endpoint string) http.HandlerFunc {
	options := map[string]any{
		"initialSettings": map[string]any{
			"script.allowedCookies": []string{"csrf_token"},
			"addQueryDepthLimit":    3,
		},
		"initialPreRequestScript": `const csrfToken = altair.helpers.getCookie('csrf_token');
if (csrfToken) {
  altair.helpers.setEnvironment('csrf_token', csrfToken);
}`,
		"initialHeaders": map[string]string{
			"X-CSRF-Token": "{{csrf_token}}",
		},
	}

	return playground.AltairHandler(title, endpoint, options)
}
