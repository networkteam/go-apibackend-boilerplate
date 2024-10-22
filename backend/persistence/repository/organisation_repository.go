package repository

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/networkteam/construct/v2/constructsql"
	. "github.com/networkteam/qrb"
	"github.com/networkteam/qrb/builder"
	"github.com/networkteam/qrb/fn"
	"github.com/networkteam/qrb/qrbsql"

	"myvendor.mytld/myproject/backend/domain/model"
	domain_query "myvendor.mytld/myproject/backend/domain/query"
)

type OrganisationsFilter struct {
	Opts       *domain_query.OrganisationQueryOpts
	IDs        []uuid.UUID
	SearchTerm string
}

func organisationBuildFindQuery(opts *domain_query.OrganisationQueryOpts) builder.SelectBuilder {
	return Select(buildOrganisationJSON(opts)).
		From(organisation).
		SelectBuilder
}

func FindOrganisationByID(ctx context.Context, executor qrbsql.Executor, id uuid.UUID, opts *domain_query.OrganisationQueryOpts) (model.Organisation, error) {
	query := organisationBuildFindQuery(opts).
		Where(organisation.ID.Eq(Arg(id)))

	return constructsql.ScanRow[model.Organisation](
		qrbsql.Build(query).WithExecutor(executor).QueryRow(ctx),
	)
}

func applyOrganisationFilter(filter OrganisationsFilter) func(q builder.SelectBuilder) builder.SelectBuilder {
	return func(q builder.SelectBuilder) builder.SelectBuilder {
		return q.
			ApplyIf(len(filter.IDs) > 0, func(q builder.SelectBuilder) builder.SelectBuilder {
				return q.Where(organisation.ID.Eq(Any(Arg(filter.IDs))))
			}).
			ApplyIf(filter.SearchTerm != "", func(q builder.SelectBuilder) builder.SelectBuilder {
				return q.Where(organisation.Name.ILike(Arg("%" + filter.SearchTerm + "%")))
			})
	}
}

func FindAllOrganisations(ctx context.Context, executor qrbsql.Executor, filter OrganisationsFilter, pagingOpts ...PagingOption) ([]model.Organisation, error) {
	query := organisationBuildFindQuery(filter.Opts).
		ApplyIf(true, applyOrganisationFilter(filter))

	query, err := applyPagingOptions(query, pagingOpts, organisationSortFields)
	if err != nil {
		return nil, err
	}

	return constructsql.CollectRows[model.Organisation](
		qrbsql.Build(query).WithExecutor(executor).Query(ctx),
	)
}

func CountOrganisations(ctx context.Context, executor qrbsql.Executor, filter OrganisationsFilter) (count int, err error) {
	query := Select(fn.Count(N("*"))).
		From(organisation).
		ApplyIf(true, applyOrganisationFilter(filter))

	return constructsql.ScanRow[int](
		qrbsql.Build(query).WithExecutor(executor).QueryRow(ctx),
	)
}

func InsertOrganisation(ctx context.Context, executor qrbsql.Executor, changeSet OrganisationChangeSet) error {
	q := InsertInto(organisation).
		SetMap(changeSet.toMap())

	_, err := qrbsql.Build(q).WithExecutor(executor).Exec(ctx)
	return err
}

func UpdateOrganisation(ctx context.Context, executor qrbsql.Executor, id uuid.UUID, changeSet OrganisationChangeSet) error {
	q := Update(organisation).
		Where(organisation.ID.Eq(Arg(id))).
		SetMap(changeSet.toMap())

	return constructsql.AssertRowsAffected("update", 1)(
		qrbsql.Build(q).WithExecutor(executor).Exec(ctx),
	)
}

func DeleteOrganisation(ctx context.Context, executor qrbsql.Executor, id uuid.UUID) error {
	query := DeleteFrom(organisation).
		Where(organisation.ID.Eq(Arg(id)))

	return constructsql.AssertRowsAffected("delete", 1)(
		qrbsql.Build(query).WithExecutor(executor).Exec(ctx),
	)
}

func OrganisationConstraintErr(error) error {
	return nil
}

func buildOrganisationJSON(*domain_query.OrganisationQueryOpts) builder.JsonBuildObjectBuilder {
	return organisationDefaultJson
}
