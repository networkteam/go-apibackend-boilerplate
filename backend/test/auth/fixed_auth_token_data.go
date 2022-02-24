package auth

import "github.com/gofrs/uuid"

type FixedAuthTokenData struct {
	TokenSecret    []byte
	AccountID      uuid.UUID
	OrganisationID uuid.NullUUID
	RoleIdentifier string
}

func (d FixedAuthTokenData) GetTokenSecret() []byte {
	return d.TokenSecret
}

func (d FixedAuthTokenData) GetAccountID() uuid.UUID {
	return d.AccountID
}

func (d FixedAuthTokenData) GetOrganisationID() uuid.NullUUID {
	return d.OrganisationID
}

func (d FixedAuthTokenData) GetRoleIdentifier() string {
	return d.RoleIdentifier
}
