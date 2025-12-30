package main

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/networkteam/devlog/collector"
)

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
