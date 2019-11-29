package authentication

import (
	"context"

	"github.com/friendsofgo/errors"

	"myvendor.mytld/myproject/backend/api"
	"myvendor.mytld/myproject/backend/api/helper"
	"myvendor.mytld/myproject/backend/domain"
)

type AppAccountResolver struct{ api.ResolverDependencies }

var _ api.AppAccountResolver = &AppAccountResolver{}

func (a *AppAccountResolver) Organisation(ctx context.Context, appAccount *api.AppAccount) (*api.Organisation, error) {
	f := a.Finder()

	organisation, err := f.FindOrganisationQuery(ctx, domain.OrganisationQuery{
		OrganisationID: appAccount.OrganisationID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to query organisation")
	}
	return helper.MapToOrganisation(organisation)
}
