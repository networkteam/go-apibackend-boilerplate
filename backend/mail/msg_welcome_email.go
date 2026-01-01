package mail

import (
	"github.com/friendsofgo/errors"
)

type WelcomeEmailMsg struct {
	EmailAddress string
}

func (m WelcomeEmailMsg) ToMessage(config Config) (*Message, error) {
	body, err := ExecuteTextTemplate("welcome_email", m)
	if err != nil {
		return nil, errors.Wrap(err, "executing template")
	}

	msg := NewMessage(config)
	err = msg.To(m.EmailAddress)
	if err != nil {
		return nil, errors.Wrap(err, "setting to")
	}
	err = msg.From(config.DefaultFrom)
	if err != nil {
		return nil, errors.Wrap(err, "setting from")
	}
	msg.Subject("Welcome to myproject")
	msg.SetBodyString("text/plain", body)

	return msg, nil
}
