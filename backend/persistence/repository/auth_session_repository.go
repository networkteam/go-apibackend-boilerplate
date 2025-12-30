package repository

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/networkteam/construct/v2/constructsql"
	. "github.com/networkteam/qrb"
	"github.com/networkteam/qrb/builder"
	"github.com/networkteam/qrb/qrbsql"

	"myvendor.mytld/myproject/backend/domain/model"
	domain_query "myvendor.mytld/myproject/backend/domain/query"
)

type AuthSessionsFilter struct {
	Opts       *domain_query.AuthSessionQueryOpts
	AccountIDs []uuid.UUID
}

func authSessionBuildFindQuery(opts *domain_query.AuthSessionQueryOpts) builder.SelectBuilder {
	return Select(buildAuthSessionJSON(opts)).
		From(authSession).
		SelectBuilder
}

func FindAuthSessionByID(ctx context.Context, executor qrbsql.Executor, id uuid.UUID, opts *domain_query.AuthSessionQueryOpts) (model.AuthSession, error) {
	query := authSessionBuildFindQuery(opts).
		Where(authSession.ID.Eq(Arg(id)))

	return constructsql.ScanRow[model.AuthSession](
		qrbsql.Build(query).WithExecutor(executor).QueryRow(ctx),
	)
}

func FindAllAuthSessions(ctx context.Context, executor qrbsql.Executor, filter AuthSessionsFilter, pagingOpts ...PagingOption) ([]model.AuthSession, error) {
	query := authSessionBuildFindQuery(filter.Opts).
		ApplyIf(true, applyAuthSessionFilter(filter))

	query, err := applyPagingOptions(query, pagingOpts, authSessionSortFields)
	if err != nil {
		return nil, err
	}

	return constructsql.CollectRows[model.AuthSession](
		qrbsql.Build(query).WithExecutor(executor).Query(ctx),
	)
}

func applyAuthSessionFilter(filter AuthSessionsFilter) func(q builder.SelectBuilder) builder.SelectBuilder {
	return func(q builder.SelectBuilder) builder.SelectBuilder {
		return q.
			ApplyIf(filter.AccountIDs != nil, func(q builder.SelectBuilder) builder.SelectBuilder {
				return q.Where(authSession.AccountID.Eq(Any(Arg(filter.AccountIDs))))
			})
	}
}

func InsertAuthSession(ctx context.Context, executor qrbsql.Executor, changeSet AuthSessionChangeSet) error {
	query := InsertInto(authSession).
		SetMap(changeSet.toMap())

	_, err := qrbsql.Build(query).WithExecutor(executor).Exec(ctx)
	return err
}

func DeleteAuthSession(ctx context.Context, executor qrbsql.Executor, id uuid.UUID) error {
	query := DeleteFrom(authSession).
		Where(authSession.ID.Eq(Arg(id)))

	return constructsql.AssertRowsAffected("delete", 1)(
		qrbsql.Build(query).WithExecutor(executor).Exec(ctx),
	)
}

func DeleteAllAuthSessionsByAccountID(ctx context.Context, executor qrbsql.Executor, accountID uuid.UUID) error {
	query := DeleteFrom(authSession).
		Where(authSession.AccountID.Eq(Arg(accountID)))

	_, err := qrbsql.Build(query).WithExecutor(executor).Exec(ctx)
	return err
}

func DeleteAllAuthSessionsExceptCurrent(ctx context.Context, executor qrbsql.Executor, accountID, currentSessionID uuid.UUID) error {
	query := DeleteFrom(authSession).
		Where(authSession.AccountID.Eq(Arg(accountID))).
		Where(authSession.ID.Neq(Arg(currentSessionID)))

	_, err := qrbsql.Build(query).WithExecutor(executor).Exec(ctx)
	return err
}

func DeleteAuthSessionsWithoutAccessToken(ctx context.Context, executor qrbsql.Executor) (int, error) {
	query := DeleteFrom(authSession).
		Where(
			Not(Exists(
				Select().
					From(authAccessToken).
					Where(authAccessToken.AuthSessionID.Eq(authSession.ID)),
			)),
		)

	res, err := qrbsql.Build(query).WithExecutor(executor).Exec(ctx)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := res.RowsAffected()
	return int(rowsAffected), err
}

func AuthSessionConstraintErr(_ error) error {
	return nil
}

func buildAuthSessionJSON(opts *domain_query.AuthSessionQueryOpts) builder.JsonBuildObjectBuilder {
	if opts == nil {
		opts = &domain_query.AuthSessionQueryOpts{}
	}

	return authSessionDefaultJson.
		ApplyIf(opts.IncludeAccount, func(b builder.JsonBuildObjectBuilder) builder.JsonBuildObjectBuilder {
			return b.Prop(
				"Account",
				Select(buildAccountJSON(opts.AccountQueryOpts)).
					From(account).
					Where(account.ID.Eq(authSession.AccountID)),
			)
		})
}
