package mail

import (
	"github.com/friendsofgo/errors"
	gomail "github.com/wneessen/go-mail"
)

type WelcomeEmailMsg struct {
	EmailAddress string
}

func (m WelcomeEmailMsg) ToMessage(config Config) (*gomail.Msg, error) {
	subject, body, err := executeTemplate("welcome_email", m)
	if err != nil {
		return nil, errors.Wrap(err, "executing template")
	}

	msg := gomail.NewMsg()
	err = msg.To(m.EmailAddress)
	if err != nil {
		return nil, errors.Wrap(err, "setting to")
	}
	err = msg.From(config.DefaultFrom)
	if err != nil {
		return nil, errors.Wrap(err, "setting from")
	}
	msg.Subject(subject)
	msg.SetBodyString("text/plain", body)

	return msg, nil
}
