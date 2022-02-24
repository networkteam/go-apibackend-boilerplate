package middleware

import (
	"database/sql"
	"net/http"
	"time"

	logger "github.com/apex/log"
	"github.com/friendsofgo/errors"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/persistence/repository"
	"myvendor.mytld/myproject/backend/security/authentication"
)

const (
	AuthTokenRefreshThreshold = 15 * time.Minute
)

func RefreshTokensMiddleware(db *sql.DB, timeSource domain.TimeSource, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := logger.FromContext(ctx)

		authCtx := authentication.GetAuthContext(ctx)
		if authCtx.Authenticated {
			delta := timeSource.Now().Sub(authCtx.IssuedAt)
			if delta > AuthTokenRefreshThreshold {
				err := refreshTokens(w, r, authCtx, db, timeSource)
				if err != nil {
					log.
						// err already has stacktrace
						WithError(err).
						Error("could not refresh tokens")
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func refreshTokens(w http.ResponseWriter, r *http.Request, authCtx authentication.AuthContext, db *sql.DB, timeSource domain.TimeSource) error {
	account, err := repository.FindAccountByID(r.Context(), db, authCtx.AccountID)

	if err != nil {
		return errors.Wrap(err, "could not find account")
	}
	tokenOpts := authentication.TokenOptsForAccount(account)
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
