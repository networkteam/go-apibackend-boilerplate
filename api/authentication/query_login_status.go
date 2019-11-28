package authentication

import (
	"context"

	"github.com/apex/log"

	"myvendor/myproject/backend/security/authentication"
)

func (r *QueryResolver) LoginStatus(ctx context.Context) (bool, error) {
	log.Debug("Querying login status")

	authCtx := authentication.GetAuthContext(ctx)
	return authCtx.Authenticated, nil
}
