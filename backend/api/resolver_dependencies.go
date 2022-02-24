package api

import (
	"database/sql"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/finder"
	"myvendor.mytld/myproject/backend/handler"
	"myvendor.mytld/myproject/backend/mail"
)

// ResolverDependencies provides common dependencies for api resolvers
type ResolverDependencies struct {
	DB         *sql.DB
	TimeSource domain.TimeSource
	Mailer     *mail.Mailer
	Config     domain.Config

	// Filestore   filestore.Filestore
}

func (r ResolverDependencies) Handler() *handler.Handler {
	return handler.NewHandler(r.DB, r.TimeSource, r.Mailer, r.Config)
}

func (r ResolverDependencies) Finder() *finder.Finder {
	return finder.NewFinder(r.DB, r.TimeSource)
}
