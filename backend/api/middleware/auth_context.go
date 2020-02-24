package middleware

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/friendsofgo/errors"
	"github.com/gofrs/uuid"
	"github.com/zbyte/go-kallax"
	"gopkg.in/square/go-jose.v2/jwt"

	"myvendor.mytld/myproject/backend/api"
	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/logger"
	"myvendor.mytld/myproject/backend/persistence/records"
	"myvendor.mytld/myproject/backend/security/authentication"
)

// AuthContextMiddleware sets an auth context from a HTTP request
// considering auth token and CSRF token
func AuthContextMiddleware(db *sql.DB, timeSource domain.TimeSource, next http.Handler) http.Handler {
	accountStore := records.NewAccountStore(db)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var authCtx authentication.AuthContext
		if authToken := api.GetAuthToken(ctx); authToken != "" {
			authCtx = authCtxFromToken(ctx, accountStore, authToken, timeSource)
			authCtx.SkipCsrfCheck = api.GetSkipCsrfCheck(ctx)
			if authCtx.Error == nil && !authCtx.SkipCsrfCheck {
				csrfToken := api.GetCsrfToken(ctx)
				if err := checkCsrfToken(ctx, authCtx, csrfToken, timeSource); err != nil {
					authCtx = authentication.AuthContextWithError(err)
				}
			}
		}
		ctx = authentication.WithAuthContext(ctx, authCtx)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func authCtxFromToken(ctx context.Context, accountStore *records.AccountStore, authTokenValue string, timeSource domain.TimeSource) (authCtx authentication.AuthContext) {
	log := logger.GetLogger(ctx)

	authToken, err := jwt.ParseSigned(authTokenValue)
	if err != nil {
		log.
			WithError(errors.WithStack(err)).
			Warn("could not parse signed auth token")
		return authentication.AuthContextWithError(api.ErrAuthTokenInvalid)
	}
	var unverifiedClaims jwt.Claims
	if err := authToken.UnsafeClaimsWithoutVerification(&unverifiedClaims); err != nil {
		log.
			WithError(errors.WithStack(err)).
			Warn("could not get claims from auth token")
		return authentication.AuthContextWithError(api.ErrAuthTokenInvalid)
	}
	accountID, err := uuid.FromString(unverifiedClaims.Subject)
	if err != nil {
		log.
			WithError(errors.WithStack(err)).
			WithField("subject", unverifiedClaims.Subject).
			Warn("could not get account ID from subject claim in auth token")
		return authentication.AuthContextWithError(api.ErrAuthTokenInvalid)
	}

	account, err := accountStore.FindOne(records.NewAccountQuery().FindByID(kallax.UUID(accountID)))
	if err != nil {
		log.
			WithError(errors.WithStack(err)).
			WithField("accountID", accountID).
			Warn("could not find account for subject claim in auth token")
		return authentication.AuthContextWithError(api.ErrAuthTokenInvalid)
	}

	var verifiedClaims jwt.Claims
	if err := authToken.Claims([]byte(account.Secret), &verifiedClaims); err != nil {
		log.
			WithError(errors.WithStack(err)).
			WithField("accountID", accountID).
			Warn("could not verify claims in auth token")
		return authentication.AuthContextWithError(api.ErrAuthTokenInvalid)
	}

	err = verifiedClaims.Validate(jwt.Expected{}.WithTime(timeSource.Now()))
	if err != nil {
		log.
			WithError(errors.WithStack(err)).
			WithField("accountID", accountID).
			Warn("could not validate claims in auth token")
		if err == jwt.ErrExpired {
			return authentication.AuthContextWithError(api.ErrAuthTokenExpired)
		}
		return authentication.AuthContextWithError(api.ErrAuthTokenInvalid)
	}

	authCtx.Authenticated = true
	authCtx.AccountID = accountID
	if organisationID := account.GetOrganisationID(); organisationID != uuid.Nil {
		authCtx.OrganisationID = &organisationID
	}
	authCtx.IssuedAt = verifiedClaims.IssuedAt.Time()
	authCtx.Secret = []byte(account.Secret)
	authCtx.Role, err = account.Role()
	if err != nil {
		log.WithError(errors.WithStack(err)).WithField("accountID", accountID).Error("invalid role for account")
		return authentication.AuthContextWithError(api.ErrAuthTokenInvalid)
	}

	return
}

func checkCsrfToken(ctx context.Context, authCtx authentication.AuthContext, csrfTokenValue string, timeSource domain.TimeSource) error {
	log := logger.GetLogger(ctx)

	if csrfTokenValue == "" {
		return api.ErrCsrfTokenMissing
	}

	csrfToken, err := jwt.ParseSigned(csrfTokenValue)
	if err != nil {
		log.
			WithError(errors.WithStack(err)).
			Warn("could not parse signed CSRF token")
		return api.ErrCsrfTokenInvalid
	}

	var verifiedClaims jwt.Claims
	if err := csrfToken.Claims(authCtx.Secret, &verifiedClaims); err != nil {
		log.
			WithError(errors.WithStack(err)).
			WithField("accountID", authCtx.AccountID).
			Warn("could not verify claims in CSRF token")
		return api.ErrCsrfTokenInvalid
	}

	err = verifiedClaims.Validate(jwt.Expected{}.WithTime(timeSource.Now()))
	if err != nil {
		log.
			WithError(errors.WithStack(err)).
			WithField("accountID", authCtx.AccountID).
			Warn("could not validate claims in CSRF token")
		return api.ErrCsrfTokenInvalid
	}

	return nil
}
