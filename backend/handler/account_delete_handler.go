package handler

import (
	"context"
	"database/sql"

	logger "github.com/apex/log"
	"github.com/friendsofgo/errors"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/persistence/repository"
	"myvendor.mytld/myproject/backend/security/authentication"
	"myvendor.mytld/myproject/backend/security/authorization"
)

func (h *Handler) AccountDelete(ctx context.Context, cmd domain.AccountDeleteCmd) error {
	log := logger.FromContext(ctx).
		WithField("component", "handler").
		WithField("handler", "AccountDelete")

	log.
		WithField("cmd", cmd).
		Debug("Handling account delete command")

	authCtx := authentication.GetAuthContext(ctx)
	if err := authorization.NewAuthorizer(authCtx).AllowsAccountDeleteCmd(cmd); err != nil {
		return err
	}

	var (
		organisationID string
		username       string
		role           domain.Role
	)
	err := repository.Transactional(ctx, h.db, func(tx *sql.Tx) error {
		record, err := repository.FindAccountByID(ctx, tx, cmd.AccountID)
		if err == repository.ErrNotFound {
			return domain.FieldError{
				Field: "accountId",
				Code:  domain.ErrorCodeNotExists,
			}
		} else if err != nil {
			return errors.Wrap(err, "finding account")
		}

		// For logging
		username = record.EmailAddress
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
		return errors.Wrap(err, "running transaction")
	}

	log.
		WithField("accountID", cmd.AccountID).
		WithField("organisationID", organisationID).
		WithField("username", username).
		WithField("role", role).
		Info("Deleted account")

	return nil
}
