package repository

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/friendsofgo/errors"
	"github.com/gofrs/uuid"

	"myvendor.mytld/myproject/backend/domain"
)

func organisationBuildFindQuery(runner squirrel.BaseRunner) squirrel.SelectBuilder {
	return queryBuilder(runner).
		Select(buildOrganisationJSON()).
		From("organisations")
}

func FindOrganisationByID(ctx context.Context, runner squirrel.BaseRunner, id uuid.UUID) (domain.Organisation, error) {
	row := organisationBuildFindQuery(runner).
		Where(squirrel.Eq{organisation_id: id}).
		QueryRowContext(ctx)
	return organisationScanJsonRow(row)
}

func FindAllOrganisations(ctx context.Context, runner squirrel.BaseRunner, paging Paging, filter domain.OrganisationsQuery) (result []domain.Organisation, err error) {
	query := organisationBuildFindQuery(runner)

	query, err = applyPaging(query, paging, organisationSortFields)
	if err != nil {
		return
	}
	query = applyOrganisationFilter(query, filter)

	rows, err := query.QueryContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "executing query")
	}
	defer rows.Close()
	for rows.Next() {
		organisation, err := organisationScanJsonRow(rows)
		if err != nil {
			return nil, errors.Wrap(err, "scanning row")
		}
		result = append(result, organisation)
	}
	return
}

func applyOrganisationFilter(query squirrel.SelectBuilder, filter domain.OrganisationsQuery) squirrel.SelectBuilder {
	if len(filter.IDs) > 0 {
		query = query.Where(
			squirrel.Eq{organisation_id: filter.IDs},
		)
	}
	if filter.Q != nil {
		conds := squirrel.Or{
			squirrel.ILike{organisation_name: "%" + *filter.Q + "%"},
		}
		query = query.Where(conds)
	}
	return query
}

func CountAllOrganisations(ctx context.Context, runner squirrel.BaseRunner, filter domain.OrganisationsQuery) (count int, err error) {
	query := queryBuilder(runner).
		Select("COUNT(*)").
		From("organisations")

	query = applyOrganisationFilter(query, filter)

	row := query.QueryRowContext(ctx)
	err = row.Scan(&count)
	return
}

func InsertOrganisation(ctx context.Context, runner squirrel.BaseRunner, changeSet OrganisationChangeSet) error {
	_, err := queryBuilder(runner).
		Insert("organisations").
		SetMap(changeSet.toMap()).
		ExecContext(ctx)
	return err
}

func UpdateOrganisation(ctx context.Context, runner squirrel.BaseRunner, id uuid.UUID, changeSet OrganisationChangeSet) error {
	res, err := queryBuilder(runner).
		Update("organisations").
		Where(squirrel.Eq{organisation_id: id}).
		SetMap(changeSet.toMap()).
		ExecContext(ctx)
	if err != nil {
		return errors.Wrap(err, "executing update")
	}
	return assertRowsAffected(res, "update")
}

func DeleteOrganisation(ctx context.Context, runner squirrel.BaseRunner, id uuid.UUID) error {
	res, err := queryBuilder(runner).
		Delete("organisations").
		Where(squirrel.Eq{organisation_id: id}).
		ExecContext(ctx)
	if err != nil {
		return errors.Wrap(err, "executing delete")
	}
	return assertRowsAffected(res, "delete")
}

func OrganisationConstraintErr(err error) error {
	return nil
}

func buildOrganisationJSON() string {
	return organisationDefaultSelectJson.
		ToSql()
}
