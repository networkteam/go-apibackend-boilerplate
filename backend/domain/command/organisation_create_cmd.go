package command

import (
	"github.com/friendsofgo/errors"
	"github.com/gofrs/uuid"

	"myvendor.mytld/myproject/backend/domain/types"
)

type OrganisationCreateCmd struct {
	OrganisationID uuid.UUID
	Name           string
}

func NewOrganisationCreateCmd() (cmd OrganisationCreateCmd, err error) {
	organisationID, err := uuid.NewV7()
	if err != nil {
		return cmd, errors.Wrap(err, "generating id")
	}

	return OrganisationCreateCmd{
		OrganisationID: organisationID,
	}, nil
}

func (c OrganisationCreateCmd) Validate() error {
	if isBlank(c.Name) {
		return types.FieldError{
			Field: "name",
			Code:  types.ErrorCodeRequired,
		}
	}

	return nil
}
