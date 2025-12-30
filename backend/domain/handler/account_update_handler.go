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

func (h *Handler) AccountUpdate(ctx context.Context, cmd command.AccountUpdateCmd) error {
	logger := slogutils.FromContext(ctx).
		With(
			"component", "handler",
			"handler", "AccountUpdate",
		)

	logger.DebugContext(ctx, "Handling account update command", "cmd", cmd)

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
		prevRecord, err := repository.FindAccountByID(ctx, tx, cmd.AccountID, nil)
		if errors.Is(err, repository.ErrNotFound) {
			return types.FieldError{
				Field: "accountId",
				Code:  types.ErrorCodeNotExists,
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

	logger.InfoContext(ctx, "Updated account",
		"accountID", cmd.AccountID,
		"prevOrganisationID", prevOrganisationID,
		"organisationID", organisationID,
		"prevUsername", prevUsername,
		"emailAddress", cmd.EmailAddress,
		"prevRole", prevRole,
		"role", cmd.Role,
	)

	return nil
}
