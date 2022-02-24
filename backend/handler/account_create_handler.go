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

func (h *Handler) AccountCreate(ctx context.Context, cmd domain.AccountCreateCmd) error {
	log := logger.FromContext(ctx).
		WithField("component", "handler").
		WithField("handler", "AccountCreate")

	log.
		WithField("cmd", cmd).
		Debug("Handling account create command")

	if err := cmd.Validate(h.config); err != nil {
		return err
	}

	authCtx := authentication.GetAuthContext(ctx)
	if err := authorization.NewAuthorizer(authCtx).AllowsAccountCreateCmd(cmd); err != nil {
		return err
	}

	err := repository.Transactional(ctx, h.db, func(tx *sql.Tx) error {
		account, err := cmd.NewAccount(h.config)
		if err != nil {
			return err
		}
		err = repository.InsertAccount(ctx, tx, repository.AccountToChangeSet(account))
		if err != nil {
			if constraintErr := repository.AccountConstraintErr(err); constraintErr != nil {
				return constraintErr
			}
			return errors.Wrap(err, "inserting account")
		}

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "running transaction")
	}

	var organisationID string
	if cmd.OrganisationID.Valid {
		organisationID = cmd.OrganisationID.UUID.String()
	}

	log.
		WithField("accountID", cmd.AccountID).
		WithField("organisationID", organisationID).
		WithField("username", cmd.EmailAddress).
		WithField("role", cmd.Role).
		Info("Created account")

	return nil
}
