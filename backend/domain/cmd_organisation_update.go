package domain

import (
	"github.com/gofrs/uuid"
)

type OrganisationUpdateCmd struct {
	OrganisationID uuid.UUID
	Name           string
}

func (c OrganisationUpdateCmd) Validate() error {
	if IsBlank(c.Name) {
		return FieldError{
			Field: "name",
			Code:  ErrorCodeRequired,
		}
	}

	return nil
}
