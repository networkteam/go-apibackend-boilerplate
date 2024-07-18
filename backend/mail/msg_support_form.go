package mail

import (
	"io"

	"github.com/friendsofgo/errors"
	gomail "github.com/wneessen/go-mail"
)

type SupportFormMsg struct {
	SenderEmailAddress string
	SenderName         string
	OrganisationName   string
	Subject            string
	Message            string
	FileName           string
	AttachedFile       io.Reader
}

func (m SupportFormMsg) ToMessage(config Config) (*gomail.Msg, error) {
	sender := m.SenderEmailAddress
	recipient := config.DefaultFrom

	subject, body, err := executeTemplate("support_form", m)
	if err != nil {
		return nil, errors.Wrap(err, "executing template")
	}

	msg := gomail.NewMsg()
	err = msg.To(recipient)
	if err != nil {
		return nil, errors.Wrap(err, "setting to")
	}
	err = msg.From(sender)
	if err != nil {
		return nil, errors.Wrap(err, "setting from")
	}
	msg.Subject(subject)
	msg.SetBodyString("text/plain", body)

	if m.AttachedFile != nil {
		err = msg.AttachReader(m.FileName, m.AttachedFile)
		if err != nil {
			return nil, errors.Wrap(err, "attaching file")
		}
	}

	return msg, nil
}
