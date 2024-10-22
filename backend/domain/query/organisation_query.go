package query

import "github.com/gofrs/uuid"

type OrganisationQuery struct {
	Opts           *OrganisationQueryOpts
	OrganisationID uuid.UUID
}

type OrganisationQueryOpts struct {
}

type OrganisationsQuery struct {
	Opts       *OrganisationQueryOpts
	IDs        []uuid.UUID
	SearchTerm string
}

func (f *OrganisationsQuery) SetOrganisationID(organisationID *uuid.UUID) {
	if organisationID != nil {
		f.IDs = []uuid.UUID{*organisationID}
	} else {
		f.IDs = nil
	}
}
