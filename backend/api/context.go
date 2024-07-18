package api

import (
	"context"
	"net/http"
)

type ctxKey string

const (
	httpRequestKey   ctxKey = "httpRequest"
	httpResponseKey  ctxKey = "httpResponse"
	authTokenKey     ctxKey = "authToken"
	csrfTokenKey     ctxKey = "csrfToken"
	skipCsrfCheckKey ctxKey = "skipCsrfCheck"
)

func WithHTTPResponse(ctx context.Context, w http.ResponseWriter) context.Context {
	return context.WithValue(ctx, httpResponseKey, w)
}

// GetHTTPResponse gets the http.ResponseWriter from context
func GetHTTPResponse(ctx context.Context) http.ResponseWriter {
	return ctx.Value(httpResponseKey).(http.ResponseWriter) //nolint:forcetypeassert
}

func WithHTTPRequest(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, httpRequestKey, r)
}

// GetHTTPRequest gets the *http.Request from context
func GetHTTPRequest(ctx context.Context) *http.Request {
	return ctx.Value(httpRequestKey).(*http.Request) //nolint:forcetypeassert
}

func WithAuthToken(ctx context.Context, authToken string) context.Context {
	return context.WithValue(ctx, authTokenKey, authToken)
}

// GetAuthToken gets the auth token (e.g. from an underlying http request) from context
func GetAuthToken(ctx context.Context) string {
	return ctx.Value(authTokenKey).(string) //nolint:forcetypeassert
}

func WithCsrfToken(ctx context.Context, csrfToken string) context.Context {
	return context.WithValue(ctx, csrfTokenKey, csrfToken)
}

// GetCsrfToken gets the CSRF token (e.g. from an underlying http request) from context
func GetCsrfToken(ctx context.Context) string {
	return ctx.Value(csrfTokenKey).(string) //nolint:forcetypeassert
}

func WithSkipCsrfCheck(ctx context.Context, skipCsrfCheck bool) context.Context {
	return context.WithValue(ctx, skipCsrfCheckKey, skipCsrfCheck)
}

// GetSkipCsrfCheck tells if the csrf check should be skipped
func GetSkipCsrfCheck(ctx context.Context) bool {
	return ctx.Value(skipCsrfCheckKey).(bool) //nolint:forcetypeassert
}
