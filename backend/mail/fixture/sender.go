package fixture

import (
	"bytes"

	"github.com/apex/log"
	"github.com/friendsofgo/errors"
	"gopkg.in/gomail.v2"

	"myvendor.mytld/myproject/backend/mail"
)

type Sender struct {
	LastMail     string
	TemplatePath string
	SendCallback func(message *gomail.Message)
}

func NewSender() *Sender {
	return &Sender{}
}

var _ mail.Sender = &Sender{}

func (m *Sender) Send(message *gomail.Message) error {
	if m.SendCallback != nil {
		defer m.SendCallback(message)
	}

	var b = new(bytes.Buffer)
	_, err := message.WriteTo(b)
	if err != nil {
		return errors.Wrap(err, "writing message to buffer")
	}
	m.LastMail = b.String()

	log.
		WithField("message", b.String()).
		Debug("Sent message via fixture mailer")
	return nil
}
