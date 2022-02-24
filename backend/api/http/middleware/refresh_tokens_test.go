package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	http_middleware "myvendor.mytld/myproject/backend/api/http/middleware"
	"myvendor.mytld/myproject/backend/test"
	"myvendor.mytld/myproject/backend/test/auth"
	test_db "myvendor.mytld/myproject/backend/test/db"
)

func TestRefreshTokensMiddleware(t *testing.T) {
	db := test_db.CreateTestDatabase(t)

	test_db.ExecFixtures(t, db, "base")

	timeSource := test.FixedTime()

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	// Create the stack of middlewares to test behaviour like in handler.NewGraphqlHandler
	srv := http_middleware.AuthTokenMiddleware(
		http_middleware.CsrfTokenMiddleware(
			http_middleware.AuthContextMiddleware(
				db,
				timeSource,
				http_middleware.RefreshTokensMiddleware(
					db,
					timeSource,
					http_middleware.RequireAuthenticationMiddleware(h),
				),
			),
		),
	)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "http://localhost/query", nil)
	auth.ApplyFixedAuthValuesSystemAdministrator(t, timeSource, req)

	// Test initial request is allowed
	srv.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	setCookieHeader := w.Header().Get("Set-Cookie")
	require.Empty(t, setCookieHeader)
	refreshCsrfTokenHeader := w.Header().Get("X-Refresh-CSRF-Token")
	require.Empty(t, refreshCsrfTokenHeader)

	// Try again with older token

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "http://localhost/query", nil)
	// Use a time source that is before the token refresh threshold
	previousTimeSource := timeSource.Add(-2 * http_middleware.AuthTokenRefreshThreshold)
	auth.ApplyFixedAuthValuesSystemAdministrator(t, previousTimeSource, req)

	// Test that we now get refreshed tokens
	srv.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	setCookieHeader = w.Header().Get("Set-Cookie")
	require.NotEmpty(t, setCookieHeader, "Set-Cookie header")
	refreshCsrfTokenHeader = w.Header().Get("X-Refresh-CSRF-Token")
	require.NotEmpty(t, refreshCsrfTokenHeader, "X-Refresh-CSRF-Token")

	// Let's use the refreshed tokens

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "http://localhost/query", nil)
	req.Header.Set("Cookie", setCookieHeader)
	req.Header.Set("X-CSRF-Token", refreshCsrfTokenHeader)

	// Test the request is still authenticated
	srv.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}
