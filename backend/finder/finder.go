package finder

import (
	"context"
	"database/sql"
	std_errors "errors"

	"github.com/friendsofgo/errors"
	"github.com/networkteam/qrb/qrbsql"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/persistence/repository"
)

// Finder is a higher level executor for queries that includes authorization.
type Finder struct {
	executor   qrbsql.Executor
	timeSource domain.TimeSource
}

// NewFinder creates a new Finder.
func NewFinder(db *sql.DB, timeSource domain.TimeSource) *Finder {
	return &Finder{
		executor:   db,
		timeSource: timeSource,
	}
}

var errTransactionalNoSQLDB = std_errors.New("finder: executor for Transactional must be a *sql.DB")

func (f *Finder) Transactional(ctx context.Context, isolationLevel sql.IsolationLevel, callback func(txFinder *Finder) error) error {
	db, ok := f.executor.(*sql.DB)
	if !ok {
		return errors.WithStack(errTransactionalNoSQLDB)
	}

	return repository.TransactionalWithOpts(ctx, db, &sql.TxOptions{
		ReadOnly:  true,
		Isolation: isolationLevel,
	}, func(tx *sql.Tx) error {
		txFinder := &Finder{
			executor:   tx,
			timeSource: f.timeSource,
		}
		err := callback(txFinder)
		return err
	})
}
