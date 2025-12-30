package handler

import (
	"context"
	std_errors "errors"

	fog_errors "github.com/friendsofgo/errors"
	"github.com/networkteam/slogutils"

	"myvendor.mytld/myproject/backend/domain/command"
	"myvendor.mytld/myproject/backend/domain/model"
	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/persistence/repository"
	security_helper "myvendor.mytld/myproject/backend/security/helper"
)

var ErrLoginInvalidCredentials = std_errors.New("invalid credentials")

func (h *Handler) Login(ctx context.Context, cmd command.LoginCmd) (err error) {
	logger := slogutils.FromContext(ctx).
		With("handler", "login")

	logger.DebugContext(ctx, "Handling login", "emailAddress", cmd.EmailAddress)

	account := cmd.Account
	if cmd.Account == nil {
		// Use an empty user to have constant password compare times
		account = model.Account{
			PasswordHash: security_helper.DefaultHashForComparison(h.config.HashCost),
		}
	}

	err = security_helper.CompareHashAndPassword(account.GetPasswordHash(), []byte(cmd.Password))
	if err != nil || cmd.Account == nil {
		// Log warning to find potential attacks
		if cmd.Account == nil {
			logger.WarnContext(ctx, "Login failed, account not found",
				"emailAddress", cmd.EmailAddress,
				"errorCode", types.ErrorCodeNotExists,
			)
		} else {
			logger.WarnContext(ctx, "Login failed, invalid password",
				"emailAddress", cmd.EmailAddress,
				"errorCode", "invalidPassword",
				slogutils.Err(err),
			)
		}

		h.instrumentation.loginFailedCounter.Add(ctx, 1)

		return ErrLoginInvalidCredentials
	}

	now := h.timeSource.Now()
	ptrNow := &now
	err = repository.UpdateAccount(ctx, h.db, account.GetAccountID(), repository.AccountChangeSet{LastLogin: &ptrNow})
	if err != nil {
		return fog_errors.Wrap(err, "updating account last login")
	}

	h.instrumentation.loginSuccessCounter.Add(ctx, 1)

	logger.InfoContext(ctx, "Login success",
		"emailAddress", cmd.EmailAddress,
		"accountID", account.GetAccountID(),
	)

	return nil
}
