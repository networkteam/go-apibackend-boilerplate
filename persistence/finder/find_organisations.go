package finder

import (
	"context"

	"github.com/pkg/errors"
	"github.com/zbyte/go-kallax"

	"myvendor/myproject/backend/domain"
	"myvendor/myproject/backend/persistence/records"
	"myvendor/myproject/backend/security/authentication"
	"myvendor/myproject/backend/security/authorization"
)

func (f *Finder) FindOrganisationsQuery(ctx context.Context, query domain.OrganisationsQuery) (result *records.OrganisationResultSet, err error) {
	if err = authorization.NewAuthorizer(authentication.GetAuthContext(ctx)).AllowsOrganisationsQuery(query); err != nil {
		return
	}

	q := records.NewOrganisationQuery().
		Order(kallax.Asc(records.Schema.Organisation.Name))

	result, err = f.organisationStore.Find(q)
	if err != nil {
		return nil, errors.Wrap(err, "fetching organisations")
	}
	return
}
