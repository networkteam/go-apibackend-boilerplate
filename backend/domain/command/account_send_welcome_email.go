package command

import "github.com/gofrs/uuid"

type AccountSendWelcomeEmail struct {
	AccountID    uuid.UUID
	EmailAddress string
}
