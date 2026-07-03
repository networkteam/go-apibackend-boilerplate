package handler

import (
	"context"
	"database/sql"

	"github.com/friendsofgo/errors"
	"github.com/gofrs/uuid"
	"github.com/networkteam/slogutils"

	"myvendor.mytld/myproject/backend/domain/model"
	"myvendor.mytld/myproject/backend/domain/query"
	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/persistence/repository"
	"myvendor.mytld/myproject/backend/security/authentication"
	"myvendor.mytld/myproject/backend/security/authorization"

	"myvendor.mytld/myproject/backend/domain/command"
)

func (h *Handler) AuthSessionRefresh(ctx context.Context, cmd command.AuthSessionRefreshCmd) error {
	logger := slogutils.FromContext(ctx).
		With(
			"component", "handler",
			"handler", "AuthSessionRefresh",
		)
	logger.DebugContext(ctx, "Handling auth session refresh command", "cmd", cmd)

	if err := cmd.Validate(); err != nil {
		return err
	}

	authContext := authentication.GetAuthContext(ctx)
	if err := authorization.NewAuthorizer(authContext).AllowsAuthSessionRefreshCmd(cmd); err != nil {
		return err
	}

	var (
		newAccessTokenID uuid.UUID
		accountID        uuid.UUID
		role             types.Role
	)
	err := repository.Transactional(ctx, h.db, func(tx *sql.Tx) error {
		authSession, err := repository.FindAuthSessionByID(ctx, tx, cmd.AuthSessionID, &query.AuthSessionQueryOpts{
			IncludeAccount: true,
		})
		if errors.Is(err, repository.ErrNotFound) {
			return types.NotExistsError("authSessionId")
		} else if err != nil {
			return errors.Wrap(err, "finding auth session by ID")
		}
		accountID = authSession.AccountID
		role = authSession.Account.Role

		accessToken, err := h.generateAccessToken(cmd.AuthSessionID, cmd.ExpiresAt)
		if err != nil {
			return errors.Wrap(err, "generating auth access token")
		}
		newAccessTokenID = accessToken.ID

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

	tokenExpiryType := authContext.GetTokenExpiryType() // Use the same expiry type as before
	err = h.setAccessTokenCookieForAccount(ctx, model.Account{
		ID:   accountID,
		Role: role,
	}, newAccessTokenID, cmd.AuthSessionID, tokenExpiryType)
	if err != nil {
		return errors.Wrap(err, "setting access token cookie")
	}

	logger.InfoContext(ctx,
		"Refreshed auth session",
		"authSessionID", cmd.AuthSessionID,
	)

	return nil
}
