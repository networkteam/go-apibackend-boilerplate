package handler

import (
	"context"
	"database/sql"

	"github.com/friendsofgo/errors"
	"github.com/networkteam/slogutils"

	"myvendor.mytld/myproject/backend/persistence/repository"
	"myvendor.mytld/myproject/backend/security/authentication"
	"myvendor.mytld/myproject/backend/security/authorization"

	"myvendor.mytld/myproject/backend/domain/command"
)

func (h *Handler) AuthSessionCreate(ctx context.Context, cmd command.AuthSessionCreateCmd) error {
	logger := slogutils.FromContext(ctx).
		With(
			"component", "handler",
			"handler", "AuthSessionCreate",
		)
	logger.DebugContext(ctx, "Handling auth session create command", "cmd", cmd)

	if err := cmd.Validate(); err != nil {
		return err
	}

	authContext := authentication.GetAuthContext(ctx)
	if err := authorization.NewAuthorizer(authContext).AllowsAuthSessionCreateCmd(cmd); err != nil {
		return err
	}

	err := repository.Transactional(ctx, h.db, func(tx *sql.Tx) error {
		authSession := cmd.NewAuthSession()
		accessToken := cmd.NewAccessToken()

		err := repository.InsertAuthSession(ctx, tx, repository.AuthSessionToChangeSet(authSession))
		if err != nil {
			if constraintErr := repository.AuthSessionConstraintErr(err); constraintErr != nil {
				return constraintErr
			}
			return errors.Wrap(err, "inserting auth session")
		}

		err = repository.InsertAuthAccessToken(ctx, tx, repository.AuthAccessTokenToChangeSet(accessToken))
		if err != nil {
			if constraintErr := repository.AuthAccessTokenConstraintErr(err); constraintErr != nil {
				return constraintErr
			}
			return errors.Wrap(err, "inserting auth access token")
		}

		return nil
	})
	if err != nil {
		return err
	}

	logger.InfoContext(ctx,
		"Created auth session",
		"accountID", cmd.AccountID,
		"authSessionID", cmd.AuthSessionID,
	)

	return nil
}
