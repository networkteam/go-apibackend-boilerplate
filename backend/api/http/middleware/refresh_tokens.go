package middleware

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/networkteam/slogutils"

	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/persistence/repository"
	"myvendor.mytld/myproject/backend/security/authentication"
)

const (
	AuthTokenRefreshThreshold = 15 * time.Minute
)

func RefreshTokensMiddleware(db *sql.DB, timeSource types.TimeSource, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := slogutils.FromContext(ctx)

		authCtx := authentication.GetAuthContext(ctx)
		if authCtx.Authenticated {
			delta := timeSource.Now().Sub(authCtx.IssuedAt)
			if delta > AuthTokenRefreshThreshold {
				err := refreshTokens(w, r, authCtx, db, timeSource)
				if err != nil {
					// err already has stacktrace
					logger.ErrorContext(ctx, "Could not refresh tokens", slogutils.Err(err))
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func refreshTokens(w http.ResponseWriter, r *http.Request, authCtx authentication.AuthContext, db *sql.DB, timeSource types.TimeSource) error {
	account, err := repository.FindAccountByID(r.Context(), db, authCtx.AccountID, nil)
	if err != nil {
		return errors.Wrap(err, "could not find account")
	}

	tokenOpts := authentication.TokenOptsForAccount(account, authCtx.HasExtendedExpiry())
	authToken, err := authentication.GenerateAuthToken(account, timeSource, tokenOpts)
	if err != nil {
		return errors.Wrap(err, "could not generate auth token")
	}

	csrfToken, err := authentication.GenerateCsrfToken(account, timeSource, tokenOpts)
	if err != nil {
		return errors.Wrap(err, "could not generate CSRF token")
	}

	authentication.SetRefreshCsrfTokenHeader(w, csrfToken)
	authentication.SetRefreshAuthTokenHeader(w, authToken)
	authentication.SetAuthTokenCookie(w, r, authToken)

	return nil
}
