package finder

import (
	"database/sql"

	"myvendor.mytld/myproject/backend/persistence/records"
)

type Finder struct {
	db *sql.DB

	organisationStore *records.OrganisationStore
}

func NewFinder(db *sql.DB) *Finder {
	return &Finder{
		db:                db,
		organisationStore: records.NewOrganisationStore(db),
	}
}
