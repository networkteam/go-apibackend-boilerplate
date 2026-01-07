package main

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/jackc/pgx/v5"
	"github.com/networkteam/devlog"
	"github.com/networkteam/devlog/collector"
	"github.com/networkteam/devlog/dashboard"
	"github.com/networkteam/slogutils"
	"github.com/urfave/cli/v2"

	"myvendor.mytld/myproject/backend/security/authentication/basic"
	"myvendor.mytld/myproject/backend/security/authentication/oidc"
	"myvendor.mytld/myproject/backend/security/helper"
)

func devlogFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    "devlog",
			Usage:   "Enable devlog for development tracing (HTTP requests and logs), do not enable in production!",
			Value:   false,
			EnvVars: []string{"DEVLOG_ENABLED"},
		},
		&cli.StringFlag{
			Name:    "devlog-auth",
			Usage:   "Enable devlog authentication. Currently 'basic' and 'oidc' are supported.",
			EnvVars: []string{"DEVLOG_AUTH"},
		},
		&cli.StringFlag{
			Name:    "devlog-oidc-url",
			Usage:   "Enable OIDC authentication for devlog (e.g. https://gitlab.example.com)",
			EnvVars: []string{"DEVLOG_OIDC_URL"},
		},
		&cli.StringFlag{
			Name:    "devlog-oidc-client-id",
			Usage:   "Client ID for OIDC authentication for devlog",
			EnvVars: []string{"DEVLOG_OIDC_CLIENT_ID"},
		},
		&cli.StringFlag{
			Name:    "devlog-oidc-client-secret",
			Usage:   "Client secret for OIDC authentication for devlog",
			EnvVars: []string{"DEVLOG_OIDC_CLIENT_SECRET"},
		},
		&cli.StringFlag{
			Name:    "devlog-oidc-redirect-url",
			Usage:   "Redirect URL for OIDC authentication for devlog",
			EnvVars: []string{"DEVLOG_OIDC_REDIRECT_URL"},
		},
		&cli.StringSliceFlag{
			Name:    "devlog-oidc-allowed-groups",
			Usage:   "Allowed groups for OIDC authentication for devlog",
			EnvVars: []string{"DEVLOG_OIDC_ALLOWED_GROUPS"},
		},
		&cli.BoolFlag{
			Name:    "devlog-oidc-session-secure",
			Usage:   "Set secure flag on devlog OIDC session cookies",
			EnvVars: []string{"DEVLOG_OIDC_SESSION_SECURE"},
		},
		&cli.StringFlag{
			Name:    "devlog-basic-username",
			Usage:   "Username for basic authentication for devlog (only if devlog-auth is set to 'basic')",
			EnvVars: []string{"DEVLOG_BASIC_USERNAME"},
		},
		&cli.StringFlag{
			Name:    "devlog-basic-password",
			Usage:   "Password for basic authentication for devlog (only if devlog-auth is set to 'basic')",
			EnvVars: []string{"DEVLOG_BASIC_PASSWORD"},
		},
	}
}

type devlogQueryTracer struct {
	collect func(ctx context.Context, dbQuery collector.DBQuery)
}

type dbQueryCtxKey struct{}

func (d *devlogQueryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	return context.WithValue(ctx, dbQueryCtxKey{}, collector.DBQuery{
		Query:     data.SQL,
		Language:  "postgresql",
		Args:      toNamedValue(data.Args),
		Timestamp: time.Now(),
	})
}

func (d *devlogQueryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	if dbQuery, ok := ctx.Value(dbQueryCtxKey{}).(collector.DBQuery); ok {
		dbQuery.Duration = time.Since(dbQuery.Timestamp)
		dbQuery.Error = data.Err
		d.collect(ctx, dbQuery)
	}
}

func newDevlogQueryTracer(collect func(ctx context.Context, dbQuery collector.DBQuery)) pgx.QueryTracer {
	return &devlogQueryTracer{
		collect: collect,
	}
}

func toNamedValue(args []any) []driver.NamedValue {
	// Skip the first argument if it is of type pgx.QueryResultFormatsByOID
	if len(args) > 0 {
		if _, ok := args[0].(pgx.QueryResultFormatsByOID); ok {
			args = args[1:]
		}
	}

	namedValues := make([]driver.NamedValue, len(args))
	for i, arg := range args {
		namedValues[i] = driver.NamedValue{
			Value:   arg,
			Ordinal: i + 1,
		}
	}
	return namedValues
}

func devlogHTTPServerOptions() *collector.HTTPServerOptions {
	httpServerOptions := collector.DefaultHTTPServerOptions()
	httpServerOptions.Transformers = []collector.HTTPServerRequestTransformer{
		// Add a request transformer to add a custom tag for the GraphQL operation name
		func(request collector.HTTPServerRequest) collector.HTTPServerRequest {
			// Check if we have a GraphQL request
			if request.Path != "/query" {
				return request
			}

			requestContentType := request.RequestHeaders.Get("Content-Type")
			// POST request with JSON body
			if strings.Contains(requestContentType, "application/json") {
				var graphqlBody struct {
					OperationName string `json:"operationName"`
				}
				if err := json.Unmarshal(request.RequestBody.Bytes(), &graphqlBody); err == nil {
					request.Tags["operation"] = graphqlBody.OperationName
				}
				return request
			}

			// GET request with query parameters
			if parsedURL, err := url.Parse(request.URL); err == nil {
				if operationName := parsedURL.Query().Get("operationName"); operationName != "" {
					request.Tags["operation"] = operationName
				}
			}
			return request
		},
	}
	return &httpServerOptions
}

func addDevlogDashboard(c *cli.Context, mux *http.ServeMux, dlog *devlog.Instance) error {
	logger := slogutils.FromContext(c.Context)

	dashboardHandler := dlog.DashboardHandler(
		"/_devlog",
		dashboard.WithMaxSessions(10),
		dashboard.WithStorageCapacity(100),
	)

	switch c.String("devlog-auth") {
	case "":
		// No authentication
		logger.WarnContext(c.Context, "Devlog enabled without authentication, don't do this in production!")
	case "basic":
		// Basic authentication
		username := c.String("devlog-basic-username")
		password := c.String("devlog-basic-password")
		if username == "" || password == "" {
			return errors.New("both devlog-basic-username and devlog-basic-password must be set for basic authentication")
		}
		dashboardHandler = basic.SimpleAuth(dashboardHandler, username, password)
	case "oidc":
		// OIDC authentication (GitLab)
		sessionSecret, err := helper.GenerateRandomBytes(32)
		if err != nil {
			return errors.Wrap(err, "generating random bytes for session secret")
		}

		authMiddleware, err := oidc.NewMiddleware(&oidc.Config{
			PathPrefix:          "/_devlog",
			GitLabURL:           c.String("devlog-oidc-url"),
			ClientID:            c.String("devlog-oidc-client-id"),
			ClientSecret:        c.String("devlog-oidc-client-secret"),
			RedirectURL:         c.String("devlog-oidc-redirect-url"),
			GitLabAllowedGroups: c.StringSlice("devlog-oidc-allowed-groups"),
			SessionSecret:       sessionSecret,
			SessionName:         "devlog-auth",
			SessionMaxAge:       24 * time.Hour,
			SessionSecure:       c.Bool("devlog-oidc-session-secure"),
			SessionHTTPOnly:     true,
		})
		if err != nil {
			return errors.Wrap(err, "creating OIDC middleware")
		}
		dashboardHandler = authMiddleware.Wrap(dashboardHandler)

		logger.InfoContext(c.Context, "Devlog OIDC authentication enabled", "url", c.String("devlog-oidc-url"))
	default:
		return errors.Errorf("unsupported devlog authentication method: %s", c.String("devlog-auth"))
	}

	mux.Handle("/_devlog/", http.StripPrefix("/_devlog", dashboardHandler))

	return nil
}
