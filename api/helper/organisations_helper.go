package helper

import (
	"database/sql"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/zbyte/go-kallax"

	"myvendor/myproject/backend/api"
	"myvendor/myproject/backend/persistence/records"
)

func MapToOrganisation(o *records.Organisation) (api.Organisation, error) {
	return api.Organisation{
		ID:   uuid.UUID(o.ID),
		Name: o.Name,
	}, nil
}

func MapResultSetToOrganisations(res *records.OrganisationResultSet) (orgs []api.Organisation, err error) {
	err = res.ForEach(func(organisation *records.Organisation) error {
		org, err := MapToOrganisation(organisation)
		if err != nil {
			return errors.Wrap(err, "could not map to organisation")
		}
		orgs = append(orgs, org)
		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "mapping organisations")
	}

	return
}

func GetOrganisationForAccount(account *records.Account, db *sql.DB) (*api.Organisation, error) {
	organisationStore := records.NewOrganisationStore(db)
	var organisation *api.Organisation
	if account.OrganisationID != nil {
		organisationID := *account.OrganisationID
		organisationDb, err := organisationStore.FindOne(records.NewOrganisationQuery().FindByID(organisationID))
		if err != nil && err != kallax.ErrNotFound {
			return nil, errors.Wrap(err, "could not query organisation")
		}

		if organisationDb != nil {
			mappedOrganisation, err := MapToOrganisation(organisationDb)
			if err != nil {
				return nil, errors.Wrap(err, "failed to map organisation")
			}
			organisation = &mappedOrganisation
		}
	}

	return organisation, nil
}
