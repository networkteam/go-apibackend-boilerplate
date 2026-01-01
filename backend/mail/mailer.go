package mail

import (
	"context"

	"github.com/friendsofgo/errors"
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

// MessageProvider creates a Message from configuration.
type MessageProvider interface {
	ToMessage(Config) (*Message, error)
}

func (m *Mailer) Send(ctx context.Context, msg MessageProvider) error {
	if m == nil {
		return errors.New("nil mailer cannot send")
	}
	if msg == nil {
		return errors.New("nil message provider given")
	}

	message, err := msg.ToMessage(m.config)
	if err != nil {
		return errors.Wrap(err, "building message")
	}

	err = m.sender.Send(ctx, message.Msg)
	if err != nil {
		return errors.Wrap(err, "sending message")
	}

	return nil
}
