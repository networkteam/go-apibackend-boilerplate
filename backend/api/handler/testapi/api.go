package testapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"log/slog"
	"net/http"

	"github.com/networkteam/slogutils"
)

type API struct {
	db  *sql.DB
	mux http.Handler
}

var _ http.Handler = &API{}

func NewAPI(db *sql.DB) *API {
	mux := http.NewServeMux()
	apiHandler := &API{
		db:  db,
		mux: mux,
	}

	// mux.HandleFunc("POST /fixtures/accounts", apiHandler.accountCreate)
	// mux.HandleFunc("DELETE /fixtures/accounts", apiHandler.accountDelete)

	return apiHandler
}

func (t *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.mux.ServeHTTP(w, r)
}

func (t *API) handleInternalError(logger *slog.Logger, w http.ResponseWriter, message string, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = fmt.Fprintf(w, "ERROR: %s\n", html.EscapeString(err.Error()))
	logger.Error(message, slogutils.Err(err))
}

func (t *API) sendJSONResponse(logger *slog.Logger, w http.ResponseWriter, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		t.handleInternalError(logger, w, "Error encoding response", err)
	}
}
