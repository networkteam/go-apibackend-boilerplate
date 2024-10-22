package query

import (
	"github.com/gofrs/uuid"
)

type AccountQuery struct {
	Opts      *AccountQueryOpts
	AccountID uuid.UUID
}

type AccountQueryNotAuthorized struct {
	Opts              *AccountQueryOpts
	AccountID         *uuid.UUID
	EmailAddress      *string
	ConfirmationToken *string
}

type AccountsQuery struct {
	Opts           *AccountQueryOpts
	IDs            []uuid.UUID
	SearchTerm     string
	OrganisationID *uuid.UUID
}

func (f *AccountsQuery) SetOrganisationID(organisationID *uuid.UUID) {
	f.OrganisationID = organisationID
}

type AccountQueryOpts struct {
	IncludeOrganisation   bool
	OrganisationQueryOpts *OrganisationQueryOpts
}
