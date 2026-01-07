package basic_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"myvendor.mytld/myproject/backend/security/authentication/basic"
)

func TestSimpleAuth(t *testing.T) {
	const (
		configuredUsername = "admin"
		configuredPassword = "secret"
	)

	nextHandlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextHandlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name               string
		requestUsername    string
		requestPassword    string
		setBasicAuth       bool
		expectedStatus     int
		expectWWWAuth      bool
		expectNextHandlerCalled bool
	}{
		{
			name:               "valid credentials",
			requestUsername:    "admin",
			requestPassword:    "secret",
			setBasicAuth:       true,
			expectedStatus:     http.StatusOK,
			expectWWWAuth:      false,
			expectNextHandlerCalled: true,
		},
		{
			name:               "invalid username",
			requestUsername:    "wrong",
			requestPassword:    "secret",
			setBasicAuth:       true,
			expectedStatus:     http.StatusUnauthorized,
			expectWWWAuth:      true,
			expectNextHandlerCalled: false,
		},
		{
			name:               "invalid password",
			requestUsername:    "admin",
			requestPassword:    "wrong",
			setBasicAuth:       true,
			expectedStatus:     http.StatusUnauthorized,
			expectWWWAuth:      true,
			expectNextHandlerCalled: false,
		},
		{
			name:               "both invalid",
			requestUsername:    "wrong",
			requestPassword:    "wrong",
			setBasicAuth:       true,
			expectedStatus:     http.StatusUnauthorized,
			expectWWWAuth:      true,
			expectNextHandlerCalled: false,
		},
		{
			name:               "missing auth header",
			requestUsername:    "",
			requestPassword:    "",
			setBasicAuth:       false,
			expectedStatus:     http.StatusUnauthorized,
			expectWWWAuth:      true,
			expectNextHandlerCalled: false,
		},
		{
			name:               "empty credentials",
			requestUsername:    "",
			requestPassword:    "",
			setBasicAuth:       true,
			expectedStatus:     http.StatusUnauthorized,
			expectWWWAuth:      true,
			expectNextHandlerCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandlerCalled = false

			handler := basic.SimpleAuth(nextHandler, configuredUsername, configuredPassword)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.setBasicAuth {
				req.SetBasicAuth(tt.requestUsername, tt.requestPassword)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			require.Equal(t, tt.expectedStatus, rec.Code, "response status")
			assert.Equal(t, tt.expectNextHandlerCalled, nextHandlerCalled, "next handler should be called")

			if tt.expectWWWAuth {
				assert.Equal(t, `Basic realm="Restricted"`, rec.Header().Get("WWW-Authenticate"), "should set WWW-Authenticate header")
			} else {
				assert.Empty(t, rec.Header().Get("WWW-Authenticate"), "should not set WWW-Authenticate header")
			}
		})
	}
}
