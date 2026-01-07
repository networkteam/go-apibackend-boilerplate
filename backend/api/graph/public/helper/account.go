package helper

import (
	"context"

	common_helper "myvendor.mytld/myproject/backend/api/graph/common/helper"
	"myvendor.mytld/myproject/backend/api/graph/public/model"
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

func AccountQueryOptsFromSelection(ctx context.Context, accountSelectPath ...string) *query.AccountQueryOpts {
	selectedFields := common_helper.SelectedFields(ctx)
	return &query.AccountQueryOpts{
		IncludeOrganisation:   selectedFields.PathSelected(append(accountSelectPath, "organisation")...),
		OrganisationQueryOpts: OrganisationQueryOptsFromSelection(ctx, append(accountSelectPath, "organisation")...),
	}
}
