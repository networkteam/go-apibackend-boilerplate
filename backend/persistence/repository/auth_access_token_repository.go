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
	"myvendor.mytld/myproject/backend/domain/types"
)

func authAccessTokenBuildFindQuery(opts domain_query.AuthAccessTokenQueryOpts) builder.SelectBuilder {
	return Select(buildAuthAccessTokenJSON(opts)).
		From(authAccessToken).
		SelectBuilder
}

func FindAuthAccessTokenByID(ctx context.Context, executor qrbsql.Executor, id uuid.UUID, opts domain_query.AuthAccessTokenQueryOpts) (model.AuthAccessToken, error) {
	query := authAccessTokenBuildFindQuery(opts).
		Where(authAccessToken.ID.Eq(Arg(id)))

	return constructsql.ScanRow[model.AuthAccessToken](
		qrbsql.Build(query).WithExecutor(executor).QueryRow(ctx),
	)
}

func FindAllAuthAccessTokens(ctx context.Context, executor qrbsql.Executor, opts domain_query.AuthAccessTokenQueryOpts) ([]model.AuthAccessToken, error) {
	query := authAccessTokenBuildFindQuery(opts)

	return constructsql.CollectRows[model.AuthAccessToken](
		qrbsql.Build(query).WithExecutor(executor).Query(ctx),
	)
}

func InsertAuthAccessToken(ctx context.Context, executor qrbsql.Executor, changeSet AuthAccessTokenChangeSet) error {
	query := InsertInto(authAccessToken).
		SetMap(changeSet.toMap())

	_, err := qrbsql.Build(query).WithExecutor(executor).Exec(ctx)
	return err
}

func DeleteExpiredAuthAccessTokens(ctx context.Context, executor qrbsql.Executor, timeSource types.TimeSource) (int, error) {
	query := DeleteFrom(authAccessToken).
		Where(authAccessToken.ExpiresAt.Lt(Arg(timeSource.Now())))

	res, err := qrbsql.Build(query).WithExecutor(executor).Exec(ctx)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := res.RowsAffected()
	return int(rowsAffected), err
}

func AuthAccessTokenConstraintErr(_ error) error {
	return nil
}

func buildAuthAccessTokenJSON(opts domain_query.AuthAccessTokenQueryOpts) builder.JsonBuildObjectBuilder {
	return authAccessTokenDefaultJson.
		PropIf(
			opts.IncludeAuthSession,
			"AuthSession",
			Select(buildAuthSessionJSON(opts.AuthSessionQueryOpts)).
				From(authSession).
				Where(authSession.ID.Eq(authAccessToken.AuthSessionID)))
}
