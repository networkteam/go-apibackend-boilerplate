package handler

import (
	"database/sql"

	"github.com/apex/log"
	sentry "github.com/getsentry/sentry-go"
	"github.com/robfig/cron"

	"myvendor/myproject/backend/domain"
)

type AppAccountTokenCleanupJob struct {
	db         *sql.DB
	timeSource domain.TimeSource
}

var _ cron.Job = new(AppAccountTokenCleanupJob)

func NewAppAccountTokenCleanupJob(db *sql.DB, timeSource domain.TimeSource) AppAccountTokenCleanupJob {
	return AppAccountTokenCleanupJob{
		db:         db,
		timeSource: timeSource,
	}
}

func (j AppAccountTokenCleanupJob) Run() {
	defer sentry.Recover()

	query := "DELETE FROM app_account_request_tokens WHERE expiry < $1"
	res, err := j.db.Exec(query, j.timeSource.Now())
	if err != nil {
		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetTag("section", "cron")
			sentry.CaptureException(err)
		})

		return
	}

	deletedTokens, _ := res.RowsAffected()

	log.
		WithField("deletedTokens", deletedTokens).
		Debug("Finished app account request token cleanup job")
}
