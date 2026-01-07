package types

import (
	"context"
	"net/http"
)

type ctxKey string

const (
	httpRequestKey   ctxKey = "httpRequest"
	httpResponseKey  ctxKey = "httpResponse"
	accessTokenKey   ctxKey = "accessToken"
	csrfTokenKey     ctxKey = "csrfToken"
	skipCsrfCheckKey ctxKey = "skipCsrfCheck"
)

func WithHTTPResponse(ctx context.Context, w http.ResponseWriter) context.Context {
	return context.WithValue(ctx, httpResponseKey, w)
}

// GetHTTPResponse gets the http.ResponseWriter from context
func GetHTTPResponse(ctx context.Context) http.ResponseWriter {
	return ctx.Value(httpResponseKey).(http.ResponseWriter) //nolint:forcetypeassert,errcheck
}

func WithHTTPRequest(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, httpRequestKey, r)
}

// GetHTTPRequest gets the *http.Request from context
func GetHTTPRequest(ctx context.Context) *http.Request {
	return ctx.Value(httpRequestKey).(*http.Request) //nolint:forcetypeassert,errcheck
}

func WithAccessToken(ctx context.Context, accessToken string) context.Context {
	return context.WithValue(ctx, accessTokenKey, accessToken)
}

// GetAccessToken gets the auth token (e.g. from an underlying http request) from context
func GetAccessToken(ctx context.Context) string {
	return ctx.Value(accessTokenKey).(string) //nolint:forcetypeassert,errcheck
}

func WithCsrfToken(ctx context.Context, csrfToken string) context.Context {
	return context.WithValue(ctx, csrfTokenKey, csrfToken)
}

// GetCsrfToken gets the CSRF token (e.g. from an underlying http request) from context
func GetCsrfToken(ctx context.Context) string {
	return ctx.Value(csrfTokenKey).(string) //nolint:forcetypeassert,errcheck
}

func WithSkipCsrfCheck(ctx context.Context, skipCsrfCheck bool) context.Context {
	return context.WithValue(ctx, skipCsrfCheckKey, skipCsrfCheck)
}

// GetSkipCsrfCheck tells if the CSRF check should be skipped
func GetSkipCsrfCheck(ctx context.Context) bool {
	return ctx.Value(skipCsrfCheckKey).(bool) //nolint:forcetypeassert,errcheck
}
