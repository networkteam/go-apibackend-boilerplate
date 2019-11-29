package api

import (
	"database/sql"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/persistence/finder"
	"myvendor.mytld/myproject/backend/service/hub"
	"myvendor.mytld/myproject/backend/service/notification"
)

// ResolverDependencies provides common dependencies for api resolvers
type ResolverDependencies struct {
	Db         *sql.DB
	TimeSource domain.TimeSource
	Hub        hub.Hub
	Notifier   notification.Notifier
}

func (r ResolverDependencies) Finder() *finder.Finder {
	return finder.NewFinder(r.Db)
}
