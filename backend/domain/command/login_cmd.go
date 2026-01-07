package command

import (
	"github.com/gofrs/uuid"

	"myvendor.mytld/myproject/backend/domain/model"
	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/security/authentication"
)

type LoginDataProvider interface {
	GetAccountID() uuid.UUID
	GetPasswordHash() []byte
}

type LoginCmd struct {
	EmailAddress    string
	Password        string
	TokenExpiryType authentication.TokenExpiryType
	DeviceInfo      model.DeviceInfo

	// OnAuthenticated is called after the account can be treated as authenticated (if set)
	OnAuthenticated func(accountID uuid.UUID, role types.Role)
}

func NewLoginCmd(emailAddress, password string, keepMeLoggedIn *bool) LoginCmd {
	cmd := LoginCmd{
		EmailAddress:    SanitizeEmailAddress(emailAddress),
		Password:        SanitizePassword(password),
		TokenExpiryType: authentication.TokenExpiryTypeDefault,
		DeviceInfo:      make(model.DeviceInfo),
	}

	if keepMeLoggedIn != nil && *keepMeLoggedIn {
		cmd.TokenExpiryType = authentication.TokenExpiryTypeExtended
	}

	return cmd
}
