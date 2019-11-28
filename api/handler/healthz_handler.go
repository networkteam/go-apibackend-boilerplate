package handler

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/apex/log"
	"github.com/pkg/errors"
)

func NewHealthzHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if err := db.Ping(); err != nil {
			log.WithError(errors.WithStack(err)).WithField("handler", "healthz").Error("Could not connect to database")

			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprintln(w, "Internal server error: could not connect to database")
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "OK")
	}
}
