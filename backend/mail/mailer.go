package mail

import (
	"github.com/friendsofgo/errors"
	"gopkg.in/gomail.v2"
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
	ToMessage(Config) (*gomail.Message, error)
}

func (m *Mailer) Send(msg MessageProvider) error {
	message, err := msg.ToMessage(m.config)
	if err != nil {
		return errors.Wrap(err, "building message")
	}
	err = m.sender.Send(message)
	if err != nil {
		return errors.Wrap(err, "sending message")
	}
	return nil
}
