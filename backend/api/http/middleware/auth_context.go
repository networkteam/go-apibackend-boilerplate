package middleware

import (
	"context"
	"database/sql"
	"net/http"

	logger "github.com/apex/log"
	"github.com/friendsofgo/errors"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/gofrs/uuid"

	"myvendor.mytld/myproject/backend/api"
	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/persistence/repository"
	"myvendor.mytld/myproject/backend/security/authentication"
)

// AuthContextMiddleware sets an auth context from a HTTP request
// considering auth token and CSRF token
func AuthContextMiddleware(db *sql.DB, timeSource domain.TimeSource, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var authCtx authentication.AuthContext
		if authToken := api.GetAuthToken(ctx); authToken != "" {
			authCtx = authCtxFromToken(ctx, db, authToken, timeSource)
			authCtx.SkipCsrfCheck = api.GetSkipCsrfCheck(ctx)
			if authCtx.Error == nil && !authCtx.SkipCsrfCheck {
				csrfToken := api.GetCsrfToken(ctx)
				if err := checkCsrfToken(ctx, authCtx, csrfToken, timeSource); err != nil {
					authCtx = authentication.AuthContextWithError(err)
				}
			}
		}
		ctx = authentication.WithAuthContext(ctx, authCtx)

		// Add some additional logging information if authenticated
		if authCtx.Authenticated {
			log := logger.FromContext(ctx)
			log = log.
				WithField("authAccountID", authCtx.AccountID).
				WithField("authRole", authCtx.Role)
			ctx = logger.NewContext(ctx, log)
		}

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func authCtxFromToken(ctx context.Context, db *sql.DB, authTokenValue string, timeSource domain.TimeSource) (authCtx authentication.AuthContext) {
	log := logger.FromContext(ctx)

	authToken, err := jwt.ParseSigned(authTokenValue, []jose.SignatureAlgorithm{jose.HS256})
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

	account, err := repository.FindAccountByID(ctx, db, accountID, domain.AccountQueryOpts{})
	if err != nil {
		log.
			WithError(errors.WithStack(err)).
			WithField("accountID", accountID).
			Warn("could not find account for subject claim in auth token")
		return authentication.AuthContextWithError(api.ErrAuthTokenInvalid)
	}

	var verifiedClaims jwt.Claims
	if err := authToken.Claims(account.Secret, &verifiedClaims); err != nil {
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
		if errors.Is(err, jwt.ErrExpired) {
			return authentication.AuthContextWithError(api.ErrAuthTokenExpired)
		}
		return authentication.AuthContextWithError(api.ErrAuthTokenInvalid)
	}

	authCtx.Authenticated = true
	authCtx.AccountID = accountID
	if account.OrganisationID.Valid {
		authCtx.OrganisationID = &account.OrganisationID.UUID
	}
	if verifiedClaims.IssuedAt != nil {
		authCtx.IssuedAt = verifiedClaims.IssuedAt.Time()
	}
	if verifiedClaims.Expiry != nil {
		authCtx.Expiry = verifiedClaims.Expiry.Time()
	}
	authCtx.Secret = account.Secret
	authCtx.Role = account.Role
	if !authCtx.Role.IsValid() {
		log.
			WithField("accountID", accountID).
			Errorf("Invalid role in account: %q", account.Role)
		return authentication.AuthContextWithError(api.ErrAuthTokenInvalid)
	}

	return authCtx
}

func checkCsrfToken(ctx context.Context, authCtx authentication.AuthContext, csrfTokenValue string, timeSource domain.TimeSource) error {
	log := logger.FromContext(ctx)

	if csrfTokenValue == "" {
		return api.ErrCsrfTokenMissing
	}

	csrfToken, err := jwt.ParseSigned(csrfTokenValue, []jose.SignatureAlgorithm{jose.HS256})
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
