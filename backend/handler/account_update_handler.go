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

func (h *Handler) AccountUpdate(ctx context.Context, cmd domain.AccountUpdateCmd) error {
	log := logger.FromContext(ctx).
		WithField("component", "handler").
		WithField("handler", "AccountUpdate")

	log.
		WithField("cmd", cmd).
		Debug("Handling account update command")

	if err := cmd.Validate(h.config); err != nil {
		return err
	}

	authCtx := authentication.GetAuthContext(ctx)
	if err := authorization.NewAuthorizer(authCtx).AllowsAccountUpdateCmd(cmd); err != nil {
		return err
	}

	var (
		prevUsername       string
		prevOrganisationID string
		prevRole           string
	)
	err := repository.Transactional(ctx, h.db, func(tx *sql.Tx) error {
		prevRecord, err := repository.FindAccountByID(ctx, tx, cmd.AccountID)
		if err == repository.ErrNotFound {
			return domain.FieldError{
				Field: "accountId",
				Code:  domain.ErrorCodeNotExists,
			}
		} else if err != nil {
			return errors.Wrap(err, "finding account")
		}

		// For logging
		prevUsername = prevRecord.EmailAddress
		if prevRecord.OrganisationID.Valid {
			prevOrganisationID = prevRecord.OrganisationID.UUID.String()
		}

		changeSet := repository.AccountChangeSet{
			EmailAddress:   &cmd.EmailAddress,
			Role:           &cmd.Role,
			OrganisationID: &cmd.NewOrganisationID,
			// These will be nil if PasswordHash was not changed, so no update will occur
			Secret:       cmd.Secret,
			PasswordHash: cmd.PasswordHash,
		}

		err = repository.UpdateAccount(ctx, tx, prevRecord.ID, changeSet)
		if err != nil {
			if constraintErr := repository.AccountConstraintErr(err); constraintErr != nil {
				return constraintErr
			}
			return errors.Wrap(err, "updating account")
		}

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "running transaction")
	}

	// For logging
	var organisationID string
	if cmd.NewOrganisationID.Valid {
		organisationID = cmd.NewOrganisationID.UUID.String()
	}

	log.
		WithField("accountID", cmd.AccountID).
		WithField("prevOrganisationID", prevOrganisationID).
		WithField("organisationID", organisationID).
		WithField("prevUsername", prevUsername).
		WithField("emailAddress", cmd.EmailAddress).
		WithField("prevRole", prevRole).
		WithField("role", cmd.Role).
		Info("Updated account")

	return nil
}
