package auth

import (
	"github.com/gofrs/uuid"

	"myvendor.mytld/myproject/backend/domain/types"
)

type FixedAccessTokenData struct {
	AccountID      uuid.UUID
	OrganisationID *uuid.UUID
	Role           types.Role
	AuthSessionID  uuid.UUID
	UserID         *uuid.UUID
}

func (d FixedAccessTokenData) GetAccountID() uuid.UUID {
	return d.AccountID
}

func (d FixedAccessTokenData) GetRole() types.Role {
	return d.Role
}
