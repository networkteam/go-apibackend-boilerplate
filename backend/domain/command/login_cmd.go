package command

import (
	"github.com/gofrs/uuid"
)

type LoginDataProvider interface {
	GetAccountID() uuid.UUID
	GetPasswordHash() []byte
}

type LoginCmd struct {
	EmailAddress   string
	Password       string
	ExtendedExpiry bool

	Account LoginDataProvider
}

func NewLoginCmd(email, password string) LoginCmd {
	return LoginCmd{
		EmailAddress: email,
		Password:     password,
	}
}
