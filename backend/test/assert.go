package test

import (
	"net/http"
	"testing"
)

// RequireHasCookie finds a cookie by name in the given cookies slice.
// Fails the test immediately if the cookie is not found or has an empty value.
// Returns the cookie (guaranteed non-nil with non-empty value).
func RequireHasCookie(t *testing.T, cookies []*http.Cookie, name string) *http.Cookie {
	t.Helper()
	for _, cookie := range cookies {
		if cookie.Name == name {
			if cookie.Value == "" {
				t.Fatalf("cookie %q has empty value", name)
			}
			return cookie
		}
	}
	t.Fatalf("cookie %q not found", name)
	return nil // unreachable
}
