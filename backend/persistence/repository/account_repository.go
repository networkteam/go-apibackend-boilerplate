package repository

import (
	"context"
	"strings"

	"github.com/friendsofgo/errors"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/networkteam/construct/v2/constructsql"
	. "github.com/networkteam/qrb"
	"github.com/networkteam/qrb/builder"
	"github.com/networkteam/qrb/fn"
	"github.com/networkteam/qrb/qrbsql"

	"myvendor.mytld/myproject/backend/domain/model"
	domain_query "myvendor.mytld/myproject/backend/domain/query"
	"myvendor.mytld/myproject/backend/domain/types"
)

type AccountsFilter struct {
	Opts           *domain_query.AccountQueryOpts
	OrganisationID *uuid.UUID
	IDs            []uuid.UUID
	// SearchTerm filters accounts by text fields (email address or organisation name)
	SearchTerm string
	// Roles filters account to have one of the given roles
	Roles []types.Role
}

func accountBuildFindQuery(opts *domain_query.AccountQueryOpts) builder.SelectBuilder {
	if opts == nil {
		opts = &domain_query.AccountQueryOpts{}
	}

	return Select(buildAccountJSON(opts)).
		From(account).
		ApplyIf(opts.IncludeOrganisation, func(q builder.SelectBuilder) builder.SelectBuilder {
			return q.LeftJoin(organisation).On(organisation.ID.Eq(account.OrganisationID))
		})
}

func FindAccountByID(ctx context.Context, executor qrbsql.Executor, id uuid.UUID, opts *domain_query.AccountQueryOpts) (model.Account, error) {
	query := accountBuildFindQuery(opts).
		Where(account.ID.Eq(Arg(id)))

	return constructsql.ScanRow[model.Account](
		qrbsql.Build(query).WithExecutor(executor).QueryRow(ctx),
	)
}

func FindAccountByEmailAddress(ctx context.Context, executor qrbsql.Executor, emailAddress string, opts *domain_query.AccountQueryOpts) (model.Account, error) {
	query := accountBuildFindQuery(opts).
		Where(fn.Lower(account.EmailAddress).Eq(Arg(strings.ToLower(emailAddress))))

	return constructsql.ScanRow[model.Account](
		qrbsql.Build(query).WithExecutor(executor).QueryRow(ctx),
	)
}

func applyAccountFilter(filter AccountsFilter) func(q builder.SelectBuilder) builder.SelectBuilder {
	return func(q builder.SelectBuilder) builder.SelectBuilder {
		return q.
			ApplyIf(len(filter.IDs) > 0, func(q builder.SelectBuilder) builder.SelectBuilder {
				return q.Where(account.ID.Eq(Any(Arg(filter.IDs))))
			}).
			ApplyIf(filter.SearchTerm != "", func(q builder.SelectBuilder) builder.SelectBuilder {
				var incOrg builder.Exp
				if filter.Opts != nil && filter.Opts.IncludeOrganisation {
					incOrg = organisation.Name.ILike(Arg("%" + filter.SearchTerm + "%"))
				}

				return q.Where(Or(
					account.EmailAddress.ILike(Arg("%"+filter.SearchTerm+"%")),
					incOrg,
				))
			}).
			ApplyIf(filter.OrganisationID != nil, func(q builder.SelectBuilder) builder.SelectBuilder {
				return q.Where(account.OrganisationID.Eq(Arg(*filter.OrganisationID)))
			}).
			ApplyIf(len(filter.Roles) > 0, func(q builder.SelectBuilder) builder.SelectBuilder {
				return q.Where(account.Role.Eq(Any(Args(filter.Roles))))
			})
	}
}

func FindAllAccounts(ctx context.Context, executor qrbsql.Executor, filter AccountsFilter, pagingOpts ...PagingOption) ([]model.Account, error) {
	query := accountBuildFindQuery(filter.Opts).
		ApplyIf(true, applyAccountFilter(filter))

	query, err := applyPagingOptions(query, pagingOpts, accountSortFields)
	if err != nil {
		return nil, err
	}

	return constructsql.CollectRows[model.Account](
		qrbsql.Build(query).WithExecutor(executor).Query(ctx),
	)
}

func CountAccounts(ctx context.Context, executor qrbsql.Executor, filter AccountsFilter) (count int, err error) {
	query := Select(fn.Count(N("*"))).
		From(account).
		ApplyIf(true, applyAccountFilter(filter))

	return constructsql.ScanRow[int](
		qrbsql.Build(query).WithExecutor(executor).QueryRow(ctx),
	)
}

func InsertAccount(ctx context.Context, executor qrbsql.Executor, changeSet AccountChangeSet) error {
	query := InsertInto(account).
		SetMap(changeSet.toMap())

	_, err := qrbsql.Build(query).WithExecutor(executor).Exec(ctx)
	return err
}

func UpdateAccount(ctx context.Context, executor qrbsql.Executor, id uuid.UUID, changeSet AccountChangeSet) error {
	query := Update(account).
		SetMap(changeSet.toMap()).
		Where(account.ID.Eq(Arg(id)))

	return constructsql.AssertRowsAffected("update", 1)(
		qrbsql.Build(query).WithExecutor(executor).Exec(ctx),
	)
}

func DeleteAccount(ctx context.Context, executor qrbsql.Executor, id uuid.UUID) error {
	query := DeleteFrom(account).
		Where(account.ID.Eq(Arg(id)))

	return constructsql.AssertRowsAffected("delete", 1)(
		qrbsql.Build(query).WithExecutor(executor).Exec(ctx),
	)
}

func AccountConstraintErr(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgErrCode_unique_violation && pgErr.ConstraintName == "accounts_email_address_idx" {
			return types.FieldError{
				Field: "emailAddress",
				Code:  types.ErrorCodeAlreadyExists,
			}
		}
	}
	return nil
}

func buildAccountJSON(opts *domain_query.AccountQueryOpts) builder.JsonBuildObjectBuilder {
	if opts == nil {
		opts = &domain_query.AccountQueryOpts{}
	}

	return accountDefaultJson.
		PropIf(
			opts.IncludeOrganisation,
			"Organisation",
			Select(buildOrganisationJSON(opts.OrganisationQueryOpts)).
				From(organisation).
				Where(organisation.ID.Eq(account.OrganisationID)))
}
