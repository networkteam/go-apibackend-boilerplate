package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/friendsofgo/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/networkteam/qrb/builder"
)

const (
	SortOrderAsc  = "ASC"
	SortOrderDesc = "DESC"
)

type QueryRowerContext interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

//nolint:revive // Better readability with underscores
const (
	pgErrCode_integrity_constraint_violation = "23000"
	pgErrCode_restrict_violation             = "23001"
	pgErrCode_not_null_violation             = "23502"
	pgErrCode_foreign_key_violation          = "23503"
	pgErrCode_unique_violation               = "23505"
	pgErrCode_check_violation                = "23514"
	pgErrCode_exclusion_violation            = "23P01"
)

type TxBeginner interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

func Transactional(ctx context.Context, proxy TxBeginner, f func(tx *sql.Tx) error) (err error) {
	tx, err := proxy.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "opening transaction")
	}

	if err = f(tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Wrapf(rollbackErr, "rolling back transaction after: %v", err)
		}

		return err
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "committing transaction")
	}

	return nil
}

func TransactionalWithOpts(ctx context.Context, proxy TxBeginner, opts *sql.TxOptions, f func(tx *sql.Tx) error) (err error) {
	tx, err := proxy.BeginTx(ctx, opts)
	if err != nil {
		return errors.Wrap(err, "opening transaction")
	}

	if err = f(tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return multierror.Append(err, errors.Wrap(rollbackErr, "rolling back transaction after error"))
		}

		return err
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "committing transaction")
	}

	return nil
}

// --- Paging options

type PagingOption func(query builder.SelectBuilder, sortFieldMapping map[string]builder.IdentExp) (builder.SelectBuilder, error)

func WithSort(field, order string) PagingOption {
	return func(query builder.SelectBuilder, sortFieldMapping map[string]builder.IdentExp) (builder.SelectBuilder, error) {
		col, ok := sortFieldMapping[strings.ToLower(field)]
		if !ok {
			return query, fmt.Errorf("invalid sort field: %s", field)
		}

		orderByBuilder := query.OrderBy(col)
		if order == SortOrderDesc {
			query = orderByBuilder.Desc().NullsLast().SelectBuilder
		} else {
			query = orderByBuilder.Asc().NullsFirst().SelectBuilder
		}
		return query, nil
	}
}

func WithLimit(limit int) PagingOption {
	return func(query builder.SelectBuilder, sortFieldMapping map[string]builder.IdentExp) (builder.SelectBuilder, error) {
		return query.Limit(builder.Arg(limit)), nil
	}
}

func WithOffset(offset int) PagingOption {
	return func(query builder.SelectBuilder, sortFieldMapping map[string]builder.IdentExp) (builder.SelectBuilder, error) {
		return query.Offset(builder.Arg(offset)), nil
	}
}

func applyPagingOptions(query builder.SelectBuilder, opts []PagingOption, sortFieldMapping map[string]builder.IdentExp) (builder.SelectBuilder, error) {
	for _, opt := range opts {
		var err error
		query, err = opt(query, sortFieldMapping)
		if err != nil {
			return query, err
		}
	}
	return query, nil
}
