package domain

import "github.com/gofrs/uuid"

type AccountQuery struct {
	AccountID uuid.UUID
}

type AccountsQuery struct {
	IDs            []uuid.UUID
	Q              *string
	OrganisationID *uuid.UUID
	ExcludeRole    *Role
}

func (f *AccountsQuery) SetOrganisationID(organisationID *uuid.UUID) {
	f.OrganisationID = organisationID
}
