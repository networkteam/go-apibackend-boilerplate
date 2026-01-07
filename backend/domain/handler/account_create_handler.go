package handler

import (
	"context"
	"database/sql"

	"github.com/friendsofgo/errors"
	"github.com/networkteam/slogutils"

	"myvendor.mytld/myproject/backend/domain/command"
	"myvendor.mytld/myproject/backend/persistence/repository"
	"myvendor.mytld/myproject/backend/security/authentication"
	"myvendor.mytld/myproject/backend/security/authorization"
)

func (h *Handler) AccountCreate(ctx context.Context, cmd command.AccountCreateCmd) error {
	logger := slogutils.FromContext(ctx).
		With(
			"component", "handler",
			"handler", "AccountCreate",
		)

	logger.DebugContext(ctx, "Handling account create command", "cmd", cmd)

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

	logger.InfoContext(ctx, "Created account",
		"accountID", cmd.AccountID,
		"organisationID", organisationID,
		"emailAddress", cmd.EmailAddress,
		"role", cmd.Role,
	)

	return nil
}
