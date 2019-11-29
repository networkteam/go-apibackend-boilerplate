package domain

import (
	"github.com/gofrs/uuid"
)

type OrganisationUpdateCmd struct {
	OrganisationID uuid.UUID
	Name           string
}

func NewOrganisationUpdateCmd(organisationID uuid.UUID) (cmd OrganisationUpdateCmd, err error) {

	return OrganisationUpdateCmd{
		OrganisationID: organisationID,
	}, nil
}

func (c OrganisationUpdateCmd) Validate() error {

	if c.OrganisationID == uuid.Nil {
		return FieldError{
			Field: "organisationId",
			Code:  ErrorCodeRequired,
		}
	}

	if isBlank(c.Name) {
		return FieldError{
			Field: "name",
			Code:  ErrorCodeRequired,
		}
	}

	return nil
}
