package finder

import (
	"context"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/persistence/repository"
	"myvendor.mytld/myproject/backend/security/authentication"
	"myvendor.mytld/myproject/backend/security/authorization"
)

func (f *Finder) QueryAccount(ctx context.Context, query domain.AccountQuery) (domain.Account, error) {
	record, err := repository.FindAccountByID(ctx, f.db, query.AccountID)
	if err != nil {
		return record, err
	}
	err = authorization.NewAuthorizer(authentication.GetAuthContext(ctx)).AllowsAccountView(record)
	if err != nil {
		return record, err
	}
	return record, nil
}

func (f *Finder) QueryAccounts(ctx context.Context, query domain.AccountsQuery, paging repository.Paging) ([]domain.Account, error) {
	authorizer := authorization.NewAuthorizer(authentication.GetAuthContext(ctx))
	err := authorizer.AllowsAndFilterAllAccountsQuery(&query)
	if err != nil {
		return nil, err
	}

	return repository.FindAllAccounts(ctx, f.db, paging, query)
}

func (f *Finder) CountAccounts(ctx context.Context, query domain.AccountsQuery) (int, error) {
	authorizer := authorization.NewAuthorizer(authentication.GetAuthContext(ctx))
	err := authorizer.AllowsAndFilterAllAccountsQuery(&query)
	if err != nil {
		return 0, err
	}

	return repository.CountAllAccounts(ctx, f.db, query)
}
