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
}

func (r ResolverDependencies) Handler() *handler.Handler {
	return handler.NewHandler(r.DB, r.Config, handler.Deps{
		TimeSource: r.TimeSource,
		Mailer:     r.Mailer,
	})
}

func (r ResolverDependencies) Finder() *finder.Finder {
	return finder.NewFinder(r.DB, r.TimeSource)
}
