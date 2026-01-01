package mail

import (
	"fmt"
	"io"

	"github.com/friendsofgo/errors"
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

func (m SupportFormMsg) ToMessage(config Config) (*Message, error) {
	sender := m.SenderEmailAddress
	recipient := config.DefaultFrom

	body, err := ExecuteTextTemplate("support_form", m)
	if err != nil {
		return nil, errors.Wrap(err, "executing template")
	}

	msg := NewMessage(config)
	err = msg.To(recipient)
	if err != nil {
		return nil, errors.Wrap(err, "setting to")
	}
	err = msg.From(sender)
	if err != nil {
		return nil, errors.Wrap(err, "setting from")
	}
	msg.Subject(fmt.Sprintf("Neue Kontaktanfrage von %s (%s)", m.SenderName, m.OrganisationName))
	msg.SetBodyString("text/plain", body)

	if m.AttachedFile != nil {
		err = msg.AttachReader(m.FileName, m.AttachedFile)
		if err != nil {
			return nil, errors.Wrap(err, "attaching file")
		}
	}

	return msg, nil
}
