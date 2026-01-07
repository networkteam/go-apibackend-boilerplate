package helper

import (
	"myvendor.mytld/myproject/backend/api/graph/admin/model"
	common_helper "myvendor.mytld/myproject/backend/api/graph/common/helper"
	domain_model "myvendor.mytld/myproject/backend/domain/model"
	"myvendor.mytld/myproject/backend/domain/query"
)

func MapToOrganisation(record domain_model.Organisation) *model.Organisation {
	return &model.Organisation{
		ID:        record.ID,
		Name:      record.Name,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}
}

func MapToOrganisations(records []domain_model.Organisation) []*model.Organisation {
	result := make([]*model.Organisation, len(records))
	for i, record := range records {
		result[i] = MapToOrganisation(record)
	}
	return result
}

func MapToOrganisationsQuery(filter *model.OrganisationFilter) query.OrganisationsQuery {
	if filter == nil {
		return query.OrganisationsQuery{}
	}
	return query.OrganisationsQuery{
		IDs:        filter.Ids,
		SearchTerm: common_helper.ToVal(filter.Q),
	}
}
