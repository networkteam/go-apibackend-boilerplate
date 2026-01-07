package authentication

import (
	"github.com/gofrs/uuid"

	"myvendor.mytld/myproject/backend/domain/types"
)

type AccessTokenDataProvider interface {
	AccountIDProvider
	RoleProvider
}

type AccountIDProvider interface {
	GetAccountID() uuid.UUID
}

type RoleProvider interface {
	GetRole() types.Role
}
