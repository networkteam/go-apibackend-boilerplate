package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/friendsofgo/errors"
)

const (
	SortOrderAsc  = "ASC"
	SortOrderDesc = "DESC"

	// MaxPerPage is 1000 for React Admin exporter
	MaxPerPage     = 1000
	DefaultPerPage = 50

	DefaultSortOrder = SortOrderAsc
)

type QueryRowerContext interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type Paging struct {
	Page      int
	PerPage   *int
	SortField *string
	SortOrder *string
}

const (
	pgErrCode_integrity_constraint_violation = "23000"
	pgErrCode_restrict_violation             = "23001"
	pgErrCode_not_null_violation             = "23502"
	pgErrCode_foreign_key_violation          = "23503"
	pgErrCode_unique_violation               = "23505"
	pgErrCode_check_violation                = "23514"
	pgErrCode_exclusion_violation            = "23P01"
)

func applyPaging(query squirrel.SelectBuilder, paging Paging, sortFieldMapping map[string]string) (squirrel.SelectBuilder, error) {
	if paging.PerPage != nil {
		perPage := *paging.PerPage

		if perPage > MaxPerPage {
			return query, errors.New("max per page exceeded")
		}

		query = query.
			Limit(uint64(perPage)).
			Offset(uint64(paging.Page * perPage))
	}

	if paging.SortField != nil {
		sortOrder := DefaultSortOrder
		if paging.SortOrder != nil {
			sortOrder = *paging.SortOrder
		}
		sortField := *paging.SortField
		col, ok := sortFieldMapping[strings.ToLower(sortField)]
		if !ok {
			return query, errors.Errorf("invalid sort field: %s", sortField)
		}

		nullsOrder := "FIRST"
		if sortOrder == SortOrderDesc {
			nullsOrder = "LAST"
		}
		query = query.
			OrderBy(fmt.Sprintf("%s %s NULLS %s", col, sortOrder, nullsOrder))
	}

	return query, nil
}

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

func assertRowsAffected(res sql.Result, op string) error {
	const numberOfRows = 1
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "getting affected rows")
	}
	if rowsAffected != numberOfRows {
		return errors.Errorf("%s affected %d rows, but expected exactly %d", op, rowsAffected, numberOfRows)
	}
	return err
}
