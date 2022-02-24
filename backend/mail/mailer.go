package mail

import (
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

type MessageDataProvider interface {
	TemplateID(config Config) string
	Recipient(config Config) string
	Subject(config Config) string
	Data(config Config) interface{}
}

func (m *Mailer) Send(msg MessageDataProvider) error {
	message, err := BuildMailJetMessage(
		msg.TemplateID(m.config),
		msg.Recipient(m.config),
		m.config.DefaultFrom,
		msg.Subject(m.config),
		m.config.ErrorReportingRecipient,
		msg.Data(m.config),
	)
	if err != nil {
		return errors.Wrap(err, "building message")
	}
	err = m.sender.Send(message)
	if err != nil {
		return errors.Wrap(err, "sending message")
	}
	return nil
}
