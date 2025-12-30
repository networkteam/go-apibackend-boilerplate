package authentication

import (
	"net/http"
)

const (
	accessTokenCookieName = "access_token"
	accessTokenHeaderName = "Authorization"
	//#nosec G101 -- This constant is only the header name
	csrfTokenCookieName = "csrf_token"
	//#nosec G101 -- This constant is only the header name
	csrfTokenHeaderName = "X-CSRF-Token"
)

func SetAccessTokenCookie(w http.ResponseWriter, r *http.Request, accessToken string) {
	// nosemgrep: go.lang.security.audit.net.cookie-missing-secure.cookie-falsessing-secure
	http.SetCookie(w, &http.Cookie{
		Name:     accessTokenCookieName,
		Value:    accessToken,
		HttpOnly: true,
		Path:     "/",
		Secure:   r.URL.Scheme == "https",
		SameSite: http.SameSiteLaxMode,
	})
}

func SetCsrfTokenCookie(w http.ResponseWriter, r *http.Request, csrfToken string) {
	// nosemgrep: go.lang.security.audit.net.cookie-missing-secure.cookie-missing-secure
	http.SetCookie(w, &http.Cookie{
		Name:     csrfTokenCookieName,
		Value:    csrfToken,
		HttpOnly: false,
		Path:     "/",
		Secure:   r.URL.Scheme == "https",
		SameSite: http.SameSiteLaxMode,
	})
}

func getAccessTokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie(accessTokenCookieName)
	if err != nil {
		return ""
	}

	return cookie.Value
}

func GetCsrfTokenFromHeader(r *http.Request) string {
	return r.Header.Get(csrfTokenHeaderName)
}

func getAccessTokenFromHeader(r *http.Request) string {
	headerValue := r.Header.Get(accessTokenHeaderName)

	if len(headerValue) > 7 && headerValue[:7] == "Bearer " {
		return headerValue[7:]
	}

	return ""
}

func GetAccessTokenAndSkipCsrfCheckFromRequest(r *http.Request) (accessToken string, skipCsrfCheck bool) {
	// Require no CSRF check for safe methods
	if isMethodSafe(r.Method) {
		skipCsrfCheck = true
	}

	// First check if access token is sent as header
	accessToken = getAccessTokenFromHeader(r)
	if accessToken != "" {
		// Also skip CSRF check if Authorization header is present, since it cannot be "faked"
		return accessToken, true
	}

	// Otherwise, use access token from cookie
	accessToken = getAccessTokenFromCookie(r)

	return accessToken, skipCsrfCheck
}

func DeleteAccessTokenCookie(w http.ResponseWriter, r *http.Request) {
	SetAccessTokenCookie(w, r, "")
}

func DeleteCsrfTokenCookie(w http.ResponseWriter, r *http.Request) {
	SetCsrfTokenCookie(w, r, "")
}

func isMethodSafe(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}
