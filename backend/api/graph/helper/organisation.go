package helper

import (
	"context"

	"myvendor.mytld/myproject/backend/api/graph/model"
	model2 "myvendor.mytld/myproject/backend/domain/model"
	"myvendor.mytld/myproject/backend/domain/query"
)

func MapToOrganisation(record model2.Organisation) *model.Organisation {
	return &model.Organisation{
		ID:        record.ID,
		Name:      record.Name,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}
}

func MapToOrganisations(records []model2.Organisation) []*model.Organisation {
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
		SearchTerm: ToVal(filter.Q),
	}
}

func OrganisationQueryOptsFromSelection(ctx context.Context, organisationSelectPath ...string) *query.OrganisationQueryOpts {
	return &query.OrganisationQueryOpts{}
}
