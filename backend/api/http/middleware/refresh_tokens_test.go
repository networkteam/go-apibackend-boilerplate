package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"myvendor.mytld/myproject/backend/api"
	http_middleware "myvendor.mytld/myproject/backend/api/http/middleware"
	"myvendor.mytld/myproject/backend/test"
	"myvendor.mytld/myproject/backend/test/auth"
	test_db "myvendor.mytld/myproject/backend/test/db"
)

func TestRefreshTokensMiddleware(t *testing.T) {
	db := test_db.CreateTestDatabase(t)

	test_db.ExecFixtures(t, db, "base")

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	fixedDeps := api.ResolverDependencies{
		DB:         db,
		TimeSource: test.FixedTime(),
		Config:     auth.FixedConfig,
	}

	// Create the stack of middlewares to test behaviour like in handler.NewGraphqlHandler
	srv := http_middleware.AccessTokenMiddleware(
		http_middleware.CsrfTokenMiddleware(
			http_middleware.AuthContextMiddleware(
				fixedDeps,
				http_middleware.RefreshTokensMiddleware(
					fixedDeps,
					http_middleware.RequireAuthenticationMiddleware(h),
				),
			),
		),
	)

	authValuesFunc := auth.ApplyAuthValuesFuncSystemAdministrator("admin@example.com")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "http://localhost/query", nil)
	authValuesFunc(t, fixedDeps, req)

	// Test initial request is allowed
	srv.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	// No token refresh should happen
	require.Empty(t, rec.Result().Cookies())

	// Try again with older token
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "http://localhost/query", nil)

	fixedDeps.TimeSource = test.FixedTime().Add(-2 * http_middleware.AccessTokenRefreshThreshold) // Use a time source that is before the token refresh threshold
	authValuesFunc(t, fixedDeps, req)

	// Test that we now get refreshed tokens
	srv.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	cookies := rec.Result().Cookies()
	accessTokenCookie := test.RequireHasCookie(t, cookies, "access_token")
	csrfTokenCookie := test.RequireHasCookie(t, cookies, "csrf_token")

	// Let's use the refreshed tokens
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "http://localhost/query", nil)
	req.AddCookie(accessTokenCookie)
	req.Header.Set("X-CSRF-Token", csrfTokenCookie.Value)

	// Test the request is still authenticated
	srv.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

// TestRefreshTokensMiddleware_OldCsrfTokenStillValidAfterRefresh tests that when the access token
// is refreshed, the old CSRF token (tied to AuthSessionID, not AccessTokenID) remains valid.
// This prevents race conditions where the browser sends a new access token (HTTP-only cookie
// auto-updated) but JavaScript hasn't yet read the new CSRF token from document.cookie.
func TestRefreshTokensMiddleware_OldCsrfTokenStillValidAfterRefresh(t *testing.T) {
	db := test_db.CreateTestDatabase(t)

	test_db.ExecFixtures(t, db, "base")

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	fixedDeps := api.ResolverDependencies{
		DB:         db,
		TimeSource: test.FixedTime(),
		Config:     auth.FixedConfig,
	}

	// Create the stack of middlewares to test behaviour like in handler.NewGraphqlHandler
	srv := http_middleware.AccessTokenMiddleware(
		http_middleware.CsrfTokenMiddleware(
			http_middleware.AuthContextMiddleware(
				fixedDeps,
				http_middleware.RefreshTokensMiddleware(
					fixedDeps,
					http_middleware.RequireAuthenticationMiddleware(h),
				),
			),
		),
	)

	authValuesFunc := auth.ApplyAuthValuesFuncSystemAdministrator("admin@example.com")

	// Step 1: Create initial tokens with old time source (to trigger refresh)
	fixedDeps.TimeSource = test.FixedTime().Add(-2 * http_middleware.AccessTokenRefreshThreshold)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "http://localhost/query", nil)
	authValuesFunc(t, fixedDeps, req)

	// Capture the old CSRF token from the request header
	oldCsrfToken := req.Header.Get("X-CSRF-Token")
	require.NotEmpty(t, oldCsrfToken, "CSRF token should be set")

	// Step 2: Make request - this will trigger token refresh
	srv.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, "initial request should succeed")

	// Step 3: Get the new access token from the refresh response
	cookies := rec.Result().Cookies()
	newAccessTokenCookie := test.RequireHasCookie(t, cookies, "access_token")
	newCsrfTokenCookie := test.RequireHasCookie(t, cookies, "csrf_token")

	// Verify we got new tokens (different values)
	require.NotEqual(t, oldCsrfToken, newCsrfTokenCookie.Value, "CSRF token should have been refreshed")

	// Step 4: Simulate race condition: use NEW access token but OLD CSRF token
	// This mimics the browser sending the updated HTTP-only access_token cookie
	// while JavaScript still has the stale csrf_token from document.cookie
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "http://localhost/query", nil)
	req.AddCookie(newAccessTokenCookie)          // NEW access token (browser auto-sends)
	req.Header.Set("X-CSRF-Token", oldCsrfToken) // OLD CSRF token (JS read stale value)

	// Step 5: Request should succeed because CSRF token is tied to AuthSessionID (unchanged)
	srv.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, "request with old CSRF token should succeed after access token refresh")
}
