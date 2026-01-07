package handler

import (
	"context"
	"database/sql"

	"github.com/friendsofgo/errors"
	"github.com/networkteam/slogutils"

	"myvendor.mytld/myproject/backend/domain/command"
	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/persistence/repository"
	"myvendor.mytld/myproject/backend/security/authentication"
	"myvendor.mytld/myproject/backend/security/authorization"
)

func (h *Handler) AccountDelete(ctx context.Context, cmd command.AccountDeleteCmd) error {
	logger := slogutils.FromContext(ctx).
		With(
			"component", "handler",
			"handler", "AccountDelete",
		)

	logger.DebugContext(ctx, "Handling account delete command", "cmd", cmd)

	authCtx := authentication.GetAuthContext(ctx)
	if err := authorization.NewAuthorizer(authCtx).AllowsAccountDeleteCmd(cmd); err != nil {
		return err
	}

	var (
		organisationID string
		emailAddress   string
		role           types.Role
	)
	err := repository.Transactional(ctx, h.db, func(tx *sql.Tx) error {
		record, err := repository.FindAccountByID(ctx, tx, cmd.AccountID, nil)
		if errors.Is(err, repository.ErrNotFound) {
			return types.FieldError{
				Field: "accountId",
				Code:  types.ErrorCodeNotExists,
			}
		} else if err != nil {
			return errors.Wrap(err, "fetching account")
		}

		// For logging
		emailAddress = record.EmailAddress
		role = record.Role
		if record.OrganisationID.Valid {
			organisationID = record.OrganisationID.UUID.String()
		}

		err = repository.DeleteAccount(ctx, tx, cmd.AccountID)
		if err != nil {
			return errors.Wrap(err, "deleting account")
		}
		return nil
	})
	if err != nil {
		return err
	}

	logger.InfoContext(ctx, "Deleted account",
		"accountID", cmd.AccountID,
		"organisationID", organisationID,
		"emailAddress", emailAddress,
		"role", role,
	)

	return nil
}
