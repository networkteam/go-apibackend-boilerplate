package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"myvendor.mytld/myproject/backend/api/http/middleware"
	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/security/authentication"
)

func TestRequireAuthenticationMiddleware(t *testing.T) {
	// Define a simple next handler for testing
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	tests := []struct {
		name               string
		authContext        authentication.AuthContext
		requireRoles       []types.GlobalRole
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name: "Authenticated user with matching role",
			authContext: authentication.AuthContext{
				Authenticated: true,
				GlobalRole:    types.GlobalRoleSystemAdministrator,
			},
			requireRoles:       []types.GlobalRole{types.GlobalRoleSystemAdministrator},
			expectedStatusCode: http.StatusOK,
			expectedBody:       "success",
		},
		{
			name: "Authenticated user with different role",
			authContext: authentication.AuthContext{
				Authenticated: true,
				GlobalRole:    types.GlobalRoleUser,
			},
			requireRoles:       []types.GlobalRole{types.GlobalRoleSystemAdministrator},
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "Authentication failure: insufficient role",
		},
		{
			name: "Authenticated user with no required role",
			authContext: authentication.AuthContext{
				Authenticated: true,
				GlobalRole:    types.GlobalRoleUser,
			},
			requireRoles:       []types.GlobalRole{},
			expectedStatusCode: http.StatusOK,
			expectedBody:       "success",
		},
		{
			name: "Authentication ignored with unauthenticated and no matching role",
			authContext: authentication.AuthContext{
				Authenticated:             false,
				IgnoreAuthenticationState: true,
				GlobalRole:                types.GlobalRoleUser,
			},
			requireRoles:       []types.GlobalRole{types.GlobalRoleSystemAdministrator},
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "Authentication failure: insufficient role",
		},
		{
			name: "Authentication ignored with unauthenticated and no role",
			authContext: authentication.AuthContext{
				Authenticated:             false,
				IgnoreAuthenticationState: true,
			},
			requireRoles:       []types.GlobalRole{types.GlobalRoleSystemAdministrator},
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "Authentication failure: insufficient role",
		},
		{
			name: "Authentication ignored with authenticated and no role",
			authContext: authentication.AuthContext{
				Authenticated:             true,
				IgnoreAuthenticationState: true,
			},
			requireRoles:       []types.GlobalRole{types.GlobalRoleSystemAdministrator},
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "Authentication failure: insufficient role",
		},
		{
			name: "Authentication ignored with authenticated and no matching role",
			authContext: authentication.AuthContext{
				Authenticated:             true,
				IgnoreAuthenticationState: true,
				GlobalRole:                types.GlobalRoleUser,
			},
			requireRoles:       []types.GlobalRole{types.GlobalRoleSystemAdministrator},
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "Authentication failure: insufficient role",
		},
		{
			name: "Not authenticated with no specific error",
			authContext: authentication.AuthContext{
				Authenticated: false,
			},
			requireRoles:       []types.GlobalRole{types.GlobalRoleSystemAdministrator},
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "Authentication required",
		},
		{
			name: "Not authenticated with specific error",
			authContext: authentication.AuthContext{
				Authenticated: false,
				Error:         errors.New("token expired"),
			},
			requireRoles:       []types.GlobalRole{types.GlobalRoleSystemAdministrator},
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "Authentication failure: token expired",
		},
		{
			name: "User with one of multiple required roles",
			authContext: authentication.AuthContext{
				Authenticated: true,
				GlobalRole:    types.GlobalRoleSystemAdministrator,
			},
			requireRoles:       []types.GlobalRole{types.GlobalRoleSystemAdministrator, types.GlobalRoleAPI},
			expectedStatusCode: http.StatusOK,
			expectedBody:       "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new request
			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil)
			assert.NoError(t, err)

			// Add the auth context to the request
			ctx := authentication.WithAuthContext(context.Background(), tt.authContext)
			req = req.WithContext(ctx)

			// Create a response recorder
			recorder := httptest.NewRecorder()

			// Create the middleware with the specified required roles
			mw := middleware.RequireAuthenticationMiddleware(nextHandler, tt.requireRoles...)

			// Serve the request through the middleware
			mw.ServeHTTP(recorder, req)

			// Assert the response
			assert.Equal(t, tt.expectedStatusCode, recorder.Code)
			assert.Equal(t, tt.expectedBody, recorder.Body.String())
		})
	}
}

// TestRequireAuthenticationMiddlewareXSSPrevention tests that the middleware properly escapes error messages
func TestRequireAuthenticationMiddlewareXSSPrevention(t *testing.T) {
	// Create a mock handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create a request with an auth context containing an XSS attempt in the error
	xssError := errors.New("<script>alert('XSS')</script>")
	authCtx := authentication.AuthContext{
		Authenticated: false,
		Error:         xssError,
	}

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil)
	assert.NoError(t, err)

	ctx := authentication.WithAuthContext(context.Background(), authCtx)
	req = req.WithContext(ctx)

	// Create a response recorder
	recorder := httptest.NewRecorder()

	// Create and serve the middleware
	mw := middleware.RequireAuthenticationMiddleware(nextHandler)
	mw.ServeHTTP(recorder, req)

	// Check that the script tags are escaped
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	assert.Equal(t, "Authentication failure: &lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;", recorder.Body.String())
	assert.NotContains(t, recorder.Body.String(), "<script>")
}

// TestRequireAuthenticationMiddlewarePanic tests that the middleware handles the panic case
// when no auth context is in the request
func TestRequireAuthenticationMiddlewarePanic(t *testing.T) {
	// Create a mock handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create a request with no auth context
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil)
	assert.NoError(t, err)

	// Create a response recorder
	recorder := httptest.NewRecorder()

	// Create the middleware
	mw := middleware.RequireAuthenticationMiddleware(nextHandler)

	// The middleware should panic when no auth context is present
	assert.Panics(t, func() {
		mw.ServeHTTP(recorder, req)
	})
}
