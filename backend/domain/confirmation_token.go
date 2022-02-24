package domain

import (
	"time"

	"github.com/friendsofgo/errors"
	"github.com/gofrs/uuid"

	security_helper "myvendor.mytld/myproject/backend/security/helper"
)

type ConfirmationToken struct {
	Token     string
	AccountID uuid.UUID
	Expires   time.Time
}

const (
	ConfirmationTokenLength             = 32
	ConfirmationTokenExpirationDuration = 24 * time.Hour
)

func NewConfirmationToken(timeSource TimeSource, account Account) (ConfirmationToken, error) {
	token, err := security_helper.GenerateRandomString(ConfirmationTokenLength)
	if err != nil {
		return ConfirmationToken{}, errors.Wrap(err, "generating random token")
	}
	return ConfirmationToken{
		Token:     token,
		AccountID: account.ID,
		Expires:   timeSource.Now().Add(ConfirmationTokenExpirationDuration),
	}, nil
}
