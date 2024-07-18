package graphql

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/apex/log"
	graphql_ws "github.com/korylprince/go-graphql-ws"
	"github.com/stretchr/testify/require"

	"myvendor.mytld/myproject/backend/api"
	api_handler "myvendor.mytld/myproject/backend/api/handler"
	http_api "myvendor.mytld/myproject/backend/api/http"
	test_auth "myvendor.mytld/myproject/backend/test/auth"
)

func ServerAndSubscribe[T any](t *testing.T, deps api.ResolverDependencies, subscription *graphql_ws.MessagePayloadStart) chan T {
	t.Helper()
	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/query", nil)
	require.NoError(t, err)
	test_auth.ApplyFixedAuthValuesOrganisationAdministrator(t, deps.TimeSource, req)

	graphqlHandler := api_handler.NewGraphqlHandler(deps, api_handler.Config{
		DisableRecover: true,
	})
	srv := http_api.MiddlewareStackWithAuth(deps, graphqlHandler)

	s := httptest.NewServer(srv)
	t.Cleanup(s.Close)

	wsURL := httpToWs(t, s.URL) + "/query"

	log.Debugf("connecting to: %s", wsURL)

	conn, resp, err := graphql_ws.DefaultDialer.Dial(wsURL, req.Header, nil)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := resp.Body.Close()
		require.NoError(t, err)
		_ = conn.Close()
		log.Debug("WS client: closed connection")
	})

	notifications := make(chan T, 1)

	var payload struct {
		Data   T
		Errors graphql_ws.Errors
	}

	id, err := conn.Subscribe(subscription, func(msg *graphql_ws.Message) {
		var err error
		if msg.Type == graphql_ws.MessageTypeError {
			err = graphql_ws.ParseError(msg.Payload)
			require.NoError(t, err)
			return
		}
		if msg.Type == graphql_ws.MessageTypeComplete {
			return
		}

		err = json.Unmarshal(msg.Payload, &payload)
		require.NoError(t, err)
		require.Empty(t, payload.Errors)
		notifications <- payload.Data
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = conn.Unsubscribe(id) //nolint:errcheck
		close(notifications)
	})

	return notifications
}

func httpToWs(t *testing.T, url string) string {
	t.Helper()

	if !strings.HasPrefix(url, "http") {
		t.Fatalf("expected http(s) URL, got %q", url)
	}

	return "ws" + strings.TrimPrefix(url, "http")
}
