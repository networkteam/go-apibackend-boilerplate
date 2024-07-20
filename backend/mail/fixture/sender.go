package fixture

import (
	"bytes"
	"context"
	std_errors "errors"

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

var (
	errNoSenderSet    = std_errors.New("no sender set")
	errNoRecipientSet = std_errors.New("no recipient set")
)

func (m *Sender) Send(_ context.Context, message *gomail.Msg) error {
	if m.SendCallback != nil {
		defer m.SendCallback(message)
	}

	fromHeaders := message.GetFromString()
	if len(fromHeaders) == 0 {
		return errors.WithStack(errNoSenderSet)
	}

	toHeaders := message.GetToString()
	if len(toHeaders) == 0 {
		return errors.WithStack(errNoRecipientSet)
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
