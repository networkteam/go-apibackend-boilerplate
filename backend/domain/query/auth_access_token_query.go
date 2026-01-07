package query

import (
	"github.com/gofrs/uuid"
)

type AuthAccessTokenQuery struct {
	Opts          AuthAccessTokenQueryOpts
	AccessTokenID uuid.UUID
}

type AuthAccessTokensQuery struct{}

type AuthAccessTokenQueryOpts struct {
	IncludeAuthSession   bool
	AuthSessionQueryOpts *AuthSessionQueryOpts
}
