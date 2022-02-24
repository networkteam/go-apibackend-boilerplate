package authentication

import (
	"net/http"
)

const (
	authTokenCookieName        = "authToken"
	refreshAuthTokenHeaderName = "X-Refresh-Auth-Token"
	refreshCsrfTokenHeaderName = "X-Refresh-CSRF-Token"
	authTokenHeaderName        = "Authorization"
)

func SetRefreshAuthTokenHeader(w http.ResponseWriter, authToken string) {
	w.Header().Set(refreshAuthTokenHeaderName, authToken)
}

func SetRefreshCsrfTokenHeader(w http.ResponseWriter, csrfToken string) {
	w.Header().Set(refreshCsrfTokenHeaderName, csrfToken)
}

func SetAuthTokenCookie(w http.ResponseWriter, r *http.Request, authToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     authTokenCookieName,
		Value:    authToken,
		HttpOnly: true,
		Secure:   r.URL.Scheme == "https",
		SameSite: http.SameSiteStrictMode,
	})
}

func DeleteAuthTokenCookie(w http.ResponseWriter, r *http.Request) {
	SetAuthTokenCookie(w, r, "")
}

func GetAuthTokenAndSkipCsrfCheckFromRequest(r *http.Request) (authToken string, skipCsrfCheck bool) {
	// Require no CSRF check for safe methods
	if isMethodSafe(r.Method) {
		skipCsrfCheck = true
	}

	// First check if auth token is sent as header
	authToken = getAuthTokenFromHeader(r)
	if authToken != "" {
		// Also skip CSRF check if Authorization header is present, since it cannot be "faked"
		return authToken, true
	}

	// Otherwise use auth token from cookie
	authToken = getAuthTokenFromCookie(r)
	return authToken, skipCsrfCheck
}

func getAuthTokenFromCookie(r *http.Request) string {
	if cookie, err := r.Cookie(authTokenCookieName); err == nil {
		return cookie.Value
	}
	return ""
}

func getAuthTokenFromHeader(r *http.Request) string {
	return r.Header.Get(authTokenHeaderName)
}

func isMethodSafe(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}
