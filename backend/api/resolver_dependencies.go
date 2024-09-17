package api

import (
	"database/sql"

	"go.opentelemetry.io/otel/metric"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/mail"
)

// ResolverDependencies provides common dependencies for api resolvers
type ResolverDependencies struct {
	Config        domain.Config
	DB            *sql.DB
	TimeSource    domain.TimeSource
	MeterProvider metric.MeterProvider
	Mailer        *mail.Mailer
}
