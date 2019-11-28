package authentication

import "net/http"

const (
	authTokenCookieName        = "authToken"
	refreshCsrfTokenHeaderName = "X-Refresh-CSRF-Token"
	authTokenHeaderName        = "Authorization"
)

func SetRefreshCsrfTokenHeader(w http.ResponseWriter, csrfToken string) {
	w.Header().Set(refreshCsrfTokenHeaderName, csrfToken)
}

func SetAuthTokenCookie(w http.ResponseWriter, r *http.Request, authToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     authTokenCookieName,
		Value:    authToken,
		HttpOnly: true,
		Secure:   r.URL.Scheme == "https",
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
	var ok bool
	if authToken, ok = getAuthTokenFromCookie(r); !ok {
		authToken = getAuthTokenFromHeader(r)
		// Also skip CSRF check if Authorization header is present, since it cannot be "faked"
		if authToken != "" {
			skipCsrfCheck = true
		}
	}
	return authToken, skipCsrfCheck
}

func getAuthTokenFromCookie(r *http.Request) (token string, ok bool) {
	if cookie, err := r.Cookie(authTokenCookieName); err == nil {
		return cookie.Value, true
	}
	return "", false
}

func getAuthTokenFromHeader(r *http.Request) string {
	return r.Header.Get(authTokenHeaderName)
}

func isMethodSafe(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}
