package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/friendsofgo/errors"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgconn"

	"myvendor.mytld/myproject/backend/domain"
)

func accountBuildFindQuery(query squirrel.SelectBuilder) squirrel.SelectBuilder {
	query = query.
		From("accounts").
		LeftJoin("organisations ON organisations.organisation_id = accounts.organisation_id")
	return query
}

func FindAccountByID(ctx context.Context, runner squirrel.BaseRunner, id uuid.UUID) (domain.Account, error) {
	query := queryBuilder(runner).
		Select(buildAccountJSON())
	query = accountBuildFindQuery(query)

	row := query.
		Where(squirrel.Eq{account_id: id}).
		QueryRowContext(ctx)
	return accountScanJsonRow(row)
}

func FindAllAccounts(ctx context.Context, runner squirrel.BaseRunner, paging Paging, filter domain.AccountsQuery) (result []domain.Account, err error) {
	query := queryBuilder(runner).
		Select(buildAccountJSON())
	query = accountBuildFindQuery(query)

	query, err = applyPaging(query, paging, accountSortFields)
	if err != nil {
		return
	}
	query = applyAccountFilter(query, filter)

	rows, err := query.QueryContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "executing query")
	}
	defer rows.Close()
	for rows.Next() {
		record, err := accountScanJsonRow(rows)
		if err != nil {
			return nil, errors.Wrap(err, "scanning row")
		}
		result = append(result, record)
	}
	return
}

func applyAccountFilter(query squirrel.SelectBuilder, filter domain.AccountsQuery) squirrel.SelectBuilder {
	if len(filter.IDs) > 0 {
		query = query.Where(
			squirrel.Eq{account_id: filter.IDs},
		)
	}
	if filter.Q != nil {
		query = query.Where(
			squirrel.Or{
				squirrel.ILike{account_emailAddress: "%" + *filter.Q + "%"},
				squirrel.ILike{organisation_name: "%" + *filter.Q + "%"},
			},
		)
	}
	if filter.OrganisationID != nil {
		query = query.Where(
			squirrel.Eq{account_organisationID: *filter.OrganisationID},
		)
	}
	if filter.ExcludeRole != nil {
		query = query.Where(
			squirrel.NotEq{account_role: *filter.ExcludeRole},
		)
	}
	return query
}

func CountAllAccounts(ctx context.Context, runner squirrel.BaseRunner, filter domain.AccountsQuery) (count int, err error) {
	query := queryBuilder(runner).
		Select("COUNT(*)")
	query = accountBuildFindQuery(query)

	query = applyAccountFilter(query, filter)

	row := query.QueryRowContext(ctx)
	err = row.Scan(&count)
	return
}

func FindAccountByEmailAddress(ctx context.Context, runner squirrel.BaseRunner, emailAddress string) (domain.Account, error) {
	query := queryBuilder(runner).
		Select(buildAccountJSON())
	query = accountBuildFindQuery(query).
		Where(squirrel.Eq{fmt.Sprintf("LOWER(%s)", account_emailAddress): strings.ToLower(emailAddress)})

	row := query.QueryRowContext(ctx)
	return accountScanJsonRow(row)
}

func InsertAccount(ctx context.Context, runner squirrel.BaseRunner, changeSet AccountChangeSet) error {
	_, err := queryBuilder(runner).
		Insert("accounts").
		SetMap(changeSet.toMap()).
		ExecContext(ctx)
	return err
}

func UpdateAccount(ctx context.Context, runner squirrel.BaseRunner, id uuid.UUID, changeSet AccountChangeSet) error {
	res, err := queryBuilder(runner).
		Update("accounts").
		Where(squirrel.Eq{account_id: id}).
		SetMap(changeSet.toMap()).
		ExecContext(ctx)
	if err != nil {
		return errors.Wrap(err, "executing update")
	}
	return assertRowsAffected(res, "update")
}

func DeleteAccount(ctx context.Context, runner squirrel.BaseRunner, id uuid.UUID) error {
	res, err := queryBuilder(runner).
		Delete("accounts").
		Where(squirrel.Eq{account_id: id}).
		ExecContext(ctx)
	if err != nil {
		return errors.Wrap(err, "executing delete")
	}
	return assertRowsAffected(res, "delete")
}

func AccountConstraintErr(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch {
		case pgErr.Code == pgErrCode_unique_violation && pgErr.ConstraintName == "accounts_email_address_idx":
			return domain.FieldError{
				Field: "emailAddress",
				Code:  domain.ErrorCodeAlreadyExists,
			}
		}
	}
	return nil
}

func buildAccountJSON() string {
	return accountDefaultSelectJson.
		ToSql()
}
