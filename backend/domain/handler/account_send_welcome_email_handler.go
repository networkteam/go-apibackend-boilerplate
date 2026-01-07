package handler

import (
	"context"

	"github.com/friendsofgo/errors"
	"github.com/networkteam/slogutils"

	"myvendor.mytld/myproject/backend/domain/command"
	"myvendor.mytld/myproject/backend/mail"
)

func (h *Handler) AccountSendWelcomeEmail(ctx context.Context, cmd command.AccountSendWelcomeEmail) error {
	logger := slogutils.FromContext(ctx).
		With(
			"component", "handler",
			"handler", "AccountSendWelcomeEmail",
		)

	logger.DebugContext(ctx, "Handling account send welcome email command", "cmd", cmd)

	msg := mail.WelcomeEmailMsg{
		EmailAddress: cmd.EmailAddress,
	}

	err := h.mailer.Send(ctx, msg)
	if err != nil {
		return errors.Wrap(err, "sending welcome email")
	}

	logger.InfoContext(ctx, "Sent welcome email",
		"accountID", cmd.AccountID,
		"emailAddress", cmd.EmailAddress,
	)

	return nil
}
