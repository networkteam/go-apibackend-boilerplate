package helper

import (
	"context"

	"myvendor.mytld/myproject/backend/api/graph/model"
	model2 "myvendor.mytld/myproject/backend/domain/model"
	"myvendor.mytld/myproject/backend/domain/query"
)

func MapToAccount(record model2.Account) *model.Account {
	return &model.Account{
		ID:             record.ID,
		EmailAddress:   record.EmailAddress,
		Role:           record.Role,
		LastLogin:      record.LastLogin,
		OrganisationID: uuidOrNil(record.OrganisationID),
		CreatedAt:      record.CreatedAt,
		UpdatedAt:      record.UpdatedAt,
	}
}

func MapToAccounts(records []model2.Account) []*model.Account {
	result := make([]*model.Account, len(records))
	for i, record := range records {
		result[i] = MapToAccount(record)
	}
	return result
}

func MapFromAccountFilter(filter *model.AccountFilter) query.AccountsQuery {
	if filter == nil {
		return query.AccountsQuery{}
	}
	return query.AccountsQuery{
		IDs:            filter.Ids,
		SearchTerm:     ToVal(filter.Q),
		OrganisationID: filter.OrganisationID,
	}
}

func AccountQueryOptsFromSelection(ctx context.Context, accountSelectPath ...string) *query.AccountQueryOpts {
	selectedFields := SelectedFields(ctx)
	return &query.AccountQueryOpts{
		IncludeOrganisation:   selectedFields.PathSelected(append(accountSelectPath, "organisation")...),
		OrganisationQueryOpts: OrganisationQueryOptsFromSelection(ctx, append(accountSelectPath, "organisation")...),
	}
}
