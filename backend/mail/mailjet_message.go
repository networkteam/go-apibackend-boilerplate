package mail

import (
	"encoding/json"

	"github.com/friendsofgo/errors"
	"gopkg.in/gomail.v2"
)

func BuildMailJetMessage(templateID, recipient, sender, subject, errorReportingRecipient string, data interface{}) (*gomail.Message, error) {
	if recipient == "" {
		return nil, errors.New("recipient is needed for message")
	}

	if sender == "" {
		return nil, errors.New("sender is needed for message")
	}

	renderedVariables, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Wrap(err, "encoding template data to JSON")
	}

	m := gomail.NewMessage()
	m.SetHeader("To", recipient)
	m.SetHeader("From", sender)
	m.SetHeader("Subject", subject)
	m.SetHeaders(map[string][]string{
		"X-MJ-TEMPLATEID":             {templateID},
		"X-MJ-TEMPLATELANGUAGE":       {"1"},
		"X-MJ-TEMPLATEERRORREPORTING": {errorReportingRecipient},
		"X-MJ-VARS":                   {string(renderedVariables)},
	})
	m.SetBody("text/plain", "Mailjet template mail, only rendered when send via mailjet\n\n"+string(renderedVariables))

	return m, nil
}
