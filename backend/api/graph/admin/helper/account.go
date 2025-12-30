package helper

import (
	"myvendor.mytld/myproject/backend/api/graph/admin/model"
	common_helper "myvendor.mytld/myproject/backend/api/graph/common/helper"
	domain_model "myvendor.mytld/myproject/backend/domain/model"
	"myvendor.mytld/myproject/backend/domain/query"
)

func MapToAccount(record domain_model.Account) *model.Account {
	return &model.Account{
		ID:             record.ID,
		EmailAddress:   record.EmailAddress,
		Role:           record.Role,
		LastLogin:      record.LastLogin,
		OrganisationID: common_helper.ToUUID(record.OrganisationID),
		CreatedAt:      record.CreatedAt,
		UpdatedAt:      record.UpdatedAt,
	}
}

func MapToAccounts(records []domain_model.Account) []*model.Account {
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
		SearchTerm:     common_helper.ToVal(filter.Q),
		OrganisationID: filter.OrganisationID,
	}
}
