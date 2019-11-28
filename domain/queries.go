package domain

import "github.com/gofrs/uuid"

type OrganisationsQuery struct{}

type OrganisationQuery struct {
	OrganisationID uuid.UUID
}
