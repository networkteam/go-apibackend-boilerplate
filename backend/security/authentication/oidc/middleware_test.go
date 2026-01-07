package oidc_test

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/coreos/go-oidc/v3/oidc/oidctest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"myvendor.mytld/myproject/backend/security/authentication/oidc"
)

func newTestOIDCServer(t *testing.T) *httptest.Server {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "should generate RSA key")

	oidcServer := &oidctest.Server{
		PublicKeys: []oidctest.PublicKey{
			{
				PublicKey: priv.Public(),
				KeyID:     "test-key",
				Algorithm: "RS256",
			},
		},
	}

	srv := httptest.NewServer(oidcServer)
	oidcServer.SetIssuer(srv.URL)

	t.Cleanup(func() {
		srv.Close()
	})

	return srv
}

func newTestMiddleware(t *testing.T, oidcServerURL string) *oidc.Middleware {
	t.Helper()

	cfg := &oidc.Config{
		GitLabURL:     oidcServerURL,
		ClientID:      "test-client-id",
		ClientSecret:  "test-client-secret",
		RedirectURL:   "http://localhost/auth/callback",
		SessionSecret: []byte("test-session-secret-32-bytes-ok"),
	}

	middleware, err := oidc.NewMiddleware(cfg)
	require.NoError(t, err, "should create middleware")

	return middleware
}

func TestNewMiddleware(t *testing.T) {
	t.Run("valid config creates middleware", func(t *testing.T) {
		oidcServer := newTestOIDCServer(t)

		cfg := &oidc.Config{
			GitLabURL:     oidcServer.URL,
			ClientID:      "test-client-id",
			ClientSecret:  "test-client-secret",
			RedirectURL:   "http://localhost/auth/callback",
			SessionSecret: []byte("test-session-secret-32-bytes-ok"),
		}

		middleware, err := oidc.NewMiddleware(cfg)

		require.NoError(t, err, "should not return error")
		require.NotNil(t, middleware, "should return middleware")
	})

	t.Run("invalid config returns error", func(t *testing.T) {
		cfg := &oidc.Config{
			// Missing required fields
		}

		middleware, err := oidc.NewMiddleware(cfg)

		require.Error(t, err, "should return error")
		require.Nil(t, middleware, "should not return middleware")
	})
}

func TestMiddleware_Wrap(t *testing.T) {
	oidcServer := newTestOIDCServer(t)
	middleware := newTestMiddleware(t, oidcServer.URL)

	nextHandlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextHandlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Wrap(nextHandler)

	t.Run("unauthenticated request to protected route redirects to login", func(t *testing.T) {
		nextHandlerCalled = false
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusFound, rec.Code, "should redirect")
		location := rec.Header().Get("Location")
		assert.Contains(t, location, "/auth/login", "should redirect to login")
		assert.False(t, nextHandlerCalled, "should not call next handler")
	})

	t.Run("login flow", func(t *testing.T) {
		var loginCookies []*http.Cookie
		var extractedState string

		t.Run("initiates OAuth flow", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			require.Equal(t, http.StatusFound, rec.Code, "should redirect to provider")

			location := rec.Header().Get("Location")
			require.NotEmpty(t, location, "should have Location header")
			assert.Contains(t, location, oidcServer.URL, "should redirect to OIDC provider")

			// Extract state from redirect URL
			parsedURL, err := url.Parse(location)
			require.NoError(t, err, "should parse redirect URL")
			extractedState = parsedURL.Query().Get("state")
			assert.NotEmpty(t, extractedState, "should include state parameter")

			// Save cookies for subsequent requests
			loginCookies = rec.Result().Cookies()
			require.NotEmpty(t, loginCookies, "should set session cookie")
		})

		t.Run("callback with state mismatch fails", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/auth/callback?state=wrong-state", nil)
			for _, cookie := range loginCookies {
				req.AddCookie(cookie)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			require.Equal(t, http.StatusBadRequest, rec.Code, "should return bad request")
		})

		t.Run("callback with OAuth error fails", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/auth/callback?state="+extractedState+"&error=access_denied&error_description=User+denied+access", nil)
			for _, cookie := range loginCookies {
				req.AddCookie(cookie)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			require.Equal(t, http.StatusUnauthorized, rec.Code, "should return unauthorized")
		})

		t.Run("callback without code fails", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/auth/callback?state="+extractedState, nil)
			for _, cookie := range loginCookies {
				req.AddCookie(cookie)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			require.Equal(t, http.StatusBadRequest, rec.Code, "should return bad request")
		})
	})

	t.Run("logout redirects to login", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/logout", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusFound, rec.Code, "should redirect")
		location := rec.Header().Get("Location")
		assert.True(t, strings.HasSuffix(location, "/auth/login"), "should redirect to login")
	})
}
