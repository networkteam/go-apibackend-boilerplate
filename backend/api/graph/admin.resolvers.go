package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/gofrs/uuid"
	"myvendor.mytld/myproject/backend/api/graph/generated"
	"myvendor.mytld/myproject/backend/api/graph/helper"
	"myvendor.mytld/myproject/backend/api/graph/model"
	domain_model "myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/persistence/repository"
)

func (r *mutationResolver) CreateAccount(ctx context.Context, role model.Role, emailAddress string, password string, organisationID *uuid.UUID) (*model.Account, error) {
	cmd, err := domain_model.NewAccountCreateCmd(emailAddress, domain_model.Role(role), password)
	if err != nil {
		return nil, err
	}

	// Only set OrganisationID if the role fits (work around an issue with selecting an organisation and then changing the role in the admin UI)
	if domain_model.Role(role) != domain_model.RoleSystemAdministrator {
		cmd.OrganisationID = helper.NullUUIDVal(organisationID)
	}
	err = r.Handler().AccountCreate(ctx, cmd)
	if err != nil {
		return nil, err
	}

	record, err := r.Finder().QueryAccount(ctx, domain_model.AccountQuery{
		AccountID: cmd.AccountID,
	})
	if err != nil {
		return nil, err
	}
	return helper.MapToAccount(record), nil
}

func (r *mutationResolver) UpdateAccount(ctx context.Context, id uuid.UUID, role model.Role, emailAddress string, password *string, organisationID *uuid.UUID) (*model.Account, error) {
	// Fetch previous record to get organisation id
	prevRecord, err := r.Finder().QueryAccount(ctx, domain_model.AccountQuery{
		AccountID: id,
	})
	if err != nil {
		return nil, err
	}

	cmd, err := domain_model.NewAccountUpdateCmd(r.Config, prevRecord.OrganisationID, id, emailAddress, domain_model.Role(role), helper.StrVal(password))
	if err != nil {
		return nil, err
	}
	// Only set NewOrganisationID if the role fits (work around an issue with selecting an organisation and then changing the role in the admin UI)
	if domain_model.Role(role) != domain_model.RoleSystemAdministrator {
		cmd.NewOrganisationID = helper.NullUUIDVal(organisationID)
	}
	err = r.Handler().AccountUpdate(ctx, cmd)
	if err != nil {
		return nil, err
	}

	record, err := r.Finder().QueryAccount(ctx, domain_model.AccountQuery{
		AccountID: id,
	})
	if err != nil {
		return nil, err
	}
	return helper.MapToAccount(record), nil
}

func (r *mutationResolver) DeleteAccount(ctx context.Context, id uuid.UUID) (*model.Account, error) {
	record, err := r.Finder().QueryAccount(ctx, domain_model.AccountQuery{
		AccountID: id,
	})
	if err != nil {
		return nil, err
	}

	cmd := domain_model.NewAccountDeleteCmd(id, record.OrganisationID)
	err = r.Handler().AccountDelete(ctx, cmd)
	if err != nil {
		return nil, err
	}
	return helper.MapToAccount(record), nil
}

func (r *mutationResolver) CreateOrganisation(ctx context.Context, name string) (*model.Organisation, error) {
	cmd, err := domain_model.NewOrganisationCreateCmd()
	if err != nil {
		return nil, err
	}
	cmd.Name = name
	err = r.Handler().OrganisationCreate(ctx, cmd)
	if err != nil {
		return nil, err
	}
	record, err := r.Finder().QueryOrganisation(ctx, domain_model.OrganisationQuery{
		OrganisationID: cmd.OrganisationID,
	})
	if err != nil {
		return nil, err
	}
	return helper.MapToOrganisation(record), nil
}

func (r *mutationResolver) UpdateOrganisation(ctx context.Context, id uuid.UUID, name string) (*model.Organisation, error) {
	cmd := domain_model.OrganisationUpdateCmd{
		OrganisationID: id,
		Name:           name,
	}
	err := r.Handler().OrganisationUpdate(ctx, cmd)
	if err != nil {
		return nil, err
	}
	record, err := r.Finder().QueryOrganisation(ctx, domain_model.OrganisationQuery{
		OrganisationID: id,
	})
	if err != nil {
		return nil, err
	}
	return helper.MapToOrganisation(record), nil
}

func (r *mutationResolver) DeleteOrganisation(ctx context.Context, id uuid.UUID) (*model.Organisation, error) {
	record, err := r.Finder().QueryOrganisation(ctx, domain_model.OrganisationQuery{
		OrganisationID: id,
	})
	if err != nil {
		return nil, err
	}

	cmd := domain_model.NewOrganisationDeleteCmd(id)
	err = r.Handler().OrganisationDelete(ctx, cmd)
	if err != nil {
		return nil, err
	}
	return helper.MapToOrganisation(record), nil
}

func (r *queryResolver) Account(ctx context.Context, id uuid.UUID) (*model.Account, error) {
	record, err := r.Finder().QueryAccount(ctx, domain_model.AccountQuery{
		AccountID: id,
	})
	if err == repository.ErrNotFound {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return helper.MapToAccount(record), nil
}

func (r *queryResolver) AllAccounts(ctx context.Context, page *int, perPage *int, sortField *string, sortOrder *string, filter *model.AccountFilter) ([]*model.Account, error) {
	query := helper.MapToAccountsQuery(filter)

	records, err := r.Finder().QueryAccounts(ctx, query, helper.MapToPaging(page, perPage, sortField, sortOrder))
	if err != nil {
		return nil, err
	}
	return helper.MapToAccounts(records), nil
}

func (r *queryResolver) AllAccountsMeta(ctx context.Context, page *int, perPage *int, sortField *string, sortOrder *string, filter *model.AccountFilter) (*model.ListMetadata, error) {
	query := helper.MapToAccountsQuery(filter)
	count, err := r.Finder().CountAccounts(ctx, query)
	if err != nil {
		return nil, err
	}
	return &model.ListMetadata{
		Count: count,
	}, nil
}

func (r *queryResolver) Organisation(ctx context.Context, id uuid.UUID) (*model.Organisation, error) {
	record, err := r.Finder().QueryOrganisation(ctx, domain_model.OrganisationQuery{
		OrganisationID: id,
	})
	if err == repository.ErrNotFound {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return helper.MapToOrganisation(record), nil
}

func (r *queryResolver) AllOrganisations(ctx context.Context, page *int, perPage *int, sortField *string, sortOrder *string, filter *model.OrganisationFilter) ([]*model.Organisation, error) {
	query := helper.MapToOrganisationsQuery(filter)

	records, err := r.Finder().QueryOrganisations(ctx, query, helper.MapToPaging(page, perPage, sortField, sortOrder))
	if err != nil {
		return nil, err
	}
	return helper.MapToOrganisations(records), nil
}

func (r *queryResolver) AllOrganisationsMeta(ctx context.Context, page *int, perPage *int, sortField *string, sortOrder *string, filter *model.OrganisationFilter) (*model.ListMetadata, error) {
	query := helper.MapToOrganisationsQuery(filter)

	count, err := r.Finder().CountOrganisations(ctx, query)
	if err != nil {
		return nil, err
	}
	return &model.ListMetadata{
		Count: count,
	}, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

type mutationResolver struct{ *Resolver }
