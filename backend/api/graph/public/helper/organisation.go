package helper

import (
	"context"

	"myvendor.mytld/myproject/backend/domain/query"
)

func OrganisationQueryOptsFromSelection(ctx context.Context, organisationSelectPath ...string) *query.OrganisationQueryOpts {
	return &query.OrganisationQueryOpts{}
}
