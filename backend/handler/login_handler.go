package handler

import (
	"context"
	std_errors "errors"

	logger "github.com/apex/log"
	fog_errors "github.com/friendsofgo/errors"
	"go.opentelemetry.io/otel/metric"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/persistence/repository"
	security_helper "myvendor.mytld/myproject/backend/security/helper"
)

var ErrLoginInvalidCredentials = std_errors.New("invalid credentials")

//nolint:gochecknoglobals
var (
	loginSuccessCounter = mustInstrument(meter.Int64Counter(
		"login.success.counter",
		metric.WithDescription("Number of successful logins."),
		metric.WithUnit("{call}"),
	))
	loginFailedCounter = mustInstrument(meter.Int64Counter(
		"login.failed.counter",
		metric.WithDescription("Number of failed logins."),
		metric.WithUnit("{call}"),
	))
)

func (h *Handler) Login(ctx context.Context, cmd domain.LoginCmd) (err error) {
	log := logger.
		FromContext(ctx).
		WithField("handler", "login")

	log.
		WithField("emailAddress", cmd.EmailAddress).
		Debug("Handling login")

	account := cmd.Account
	if cmd.Account == nil {
		// Use an empty user to have constant password compare times
		account = domain.Account{
			PasswordHash: security_helper.DefaultHashForComparison(h.config.HashCost),
		}
	}

	err = security_helper.CompareHashAndPassword(account.GetPasswordHash(), []byte(cmd.Password))
	if err != nil || cmd.Account == nil {
		// Log warning to find potential attacks
		if cmd.Account == nil {
			log.
				WithField("emailAddress", cmd.EmailAddress).
				WithField("errorCode", domain.ErrorCodeNotExists).
				Warn("Login failed, account not found")
		} else {
			log.
				WithField("emailAddress", cmd.EmailAddress).
				WithField("errorCode", "invalidPassword").
				WithError(err).
				Warn("Login failed, invalid password")
		}

		loginFailedCounter.Add(ctx, 1)

		return ErrLoginInvalidCredentials
	}

	now := h.timeSource.Now()
	ptrNow := &now
	err = repository.UpdateAccount(ctx, h.db, account.GetAccountID(), repository.AccountChangeSet{LastLogin: &ptrNow})
	if err != nil {
		return fog_errors.Wrap(err, "updating account last login")
	}

	loginSuccessCounter.Add(ctx, 1)

	log.
		WithField("emailAddress", cmd.EmailAddress).
		WithField("accountID", account.GetAccountID()).
		Info("Login success")

	return nil
}
