package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/networkteam/slogutils"

	"myvendor.mytld/myproject/backend/api"
	api_types "myvendor.mytld/myproject/backend/api/types"
	"myvendor.mytld/myproject/backend/domain/command"
	"myvendor.mytld/myproject/backend/security/authentication"
)

const (
	AccessTokenRefreshThreshold = 15 * time.Minute
)

func RefreshTokensMiddleware(deps api.ResolverDependencies, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := slogutils.FromContext(ctx)

		authCtx := authentication.GetAuthContext(ctx)
		if authCtx.Authenticated {
			delta := deps.TimeSource.Now().Sub(authCtx.IssuedAt)
			if delta > AccessTokenRefreshThreshold {
				err := refreshTokens(ctx, w, r, deps, authCtx)
				if err != nil {
					logger.ErrorContext(ctx, "Could not refresh tokens", slogutils.Err(err))
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func refreshTokens(ctx context.Context, w http.ResponseWriter, r *http.Request, deps api.ResolverDependencies, authCtx authentication.AuthContext) error {
	// Make sure we have request and response in context (do not depend on RequestAndResponseWriterMiddleware to be present)
	ctx = api_types.WithHTTPRequest(ctx, r)
	ctx = api_types.WithHTTPResponse(ctx, w)

	tokenExpiryType := authCtx.GetTokenExpiryType()
	expiresAt := tokenExpiryType.GetExpiresAt(deps.TimeSource)

	cmd, err := command.NewAuthSessionRefreshCmd(authCtx.AuthSessionID, expiresAt)
	if err != nil {
		return errors.Wrap(err, "creating auth access token create command")
	}

	err = deps.Handler().AuthSessionRefresh(ctx, cmd)
	if err != nil {
		return errors.Wrap(err, "refreshing auth session")
	}

	return nil
}
