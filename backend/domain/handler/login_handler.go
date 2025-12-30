package handler

import (
	"context"
	"database/sql"

	"github.com/friendsofgo/errors"
	"github.com/networkteam/slogutils"

	"myvendor.mytld/myproject/backend/domain/command"
	"myvendor.mytld/myproject/backend/domain/model"
	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/persistence/repository"
	security_helper "myvendor.mytld/myproject/backend/security/helper"
)

func (h *Handler) Login(ctx context.Context, cmd command.LoginCmd) (err error) {
	logger := slogutils.FromContext(ctx).
		With(
			"component", "handler",
			"handler", "Login",
		)

	logger.DebugContext(ctx, "Handling login", "emailAddress", cmd.EmailAddress)

	var accountFound bool
	account, err := repository.FindAccountByEmailAddress(ctx, h.db, cmd.EmailAddress, nil)
	if errors.Is(err, repository.ErrNotFound) {
		// Use an empty user to have constant password compare times
		account = model.Account{
			PasswordHash: security_helper.DefaultHashForComparison(h.config.HashCost),
		}
	} else if err != nil {
		return errors.Wrap(err, "finding account by username")
	} else {
		accountFound = true
	}

	accountPasswordHash := account.GetPasswordHash()

	err = security_helper.CompareHashAndPassword(accountPasswordHash, []byte(cmd.Password))
	if err != nil || !accountFound {
		// Log warning to find potential attacks
		if !accountFound {
			logger.WarnContext(ctx,
				"Login failed, account not found",
				"emailAddress", cmd.EmailAddress,
				"errorCode", types.ErrorCodeNotExists,
			)
		} else {
			logger.WarnContext(ctx,
				"Login failed, invalid password",
				"emailAddress", cmd.EmailAddress,
				"errorCode", "invalidPassword",
				slogutils.Err(err),
			)
		}

		h.instrumentation.loginFailedCounter.Add(ctx, 1)

		return types.FieldError{
			Field: "username",
			Code:  types.ErrorCodeInvalidCredentials,
		}
	}

	authSession, err := h.generateAuthSession(account.ID, cmd.DeviceInfo)
	if err != nil {
		return errors.Wrap(err, "generating auth session")
	}

	accessToken, err := h.generateAccessToken(authSession.ID, cmd.TokenExpiryType.GetExpiresAt(h.timeSource))
	if err != nil {
		return errors.Wrap(err, "generating access token")
	}

	err = repository.Transactional(ctx, h.db, func(tx *sql.Tx) error {
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

	err = h.setAccessTokenCookieForAccount(ctx, account, accessToken.ID, cmd.TokenExpiryType)
	if err != nil {
		return errors.Wrap(err, "setting access token cookie")
	}

	h.instrumentation.loginSuccessCounter.Add(ctx, 1)

	if cmd.OnAuthenticated != nil {
		cmd.OnAuthenticated(account.ID, account.Role)
	}

	logger.InfoContext(ctx,
		"Login success",
		"emailAddress", cmd.EmailAddress,
		"accountID", account.GetAccountID(),
	)

	return nil
}
