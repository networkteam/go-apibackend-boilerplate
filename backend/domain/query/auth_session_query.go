package query

import (
	"github.com/gofrs/uuid"
)

type AuthSessionQuery struct {
	Opts          *AuthSessionQueryOpts
	AuthSessionID uuid.UUID
}

type AuthSessionsQuery struct {
	IDs []uuid.UUID
}

type AuthSessionQueryOpts struct {
	IncludeAccount   bool
	AccountQueryOpts *AccountQueryOpts
}
