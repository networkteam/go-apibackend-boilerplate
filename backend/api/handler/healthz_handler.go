package handler

import (
	"fmt"
	"html"
	"net/http"

	"github.com/friendsofgo/errors"
	"github.com/networkteam/slogutils"
)

type DBPinger interface {
	Ping() error
}

func NewHealthzHandler(dbPinger DBPinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := slogutils.FromContext(ctx)

		ignoreErrors := r.URL.Query().Get("ignore_errors") == "1"

		if err := dbPinger.Ping(); err != nil {
			logger.ErrorContext(ctx, "Could not connect to database",
				"handler", "healthz",
				slogutils.Err(errors.WithStack(err)))

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
