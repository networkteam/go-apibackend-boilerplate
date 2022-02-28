package mail

import (
	"io"

	"github.com/friendsofgo/errors"
	"gopkg.in/gomail.v2"
)

type SupportFormMsg struct {
	SenderEmailAddress string
	SenderName         string
	OrganisationName   string
	Subject            string
	Message            string
	FileName           string
	FileSize           int64
	AttachedFile       io.Reader
}

func (m SupportFormMsg) ToMessage(config Config) (*gomail.Message, error) {
	sender := m.SenderEmailAddress
	recipient := config.DefaultFrom

	subject, body, err := executeTemplate("support_form", m)
	if err != nil {
		return nil, errors.Wrap(err, "executing template")
	}

	msg := gomail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", sender)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", body)

	if m.AttachedFile != nil {
		msg.Attach(m.FileName, gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := io.Copy(w, m.AttachedFile)
			return err
		}))
	}

	return msg, nil
}
