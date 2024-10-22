package command

import (
	"github.com/gofrs/uuid"

	"myvendor.mytld/myproject/backend/domain/types"
)

type OrganisationUpdateCmd struct {
	OrganisationID uuid.UUID
	Name           string
}

func (c OrganisationUpdateCmd) Validate() error {
	if isBlank(c.Name) {
		return types.FieldError{
			Field: "name",
			Code:  types.ErrorCodeRequired,
		}
	}

	return nil
}
