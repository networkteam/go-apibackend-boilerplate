package domain

import (
	"github.com/friendsofgo/errors"
	"github.com/gofrs/uuid"
)

type OrganisationCreateCmd struct {
	OrganisationID uuid.UUID
	Name           string
}

func NewOrganisationCreateCmd() (cmd OrganisationCreateCmd, err error) {
	organisationID, err := uuid.NewV4()
	if err != nil {
		return cmd, errors.Wrap(err, "generating id")
	}

	return OrganisationCreateCmd{
		OrganisationID: organisationID,
	}, nil
}

func (c OrganisationCreateCmd) Validate() error {
	if IsBlank(c.Name) {
		return FieldError{
			Field: "name",
			Code:  ErrorCodeRequired,
		}
	}

	return nil
}
