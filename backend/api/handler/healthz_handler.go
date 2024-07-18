package handler

import (
	"fmt"
	"html"
	"net/http"

	logger "github.com/apex/log"
	"github.com/friendsofgo/errors"
)

type DBPinger interface {
	Ping() error
}

func NewHealthzHandler(dbPinger DBPinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logger.FromContext(r.Context())

		ignoreErrors := r.URL.Query().Get("ignore_errors") == "1"

		if err := dbPinger.Ping(); err != nil {
			log.
				WithError(errors.WithStack(err)).
				WithField("handler", "healthz").
				Error("Could not connect to database")

			respondErr(w, ignoreErrors, "could not connect to database")
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "OK")
	}
}

func respondErr(w http.ResponseWriter, ignoreErrors bool, message string) {
	if ignoreErrors {
		w.WriteHeader(http.StatusOK)
		// nosemgrep: go.lang.security.audit.xss.no-fprintf-to-responsewriter.no-fprintf-to-responsewriter
		_, _ = fmt.Fprintf(w, "WARN: %s\n", html.EscapeString(message))
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	// nosemgrep: go.lang.security.audit.xss.no-fprintf-to-responsewriter.no-fprintf-to-responsewriter
	_, _ = fmt.Fprintf(w, "ERROR: %s\n", html.EscapeString(message))
}
