package helper

import (
	"myvendor.mytld/myproject/backend/api/graph/model"
	"myvendor.mytld/myproject/backend/domain"
)

func MapToAccount(record domain.Account) *model.Account {
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

func MapToAccounts(records []domain.Account) []*model.Account {
	result := make([]*model.Account, len(records))
	for i, record := range records {
		result[i] = MapToAccount(record)
	}
	return result
}

func MapToAccountsQuery(filter *model.AccountFilter) domain.AccountsQuery {
	if filter == nil {
		return domain.AccountsQuery{}
	}
	return domain.AccountsQuery{
		IDs:            filter.Ids,
		Q:              filter.Q,
		OrganisationID: filter.OrganisationID,
	}
}
