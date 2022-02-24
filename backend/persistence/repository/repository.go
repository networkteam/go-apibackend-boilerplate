package repository

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/networkteam/construct"
)

var ErrNotFound = construct.ErrNotFound

type QueryRower interface {
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func queryBuilder(runner squirrel.BaseRunner) squirrel.StatementBuilderType {
	return squirrel.StatementBuilder.
		PlaceholderFormat(squirrel.Dollar).
		RunWith(runner)
}
