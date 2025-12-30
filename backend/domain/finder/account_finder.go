package finder

import (
	"context"

	"github.com/friendsofgo/errors"

	"myvendor.mytld/myproject/backend/domain/model"
	domain_query "myvendor.mytld/myproject/backend/domain/query"
	"myvendor.mytld/myproject/backend/persistence/repository"
	"myvendor.mytld/myproject/backend/security/authentication"
	"myvendor.mytld/myproject/backend/security/authorization"
)

func (f *Finder) QueryAccount(ctx context.Context, query domain_query.AccountQuery) (model.Account, error) {
	record, err := repository.FindAccountByID(ctx, f.executor, query.AccountID, query.Opts)
	if err != nil {
		return record, err
	}
	err = authorization.NewAuthorizer(authentication.GetAuthContext(ctx)).AllowsAccountView(record)
	if err != nil {
		return record, err
	}
	return record, nil
}

func (f *Finder) QueryAccountNotAuthorized(ctx context.Context, query domain_query.AccountQueryNotAuthorized) (model.Account, error) {
	if query.AccountID != nil {
		return repository.FindAccountByID(ctx, f.executor, *query.AccountID, query.Opts)
	}

	if query.EmailAddress != nil {
		return repository.FindAccountByEmailAddress(ctx, f.executor, *query.EmailAddress, query.Opts)
	}

	return model.Account{}, errors.Wrap(ErrInvalidQuery, "AccountID or EmailAddress must be set")
}

func (f *Finder) QueryAccounts(ctx context.Context, query domain_query.AccountsQuery, paging Paging) ([]model.Account, error) {
	authorizer := authorization.NewAuthorizer(authentication.GetAuthContext(ctx))
	err := authorizer.AllowsAndFilterAllAccountsQuery(&query)
	if err != nil {
		return nil, err
	}

	return repository.FindAllAccounts(ctx, f.executor, repository.AccountsFilter{
		Opts:           query.Opts,
		OrganisationID: query.OrganisationID,
		IDs:            query.IDs,
		SearchTerm:     query.SearchTerm,
	}, paging.options()...)
}

func (f *Finder) CountAccounts(ctx context.Context, query domain_query.AccountsQuery) (int, error) {
	authorizer := authorization.NewAuthorizer(authentication.GetAuthContext(ctx))
	err := authorizer.AllowsAndFilterAllAccountsQuery(&query)
	if err != nil {
		return 0, err
	}

	return repository.CountAccounts(ctx, f.executor, repository.AccountsFilter{
		OrganisationID: query.OrganisationID,
		IDs:            query.IDs,
		SearchTerm:     query.SearchTerm,
	})
}
