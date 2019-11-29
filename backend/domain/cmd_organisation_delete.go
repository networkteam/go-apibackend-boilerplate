package domain

import (
	"github.com/gofrs/uuid"
)

type OrganisationDeleteCmd struct {
	OrganisationID uuid.UUID
}

func NewOrganisationDeleteCmd(organisationID uuid.UUID) OrganisationDeleteCmd {

	return OrganisationDeleteCmd{
		OrganisationID: organisationID,
	}
}
