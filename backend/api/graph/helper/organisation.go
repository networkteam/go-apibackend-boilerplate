package helper

import (
	"context"

	"myvendor.mytld/myproject/backend/api/graph/model"
	"myvendor.mytld/myproject/backend/domain"
)

func MapToOrganisation(record domain.Organisation) *model.Organisation {
	return &model.Organisation{
		ID:        record.ID,
		Name:      record.Name,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}
}

func MapToOrganisations(records []domain.Organisation) []*model.Organisation {
	result := make([]*model.Organisation, len(records))
	for i, record := range records {
		result[i] = MapToOrganisation(record)
	}
	return result
}

func MapToOrganisationsQuery(filter *model.OrganisationFilter) domain.OrganisationsQuery {
	if filter == nil {
		return domain.OrganisationsQuery{}
	}
	return domain.OrganisationsQuery{
		IDs:        filter.Ids,
		SearchTerm: ToVal(filter.Q),
	}
}

func OrganisationQueryOptsFromSelection(ctx context.Context, organisationSelectPath ...string) domain.OrganisationQueryOpts {
	return domain.OrganisationQueryOpts{}
}
