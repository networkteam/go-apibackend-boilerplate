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
