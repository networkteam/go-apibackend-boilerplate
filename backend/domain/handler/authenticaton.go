package handler

import (
	"context"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/gofrs/uuid"

	api_types "myvendor.mytld/myproject/backend/api/types"
	"myvendor.mytld/myproject/backend/domain/model"
	"myvendor.mytld/myproject/backend/security/authentication"
)

func (h *Handler) generateAuthSession(accountID uuid.UUID, deviceInfo model.DeviceInfo) (model.AuthSession, error) {
	authSessionID, err := uuid.NewV7()
	if err != nil {
		return model.AuthSession{}, errors.Wrap(err, "generating auth session ID")
	}

	return model.AuthSession{
		ID:         authSessionID,
		AccountID:  accountID,
		DeviceInfo: &deviceInfo,
	}, nil
}

func (h *Handler) generateAccessToken(authSessionID uuid.UUID, expiresAt time.Time) (model.AuthAccessToken, error) {
	accessTokenID, err := uuid.NewV4()
	if err != nil {
		return model.AuthAccessToken{}, errors.Wrap(err, "generating access token ID")
	}

	return model.AuthAccessToken{
		ID:            accessTokenID,
		AuthSessionID: authSessionID,
		ExpiresAt:     expiresAt,
	}, nil
}

func (h *Handler) setAccessTokenCookieForAccount(ctx context.Context, account authentication.AccessTokenDataProvider, accessTokenID uuid.UUID, tokenExpiryType authentication.TokenExpiryType) error {
	tokenOpts := authentication.TokenOptsForAccount(accessTokenID, tokenExpiryType)

	accessToken, err := authentication.GenerateAccessToken(h.config, account, h.timeSource, tokenOpts)
	if err != nil {
		return errors.Wrap(err, "generating access token")
	}

	csrfToken, err := authentication.GenerateCsrfToken(h.config, h.timeSource, tokenOpts)
	if err != nil {
		return errors.Wrap(err, "generating CSRF token")
	}

	r := api_types.GetHTTPRequest(ctx)
	w := api_types.GetHTTPResponse(ctx)

	authentication.SetAccessTokenCookie(w, r, accessToken)
	authentication.SetCsrfTokenCookie(w, r, csrfToken)

	return nil
}
