package finder

import (
	"context"

	"myvendor.mytld/myproject/backend/domain/model"
	domain_query "myvendor.mytld/myproject/backend/domain/query"
	"myvendor.mytld/myproject/backend/persistence/repository"
	"myvendor.mytld/myproject/backend/security/authentication"
	"myvendor.mytld/myproject/backend/security/authorization"
)

func (f *Finder) QueryOrganisation(ctx context.Context, query domain_query.OrganisationQuery) (model.Organisation, error) {
	authorizer := authorization.NewAuthorizer(authentication.GetAuthContext(ctx))
	err := authorizer.AllowsOrganisationQuery(query)
	if err != nil {
		return model.Organisation{}, err
	}

	record, err := repository.FindOrganisationByID(ctx, f.executor, query.OrganisationID, query.Opts)
	if err != nil {
		return record, err
	}
	return record, nil
}

func (f *Finder) QueryOrganisations(ctx context.Context, query domain_query.OrganisationsQuery, paging Paging) ([]model.Organisation, error) {
	authorizer := authorization.NewAuthorizer(authentication.GetAuthContext(ctx))
	err := authorizer.AllowsAndFilterAllOrganisationsQuery(&query)
	if err != nil {
		return nil, err
	}
	return repository.FindAllOrganisations(ctx, f.executor, repository.OrganisationsFilter{
		Opts:       query.Opts,
		IDs:        query.IDs,
		SearchTerm: query.SearchTerm,
	}, paging.options()...)
}

func (f *Finder) CountOrganisations(ctx context.Context, query domain_query.OrganisationsQuery) (int, error) {
	authorizer := authorization.NewAuthorizer(authentication.GetAuthContext(ctx))
	err := authorizer.AllowsAndFilterAllOrganisationsQuery(&query)
	if err != nil {
		return 0, err
	}

	return repository.CountOrganisations(ctx, f.executor, repository.OrganisationsFilter{
		IDs:        query.IDs,
		SearchTerm: query.SearchTerm,
	})
}
