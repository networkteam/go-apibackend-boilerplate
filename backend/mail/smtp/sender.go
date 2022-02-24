package smtp

import (
	"github.com/friendsofgo/errors"
	"gopkg.in/gomail.v2"

	"myvendor.mytld/myproject/backend/mail"
)

type sender struct {
	Dialer *gomail.Dialer
}

var _ mail.Sender = new(sender)

func NewSender(host string, port int, username string, password string) mail.Sender {
	gomailDialer := gomail.NewDialer(host, port, username, password)

	return &sender{
		gomailDialer,
	}
}

func (m *sender) Send(message *gomail.Message) error {
	err := m.Dialer.DialAndSend(message)
	if err != nil {
		return errors.Wrap(err, "sending message")
	}
	return nil
}
