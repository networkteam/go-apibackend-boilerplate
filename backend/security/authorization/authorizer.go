package authorization

import (
	"myvendor.mytld/myproject/backend/security/authentication"
)

func NewAuthorizer(authCtx authentication.AuthContext) *Authorizer {
	return &Authorizer{
		authCtx: authCtx,
	}
}

type Authorizer struct {
	authCtx authentication.AuthContext
}

func (a *Authorizer) check(check authorizationCheck) error {
	return check(a.authCtx)
}
