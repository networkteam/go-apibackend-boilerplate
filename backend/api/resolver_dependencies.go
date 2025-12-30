package api

import (
	"database/sql"

	"go.opentelemetry.io/otel/metric"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/domain/handler"
	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/jobqueue"
	"myvendor.mytld/myproject/backend/mail"
)

// ResolverDependencies provides common dependencies for api resolvers
type ResolverDependencies struct {
	Config        domain.Config
	DB            *sql.DB
	TimeSource    types.TimeSource
	Mailer        *mail.Mailer
	Queue         jobqueue.Queue
	MeterProvider metric.MeterProvider
}

func (r ResolverDependencies) Handler() *handler.Handler {
	return handler.NewHandler(r.DB, r.Config, handler.Deps{
		TimeSource:    r.TimeSource,
		Mailer:        r.Mailer,
		Queue:         r.Queue,
		MeterProvider: r.MeterProvider,
	})
}
