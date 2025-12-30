package handler

import (
	"context"

	logger "github.com/apex/log"
	"github.com/friendsofgo/errors"

	"myvendor.mytld/myproject/backend/domain/command"
	"myvendor.mytld/myproject/backend/mail"
)

func (h *Handler) AccountSendWelcomeEmail(ctx context.Context, cmd command.AccountSendWelcomeEmail) error {
	log := logger.FromContext(ctx).
		WithField("component", "handler").
		WithField("handler", "AccountSendWelcomeEmail")

	log.
		WithField("cmd", cmd).
		Debug("Handling account send welcome email command")

	msg := mail.WelcomeEmailMsg{
		EmailAddress: cmd.EmailAddress,
	}

	err := h.mailer.Send(ctx, msg)
	if err != nil {
		return errors.Wrap(err, "sending welcome email")
	}

	log.
		WithField("accountID", cmd.AccountID).
		WithField("emailAddress", cmd.EmailAddress).
		Info("Sent welcome email")

	return nil
}
