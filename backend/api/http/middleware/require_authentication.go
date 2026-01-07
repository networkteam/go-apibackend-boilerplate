package middleware

import (
	"fmt"
	"html"
	"net/http"
	"slices"

	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/security/authentication"
)

func RequireAuthenticationMiddleware(next http.Handler, requireRoles ...types.Role) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authCtx := authentication.GetAuthContext(ctx)

		if authCtx.Error != nil {
			// If specific error is set in auth context, send it
			w.WriteHeader(http.StatusUnauthorized)
			// nosemgrep: go.lang.security.audit.xss.no-fprintf-to-responsewriter.no-fprintf-to-responsewriter
			_, _ = fmt.Fprintf(w, "Authentication failure: %s", html.EscapeString(authCtx.Error.Error()))
			return
		}

		if !authCtx.Authenticated && !authCtx.IgnoreAuthenticationState {
			// Otherwise authentication is required
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprint(w, "Authentication required")

			return
		}

		if len(requireRoles) > 0 && !slices.Contains(requireRoles, authCtx.Role) {
			// If no required role matches the authenticated user's role, return unauthorized
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintf(w, "Authentication failure: insufficient role")
			return
		}

		// Proceed if auth context is authenticated or authentication state should be ignored
		next.ServeHTTP(w, r)
	})
}
