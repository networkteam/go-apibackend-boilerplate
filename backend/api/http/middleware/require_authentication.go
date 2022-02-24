package middleware

import (
	"fmt"
	"net/http"

	"myvendor.mytld/myproject/backend/security/authentication"
)

func RequireAuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		authCtx := authentication.GetAuthContext(ctx)
		if authCtx.Authenticated || authCtx.IgnoreAuthenticationState {
			// Proceed if auth context is authenticated or authentication state should be ignored
			next.ServeHTTP(w, r)
		} else if authCtx.Error != nil {
			// If specific error is set in auth context, send it
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintf(w, "Authentication failure: %v", authCtx.Error)
		} else {
			// Otherwise authentication is required
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprint(w, "Authentication required")
		}
	})
}
