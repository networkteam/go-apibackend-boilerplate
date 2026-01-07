package basic

import (
	"crypto/subtle"
	"net/http"
)

func SimpleAuth(next http.Handler, username, password string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualUsername, actualPassword, ok := r.BasicAuth()
		usernameMatch := subtle.ConstantTimeCompare([]byte(actualUsername), []byte(username))
		passwordMatch := subtle.ConstantTimeCompare([]byte(actualPassword), []byte(password))
		if (usernameMatch+passwordMatch) != 2 || !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
