package middleware

import (
	"net/http"

	"github.com/nrednav/cuid2"
	sloghttp "github.com/samber/slog-http"
)

// RequestID is a middleware that uses an external request ID or generates a new one (short CUID2) and sets it as a header.
// The header can be used by a logger middleware to log the request ID with each log entry.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(sloghttp.RequestIDHeaderKey)
		if requestID == "" {
			requestID = cuid2.Generate()
			r.Header.Set(sloghttp.RequestIDHeaderKey, requestID)
		}
		next.ServeHTTP(w, r)
	})
}
