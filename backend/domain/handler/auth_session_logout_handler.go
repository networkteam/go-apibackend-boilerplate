package handler

import (
	"context"

	"github.com/friendsofgo/errors"
	"github.com/networkteam/slogutils"

	api_types "myvendor.mytld/myproject/backend/api/types"
	"myvendor.mytld/myproject/backend/domain/command"
	"myvendor.mytld/myproject/backend/domain/query"
	"myvendor.mytld/myproject/backend/persistence/repository"
	"myvendor.mytld/myproject/backend/security/authentication"
)

func (h *Handler) AuthSessionLogout(ctx context.Context, cmd command.AuthSessionLogoutCmd) error {
	logger := slogutils.FromContext(ctx).
		With(
			"component", "handler",
			"handler", "AuthSessionLogout",
		)
	logger.DebugContext(ctx, "Handling auth session logout command")

	err := cmd.Validate()
	if err != nil {
		return err
	}

	w := api_types.GetHTTPResponse(ctx)
	req := api_types.GetHTTPRequest(ctx)

	authCtx := authentication.GetAuthContext(ctx)

	authentication.DeleteAccessTokenCookie(w, req)
	authentication.DeleteCsrfTokenCookie(w, req)

	accessToken, err := repository.FindAuthAccessTokenByID(ctx, h.db, authCtx.AccessTokenID, query.AuthAccessTokenQueryOpts{
		IncludeAuthSession: true,
	})
	if err != nil {
		return errors.Wrap(err, "finding auth access token")
	}

	if accessToken.AuthSession == nil {
		return errors.New("access token without auth session")
	}

	err = repository.DeleteAuthSession(ctx, h.db, accessToken.AuthSession.ID)
	if err != nil {
		return errors.Wrap(err, "deleting auth session")
	}

	logger.InfoContext(ctx,
		"Logged out auth session",
		"authSessionID", accessToken.AuthSession.ID,
	)

	return nil
}
