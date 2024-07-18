package repository

import (
	"context"
	"database/sql"

	"github.com/networkteam/construct/v2"
)

var ErrNotFound = construct.ErrNotFound

type QueryRower interface {
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}
