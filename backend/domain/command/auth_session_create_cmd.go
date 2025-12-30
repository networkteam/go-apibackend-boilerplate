package command

import (
	"time"

	"github.com/friendsofgo/errors"
	"github.com/gofrs/uuid"

	"myvendor.mytld/myproject/backend/domain/model"
	"myvendor.mytld/myproject/backend/domain/types"
)

type AuthSessionCreateCmd struct {
	AuthSessionID uuid.UUID
	AccountID     uuid.UUID

	// Values for initial access token
	AuthAccessTokenID uuid.UUID
	ExpiresAt         time.Time
	DeviceInfo        model.DeviceInfo
}

func NewAuthSessionCreateCmd(accountID uuid.UUID, expiresAt time.Time) (cmd AuthSessionCreateCmd, err error) {
	authSessionID, err := uuid.NewV7()
	if err != nil {
		return cmd, errors.Wrap(err, "generating auth session ID")
	}

	accessTokenID, err := uuid.NewV4()
	if err != nil {
		return cmd, errors.Wrap(err, "generating access token ID")
	}

	cmd = AuthSessionCreateCmd{
		AuthSessionID: authSessionID,
		AccountID:     accountID,
		DeviceInfo:    make(model.DeviceInfo),

		AuthAccessTokenID: accessTokenID,
		ExpiresAt:         expiresAt,
	}

	return cmd, nil
}

func (c AuthSessionCreateCmd) Validate() error {
	if c.AuthSessionID == uuid.Nil {
		return types.FieldError{
			Field: "authSessionID",
			Code:  types.ErrorCodeRequired,
		}
	}

	if c.AuthAccessTokenID == uuid.Nil {
		return types.FieldError{
			Field: "accessTokenID",
			Code:  types.ErrorCodeRequired,
		}
	}

	if c.AccountID == uuid.Nil {
		return types.FieldError{
			Field: "accountID",
			Code:  types.ErrorCodeRequired,
		}
	}

	return nil
}

func (c AuthSessionCreateCmd) NewAuthSession() model.AuthSession {
	return model.AuthSession{
		ID:         c.AuthSessionID,
		AccountID:  c.AccountID,
		DeviceInfo: &c.DeviceInfo,
	}
}

func (c AuthSessionCreateCmd) NewAccessToken() model.AuthAccessToken {
	return model.AuthAccessToken{
		ID:            c.AuthAccessTokenID,
		AuthSessionID: c.AuthSessionID,
		ExpiresAt:     c.ExpiresAt,
	}
}
