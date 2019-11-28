package api

import (
	"database/sql"

	"myvendor/myproject/backend/domain"
	"myvendor/myproject/backend/persistence/finder"
	"myvendor/myproject/backend/service/hub"
	"myvendor/myproject/backend/service/notification"
)

// ResolverDependencies provides common dependencies for api resolvers
type ResolverDependencies struct {
	Db         *sql.DB
	TimeSource domain.TimeSource
	Hub        *hub.Hub
	Notifier   notification.Notifier
}

func (r ResolverDependencies) Finder() *finder.Finder {
	return finder.NewFinder(r.Db)
}
