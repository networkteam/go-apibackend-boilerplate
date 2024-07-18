package graphql

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type GenericResult struct {
	Data struct {
		Result struct {
			Error *FieldsError
		}
	}
	GraphqlErrors
}

func RequireResponseCookie(t *testing.T, resp *http.Response, cookieName string) *http.Cookie {
	t.Helper()

	cookies := resp.Cookies()
	var foundCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == cookieName {
			foundCookie = cookie
		}
	}
	require.NotNil(t, foundCookie, "authToken cookie should be set")
	return foundCookie
}

func ToPtr[T any](v T) *T {
	return &v
}

func TimeFromString(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return &t
}

func IncludeFragments(s string, fragments ...string) string {
	return strings.Join(append([]string{s}, fragments...), "\n")
}
