package sentry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/networkteam/slogutils"
)

func RequestCapture(handlerName string, requestContextName string) func(ctx context.Context, req *http.Request) error {
	return func(ctx context.Context, req *http.Request) error {
		logger := slogutils.FromContext(ctx).
			With("handler", handlerName)

		if req == nil {
			return nil
		}

		var (
			bodyJSON map[string]*json.RawMessage
			body     []byte
			err      error
		)
		if req.Body != nil {
			body, err = io.ReadAll(req.Body)
			if err != nil {
				logger.WarnContext(ctx, "Failed to read request body")
				return nil //nolint:nilerr
			}

			req.Body = io.NopCloser(bytes.NewReader(body))

			if len(body) > 0 {
				if err = json.Unmarshal(body, &bodyJSON); err != nil {
					logger.WarnContext(ctx, "Failed unmarschal request body")
					return nil //nolint:nilerr
				}
			}
		}

		hub := sentry.GetHubFromContext(ctx)
		if hub == nil {
			logger.WarnContext(ctx, "No sentry hub on current context")
			return nil
		}

		hub.ConfigureScope(func(scope *sentry.Scope) {
			url := req.URL
			var host, path string
			if url != nil {
				host = url.Hostname()
				path = url.EscapedPath()
			}
			scope.SetTags(map[string]string{
				"handler": handlerName,
				"host":    host,
				"path":    path,
			})

			// we're using a context here, because using `scope.SetRequest` would overwrite the original GQL query
			scope.SetContext(requestContextName, map[string]any{
				"Method": req.Method,
				"URL":    url.String(),
				"Body":   bodyJSON,
			})

			scope.AddAttachment(&sentry.Attachment{
				Filename:    fmt.Sprintf("req_%s_%s%s.json", req.Method, host, strings.ReplaceAll(path, "/", "-")),
				ContentType: "application/json",
				Payload:     body,
			})
		})

		return nil
	}
}
