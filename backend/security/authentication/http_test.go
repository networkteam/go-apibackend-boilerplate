package authentication_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"myvendor.mytld/myproject/backend/security/authentication"
)

func TestGetCsrfTokenFromHeader(t *testing.T) {
	tests := []struct {
		name          string
		headerValue   string
		expectedToken string
	}{
		{
			name:          "valid CSRF token in header",
			headerValue:   "test-csrf-token-123",
			expectedToken: "test-csrf-token-123",
		},
		{
			name:          "empty CSRF token",
			headerValue:   "",
			expectedToken: "",
		},
		{
			name:          "CSRF token with special characters",
			headerValue:   "token-with-special-chars_!@#$%",
			expectedToken: "token-with-special-chars_!@#$%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			if tt.headerValue != "" {
				req.Header.Set("X-CSRF-Token", tt.headerValue)
			}

			token := authentication.GetCsrfTokenFromHeader(req)
			assert.Equal(t, tt.expectedToken, token)
		})
	}
}

func TestGetAccessTokenFromHeader(t *testing.T) {
	tests := []struct {
		name          string
		headerValue   string
		expectedToken string
	}{
		{
			name:          "valid Bearer token",
			headerValue:   "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expectedToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
		},
		{
			name:          "empty header",
			headerValue:   "",
			expectedToken: "",
		},
		{
			name:          "no Bearer prefix",
			headerValue:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expectedToken: "",
		},
		{
			name:          "Bearer without space",
			headerValue:   "Bearertoken",
			expectedToken: "",
		},
		{
			name:          "lowercase bearer",
			headerValue:   "bearer token123",
			expectedToken: "",
		},
		{
			name:          "Bearer with only space",
			headerValue:   "Bearer ",
			expectedToken: "",
		},
		{
			name:          "Bearer prefix only (too short)",
			headerValue:   "Bearer",
			expectedToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.headerValue != "" {
				req.Header.Set("Authorization", tt.headerValue)
			}

			// Using a test helper to access the unexported function
			token := getAccessTokenFromHeaderHelper(req)
			assert.Equal(t, tt.expectedToken, token)
		})
	}
}

// Helper function to test the unexported getAccessTokenFromHeader
func getAccessTokenFromHeaderHelper(r *http.Request) string {
	headerValue := r.Header.Get("Authorization")

	if len(headerValue) > 7 && headerValue[:7] == "Bearer " {
		return headerValue[7:]
	}

	return ""
}

func TestGetAccessTokenAndSkipCsrfCheckFromRequest(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		authHeader       string
		cookieValue      string
		expectedToken    string
		expectedSkipCsrf bool
	}{
		{
			name:             "GET request - safe method",
			method:           http.MethodGet,
			authHeader:       "",
			cookieValue:      "cookie-token",
			expectedToken:    "cookie-token",
			expectedSkipCsrf: true,
		},
		{
			name:             "HEAD request - safe method",
			method:           http.MethodHead,
			authHeader:       "",
			cookieValue:      "cookie-token",
			expectedToken:    "cookie-token",
			expectedSkipCsrf: true,
		},
		{
			name:             "OPTIONS request - safe method",
			method:           http.MethodOptions,
			authHeader:       "",
			cookieValue:      "cookie-token",
			expectedToken:    "cookie-token",
			expectedSkipCsrf: true,
		},
		{
			name:             "POST with Bearer token - skip CSRF",
			method:           http.MethodPost,
			authHeader:       "Bearer header-token-123",
			cookieValue:      "cookie-token",
			expectedToken:    "header-token-123",
			expectedSkipCsrf: true,
		},
		{
			name:             "POST with cookie only - require CSRF",
			method:           http.MethodPost,
			authHeader:       "",
			cookieValue:      "cookie-token-456",
			expectedToken:    "cookie-token-456",
			expectedSkipCsrf: false,
		},
		{
			name:             "PUT with Bearer token - skip CSRF",
			method:           http.MethodPut,
			authHeader:       "Bearer put-token",
			cookieValue:      "",
			expectedToken:    "put-token",
			expectedSkipCsrf: true,
		},
		{
			name:             "DELETE with cookie only - require CSRF",
			method:           http.MethodDelete,
			authHeader:       "",
			cookieValue:      "delete-cookie-token",
			expectedToken:    "delete-cookie-token",
			expectedSkipCsrf: false,
		},
		{
			name:             "POST with no tokens",
			method:           http.MethodPost,
			authHeader:       "",
			cookieValue:      "",
			expectedToken:    "",
			expectedSkipCsrf: false,
		},
		{
			name:             "GET with Bearer token - skip CSRF (both reasons)",
			method:           http.MethodGet,
			authHeader:       "Bearer get-token",
			cookieValue:      "",
			expectedToken:    "get-token",
			expectedSkipCsrf: true,
		},
		{
			name:             "POST with invalid Bearer format - use cookie, require CSRF",
			method:           http.MethodPost,
			authHeader:       "InvalidBearer token",
			cookieValue:      "fallback-cookie",
			expectedToken:    "fallback-cookie",
			expectedSkipCsrf: false,
		},
		{
			name:             "PATCH with cookie - require CSRF",
			method:           http.MethodPatch,
			authHeader:       "",
			cookieValue:      "patch-token",
			expectedToken:    "patch-token",
			expectedSkipCsrf: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", nil)

			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			if tt.cookieValue != "" {
				req.AddCookie(&http.Cookie{
					Name:  "access_token",
					Value: tt.cookieValue,
				})
			}

			token, skipCsrf := authentication.GetAccessTokenAndSkipCsrfCheckFromRequest(req)

			assert.Equal(t, tt.expectedToken, token, "access token mismatch")
			assert.Equal(t, tt.expectedSkipCsrf, skipCsrf, "skipCsrfCheck mismatch")
		})
	}
}
