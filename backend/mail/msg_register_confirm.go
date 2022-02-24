package mail

import (
	"fmt"
)

type RegisterConfirmMsg struct {
	EmailAddress      string
	ConfirmationToken string
}

var _ MessageDataProvider = RegisterConfirmMsg{}

func (m RegisterConfirmMsg) TemplateID(config Config) string {
	return config.TemplateIDs.RegisterConfirm
}

func (m RegisterConfirmMsg) Recipient(config Config) string {
	return m.EmailAddress
}

func (m RegisterConfirmMsg) Subject(config Config) string {
	return fmt.Sprintf("Bestätige Deine Registrierung bei %s", config.AppName)
}

func (m RegisterConfirmMsg) Data(config Config) interface{} {
	return map[string]interface{}{
		"email_address":    m.EmailAddress,
		"confirmation_url": config.BuildURL("/register/confirm/" + m.ConfirmationToken),
	}
}
