package authentication

import (
	"context"

	"myvendor.mytld/myproject/backend/logger"
	"myvendor.mytld/myproject/backend/security/authentication"
)

func (r *QueryResolver) LoginStatus(ctx context.Context) (bool, error) {
	log := logger.GetLogger(ctx).
		WithField("query", "loginStatus")

	authCtx := authentication.GetAuthContext(ctx)

	log.
		Debug("Querying login status")

	return authCtx.Authenticated, nil
}
