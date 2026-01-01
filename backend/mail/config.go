package mail

import "myvendor.mytld/myproject/backend/domain"

// Config stores mail specific configuration
type Config struct {
	// Embed base config for easier passing around
	domain.Config

	// DefaultFrom is the default sender email address
	DefaultFrom string
	// SenderName is the display name for the sender (e.g., "My App Support")
	SenderName string
	// HelpdeskEmailAddress is the support/helpdesk contact email
	HelpdeskEmailAddress string
}

func DefaultConfig(c domain.Config) Config {
	return Config{
		Config:               c,
		DefaultFrom:          "app@example.com",
		SenderName:           "My App",
		HelpdeskEmailAddress: "support@example.com",
	}
}
