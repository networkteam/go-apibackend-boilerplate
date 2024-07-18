package helper

import (
	"context"

	fog_errors "github.com/friendsofgo/errors"

	"myvendor.mytld/myproject/backend/api"
	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/security/authentication"
)

func SetAuthTokenCookieForAccount(ctx context.Context, account authentication.AuthTokenDataProvider, timeSource domain.TimeSource, extendedExpiry bool) (csrfToken string, err error) {
	tokenOpts := authentication.TokenOptsForAccount(account, extendedExpiry)
	authToken, err := authentication.GenerateAuthToken(account, timeSource, tokenOpts)
	if err != nil {
		return "", fog_errors.Wrap(err, "generating auth token")
	}

	csrfToken, err = authentication.GenerateCsrfToken(account, timeSource, tokenOpts)
	if err != nil {
		return "", fog_errors.Wrap(err, "generating CSRF token")
	}

	req := api.GetHTTPRequest(ctx)
	w := api.GetHTTPResponse(ctx)
	authentication.SetAuthTokenCookie(w, req, authToken)

	return csrfToken, nil
}
