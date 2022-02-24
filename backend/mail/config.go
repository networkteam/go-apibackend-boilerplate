package mail

import (
	"myvendor.mytld/myproject/backend/domain"
)

// Config stores mail specific configuration
type Config struct {
	// Embed base config for easier passing around
	domain.Config

	DefaultFrom             string
	ErrorReportingRecipient string
	TemplateIDs             struct {
		RegisterConfirm string
	}
}

func DefaultConfig(c domain.Config) Config {
	config := Config{
		Config: c,
	}

	config.DefaultFrom = "info@example.com"
	config.ErrorReportingRecipient = "admin@example.com"

	// TODO Set MailJet template IDs
	config.TemplateIDs.RegisterConfirm = "to-be-set"
	return config
}
