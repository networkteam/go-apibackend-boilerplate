package fixture

import (
	"bytes"
	"context"

	"github.com/apex/log"
	"github.com/friendsofgo/errors"
	gomail "github.com/wneessen/go-mail"

	"myvendor.mytld/myproject/backend/mail"
)

type Sender struct {
	LastMail     string
	TemplatePath string
	SendCallback func(message *gomail.Msg)
}

func NewSender() *Sender {
	return &Sender{}
}

var _ mail.Sender = &Sender{}

func (m *Sender) Send(_ context.Context, message *gomail.Msg) error {
	if m.SendCallback != nil {
		defer m.SendCallback(message)
	}

	fromHeaders := message.GetFromString()
	if len(fromHeaders) == 0 {
		return errors.New("no sender set")
	}

	toHeaders := message.GetToString()
	if len(toHeaders) == 0 {
		return errors.New("no recipient(s) set")
	}

	var buf = new(bytes.Buffer)
	_, err := message.WriteTo(buf)
	if err != nil {
		return errors.Wrap(err, "writing message to buffer")
	}

	m.LastMail = buf.String()

	log.
		WithField("message", buf.String()).
		Debug("Sent message via fixture mailer")

	return nil
}
