package mail

import (
	gomail "github.com/wneessen/go-mail"
)

// Message wraps gomail.Msg with configuration for convenient message construction.
type Message struct {
	*gomail.Msg

	Config Config
}

// NewMessage creates a new Message with the given configuration.
func NewMessage(config Config) *Message {
	return &Message{
		Msg:    gomail.NewMsg(),
		Config: config,
	}
}
