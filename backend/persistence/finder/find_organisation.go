package finder

import (
	"context"

	"github.com/friendsofgo/errors"
	"github.com/zbyte/go-kallax"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/persistence/records"
	"myvendor.mytld/myproject/backend/security/authentication"
	"myvendor.mytld/myproject/backend/security/authorization"
)

func (f *Finder) FindOrganisationQuery(ctx context.Context, query domain.OrganisationQuery) (result *records.Organisation, err error) {
	if err = authorization.NewAuthorizer(authentication.GetAuthContext(ctx)).AllowsOrganisationQuery(query); err != nil {
		return
	}

	q := records.NewOrganisationQuery().
		FindByID(kallax.UUID(query.OrganisationID))

	result, err = f.organisationStore.FindOne(q)
	if err == kallax.ErrNotFound {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "fetching organisation")
	}
	return
}
