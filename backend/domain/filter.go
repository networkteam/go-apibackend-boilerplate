package domain

import "github.com/gofrs/uuid"

type OrganisationIDSetter interface {
	SetOrganisationID(organisationID *uuid.UUID)
}
