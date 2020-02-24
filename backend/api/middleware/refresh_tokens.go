package middleware

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/zbyte/go-kallax"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/logger"
	"myvendor.mytld/myproject/backend/persistence/records"
	"myvendor.mytld/myproject/backend/security/authentication"
)

const (
	authTokenRefreshThreshold = 15 * time.Minute
)

func RefreshTokensMiddleware(db *sql.DB, timeSource domain.TimeSource, next http.Handler) http.Handler {
	accountStore := records.NewAccountStore(db)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := logger.GetLogger(ctx)

		authCtx := authentication.GetAuthContext(ctx)
		if authCtx.Authenticated {
			delta := timeSource.Now().Sub(authCtx.IssuedAt)
			if delta > authTokenRefreshThreshold {
				err := refreshTokens(w, r, authCtx, accountStore, timeSource)
				if err != nil {
					log.
						WithField("accountID", authCtx.AccountID).
						// err already has stacktrace
						WithError(err).
						Error("could not refresh tokens")
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func refreshTokens(w http.ResponseWriter, r *http.Request, authCtx authentication.AuthContext, accountStore *records.AccountStore, timeSource domain.TimeSource) error {
	account, err := accountStore.FindOne(records.NewAccountQuery().FindByID(kallax.UUID(authCtx.AccountID)))
	if err != nil {
		return errors.Wrap(err, "could not find account")
	}
	authToken, err := authentication.GenerateAuthToken(account, timeSource, authentication.TokenOpts{Expiry: authentication.AuthTokenExpiry})
	if err != nil {
		return errors.Wrap(err, "could not generate auth token")
	}
	csrfToken, err := authentication.GenerateCsrfToken(account, timeSource)
	if err != nil {
		return errors.Wrap(err, "could not generate CSRF token")
	}
	authentication.SetRefreshCsrfTokenHeader(w, csrfToken)
	authentication.SetAuthTokenCookie(w, r, authToken)
	return nil
}
