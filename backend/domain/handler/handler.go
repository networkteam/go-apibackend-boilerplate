package handler

import (
	"database/sql"

	"go.opentelemetry.io/otel/metric"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/jobqueue"
	"myvendor.mytld/myproject/backend/mail"
)

type Handler struct {
	db         *sql.DB
	timeSource types.TimeSource
	mailer     *mail.Mailer
	queue      jobqueue.Queue
	config     domain.Config

	instrumentation instrumentation
}

type Deps struct {
	TimeSource    types.TimeSource
	Mailer        *mail.Mailer
	Queue         jobqueue.Queue
	MeterProvider metric.MeterProvider
}

func NewHandler(db *sql.DB, config domain.Config, deps Deps) *Handler {
	return &Handler{
		db:              db,
		config:          config,
		timeSource:      deps.TimeSource,
		mailer:          deps.Mailer,
		queue:           deps.Queue,
		instrumentation: initInstrumentation(deps.MeterProvider),
	}
}
