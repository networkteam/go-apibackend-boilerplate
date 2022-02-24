package finder

import (
	"database/sql"

	"myvendor.mytld/myproject/backend/domain"
)

// Finder is a higher level executor for queries that includes authorization.
type Finder struct {
	db         *sql.DB
	timeSource domain.TimeSource
}

// NewFinder creates a new Finder.
func NewFinder(db *sql.DB, timeSource domain.TimeSource) *Finder {
	return &Finder{
		db:         db,
		timeSource: timeSource,
	}
}
