package handler

import (
	"context"
	"database/sql"

	"github.com/friendsofgo/errors"
	"github.com/networkteam/slogutils"

	"myvendor.mytld/myproject/backend/domain/command"
	"myvendor.mytld/myproject/backend/persistence/repository"
	"myvendor.mytld/myproject/backend/security/authentication"
	"myvendor.mytld/myproject/backend/security/authorization"
)

func (h *Handler) AuthSessionDeleteExpired(ctx context.Context, cmd command.AuthSessionDeleteExpiredCmd) error {
	logger := slogutils.FromContext(ctx).
		With(
			"component", "handler",
			"handler", "AuthSessionDeleteExpired",
		)
	logger.DebugContext(ctx, "Handling auth session delete expired command", "cmd", cmd)

	err := authorization.NewAuthorizer(authentication.GetAuthContext(ctx)).AllowsAuthSessionDeleteExpiredCmd()
	if err != nil {
		return err
	}

	var (
		deletedAuthAccessTokens int
		deletedAuthSessions     int
	)
	err = repository.Transactional(ctx, h.db, func(tx *sql.Tx) error {
		deletedAuthAccessTokens, err = repository.DeleteExpiredAuthAccessTokens(ctx, tx, h.timeSource)
		if err != nil {
			return errors.Wrap(err, "deleting expired auth access tokens")
		}

		deletedAuthSessions, err = repository.DeleteAuthSessionsWithoutAccessToken(ctx, tx)
		if err != nil {
			return errors.Wrap(err, "deleting auth sessions without access token")
		}

		return nil
	})
	if err != nil {
		return err
	}

	logger.InfoContext(ctx, "Deleted expired auth sessions and access tokens", "deletedAuthAccessTokens", deletedAuthAccessTokens, "deletedAuthSessions", deletedAuthSessions)

	return nil
}
