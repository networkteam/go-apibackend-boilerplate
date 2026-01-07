package middleware

import (
	"context"
	"net/http"
	"runtime"

	"github.com/networkteam/slogutils"
)

// RecoverAndLog is a middleware that recovers from panics and logs them using slog.
// Sentry handling should go below this and re-panic for the panic to be logged.
func RecoverAndLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer func(ctx context.Context) {
			if recovered := recover(); recovered != nil {
				// Get a stack trace for the panic

				stack := make([]byte, 1<<16)
				stack = stack[:runtime.Stack(stack, false)]

				slogutils.FromContext(ctx).
					ErrorContext(ctx, "Panic while handling request", slogutils.ErrorKey, recovered, "stack", string(stack))
				w.WriteHeader(http.StatusInternalServerError)
			}
		}(ctx)
		next.ServeHTTP(w, r)
	})
}
