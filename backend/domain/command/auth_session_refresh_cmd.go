package command

import (
	"time"

	"github.com/gofrs/uuid"

	"myvendor.mytld/myproject/backend/domain/types"
)

type AuthSessionRefreshCmd struct {
	AuthSessionID uuid.UUID
	ExpiresAt     time.Time
}

func NewAuthSessionRefreshCmd(authSessionID uuid.UUID, expiresAt time.Time) (cmd AuthSessionRefreshCmd, err error) {
	cmd = AuthSessionRefreshCmd{
		AuthSessionID: authSessionID,
		ExpiresAt:     expiresAt,
	}

	return cmd, nil
}

func (c AuthSessionRefreshCmd) Validate() error {
	if c.AuthSessionID == uuid.Nil {
		return types.FieldError{
			Field: "authSessionID",
			Code:  types.ErrorCodeRequired,
		}
	}

	return nil
}
