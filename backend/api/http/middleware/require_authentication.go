package middleware

import (
	"fmt"
	"html"
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
			return
		}

		if authCtx.Error != nil {
			// If specific error is set in auth context, send it
			w.WriteHeader(http.StatusUnauthorized)
			// nosemgrep: go.lang.security.audit.xss.no-fprintf-to-responsewriter.no-fprintf-to-responsewriter
			_, _ = fmt.Fprintf(w, "Authentication failure: %s", html.EscapeString(authCtx.Error.Error()))
			return
		}

		// Otherwise authentication is required
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprint(w, "Authentication required")
	})
}
