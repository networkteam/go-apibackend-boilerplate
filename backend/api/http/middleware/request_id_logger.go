package middleware

import (
	"log/slog"
	"net/http"

	"github.com/networkteam/slogutils"
	sloghttp "github.com/samber/slog-http"
)

// RequestIDLogger adds the request id from sloghttp to the logger in the request context as a base attribute.
func RequestIDLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := slogutils.FromContext(r.Context())
		requestID := sloghttp.GetRequestID(r)
		if requestID != "" {
			logger = logger.With(slog.Group("http", "id", requestID))

			ctx := slogutils.WithLogger(r.Context(), logger)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}
