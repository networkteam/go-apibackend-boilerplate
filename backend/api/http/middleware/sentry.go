package middleware

import (
	"net/http"

	sentryhttp "github.com/getsentry/sentry-go/http"
)

func SentryMiddleware(next http.Handler) http.Handler {
	sentryHandler := sentryhttp.New(sentryhttp.Options{})

	return sentryHandler.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	}))
}
