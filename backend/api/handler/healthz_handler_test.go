package handler_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"myvendor.mytld/myproject/backend/api/handler"
)

type dbPinger struct {
	called    bool
	returnErr error
}

func (d *dbPinger) Ping() error {
	d.called = true
	return d.returnErr
}

func TestNewHealthzHandler_Ok(t *testing.T) {
	db := &dbPinger{}

	h := handler.NewHealthzHandler(db)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, db.called, "db.Ping() should be called")
}

func TestNewHealthzHandler_WithDBErr(t *testing.T) {
	db := &dbPinger{
		returnErr: fmt.Errorf("pong failure"),
	}

	h := handler.NewHealthzHandler(db)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.True(t, db.called, "db.Ping() should be called")
}
