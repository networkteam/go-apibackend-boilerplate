package authentication

import "github.com/gofrs/uuid"

type AuthTokenDataProvider interface {
	TokenSecretProvider
	AccountIDProvider
	OrganisationIDProvider
	RoleIdentifierProvider
}

type TokenSecretProvider interface {
	GetTokenSecret() []byte
}

type AccountIDProvider interface {
	GetAccountID() uuid.UUID
}

type OrganisationIDProvider interface {
	GetOrganisationID() uuid.NullUUID
}

type RoleIdentifierProvider interface {
	GetRoleIdentifier() string
}
