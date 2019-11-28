package authentication

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
	GetAccountID() string
}

type OrganisationIDProvider interface {
	GetOrganisationID() string
}

type RoleIdentifierProvider interface {
	GetRoleIdentifier() string
}
