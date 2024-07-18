package mail

import (
	"context"

	"github.com/friendsofgo/errors"
	gomail "github.com/wneessen/go-mail"
)

type Mailer struct {
	sender Sender
	config Config
}

func NewMailer(sender Sender, config Config) *Mailer {
	return &Mailer{
		sender: sender,
		config: config,
	}
}

type MessageProvider interface {
	ToMessage(Config) (*gomail.Msg, error)
}

func (m *Mailer) Send(ctx context.Context, msg MessageProvider) error {
	message, err := msg.ToMessage(m.config)
	if err != nil {
		return errors.Wrap(err, "building message")
	}

	err = m.sender.Send(ctx, message)
	if err != nil {
		return errors.Wrap(err, "sending message")
	}

	return nil
}
