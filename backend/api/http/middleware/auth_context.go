package middleware

import (
	"context"
	std_errors "errors"
	"log/slog"
	"net/http"

	"github.com/friendsofgo/errors"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/gofrs/uuid"
	"github.com/networkteam/slogutils"

	"myvendor.mytld/myproject/backend/api"
	api_types "myvendor.mytld/myproject/backend/api/types"
	"myvendor.mytld/myproject/backend/domain/query"
	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/persistence/repository"
	"myvendor.mytld/myproject/backend/security/authentication"
)

// AuthContextMiddleware sets an auth context from a HTTP request
// considering auth token and CSRF token.
//
// Dependencies are given in deps and need to contain Config, DB and TimeSource.
func AuthContextMiddleware(deps api.ResolverDependencies, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := slogutils.FromContext(ctx)

		var authCtx authentication.AuthContext
		if accessToken := api_types.GetAccessToken(ctx); accessToken != "" {
			authCtx = authCtxFromToken(ctx, deps, accessToken)
			authCtx.SkipCsrfCheck = api_types.GetSkipCsrfCheck(ctx)
			if authCtx.Error == nil && !authCtx.SkipCsrfCheck {
				csrfToken := api_types.GetCsrfToken(ctx)
				if err := checkCsrfToken(ctx, deps, authCtx, csrfToken); err != nil {
					authCtx = authentication.AuthContextWithError(err)
				}
			}
		}
		ctx = authentication.WithAuthContext(ctx, authCtx)

		// Add some additional logging information if authenticated
		if authCtx.Authenticated {
			logger = logger.
				With(slog.Group("auth", "accountID", authCtx.AccountID, "role", authCtx.Role))
			ctx = slogutils.WithLogger(ctx, logger)
		}

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func authCtxFromToken(ctx context.Context, deps api.ResolverDependencies, accessTokenValue string) (authCtx authentication.AuthContext) {
	logger := slogutils.FromContext(ctx)

	token, err := jwt.ParseSigned(accessTokenValue, []jose.SignatureAlgorithm{jose.HS256})
	if err != nil {
		logger.WarnContext(ctx, "Could not parse signed access token", slogutils.Err(errors.WithStack(err)))
		return authentication.AuthContextWithError(api.ErrAccessTokenInvalid)
	}

	var verifiedClaims authentication.AccessTokenClaims
	if err := token.Claims([]byte(deps.Config.JWTSecret), &verifiedClaims); err != nil {
		var unverfiedClaims jwt.Claims
		claimsErr := token.UnsafeClaimsWithoutVerification(&unverfiedClaims)
		if claimsErr != nil {
			err = std_errors.Join(err, claimsErr)
		}

		logger.WarnContext(ctx, "Could not verify claims in access token", slog.Group("auth", "accessTokenID", unverfiedClaims.ID), slogutils.Err(errors.WithStack(err)))
		return authentication.AuthContextWithError(api.ErrAccessTokenInvalid)
	}

	accessTokenID, err := uuid.FromString(verifiedClaims.ID)
	if err != nil {
		logger.WarnContext(ctx, "Invalid access token ID in access token ID claim", slog.Group("auth", "accessTokenID", verifiedClaims.ID), slogutils.Err(errors.WithStack(err)))
		return authentication.AuthContextWithError(api.ErrAccessTokenInvalid)
	}

	accessToken, err := repository.FindAuthAccessTokenByID(ctx, deps.DB, accessTokenID, query.AuthAccessTokenQueryOpts{
		IncludeAuthSession: true,
		// We side-load the account with a possible user and all permissions
		AuthSessionQueryOpts: &query.AuthSessionQueryOpts{
			IncludeAccount: true,
		},
	})
	if errors.Is(err, repository.ErrNotFound) {
		logger.WarnContext(ctx, "Auth access token not found", slog.Group("auth", "accessTokenID", accessTokenID))
		return authentication.AuthContextWithError(api.ErrAccessTokenInvalid)
	} else if err != nil {
		logger.ErrorContext(ctx, "Failed to find auth access token", slog.Group("auth", "accessTokenID", accessTokenID), slogutils.Err(errors.WithStack(err)))
		return authentication.AuthContextWithError(api.ErrAuthenticationFailed)
	}
	if accessToken.AuthSession == nil {
		panic("Auth access token without auth session, this should never happen")
	}

	err = verifiedClaims.Validate(jwt.Expected{}.WithTime(deps.TimeSource.Now()))
	if err != nil {
		logger.WarnContext(ctx, "Could not validate claims in access token", slog.Group("auth", "accessTokenID", accessTokenID), slogutils.Err(errors.WithStack(err)))
		if errors.Is(err, jwt.ErrExpired) {
			return authentication.AuthContextWithError(api.ErrAccessTokenExpired)
		}
		return authentication.AuthContextWithError(api.ErrAccessTokenInvalid)
	}

	authCtx.Role = types.Role(verifiedClaims.Role)
	if !authCtx.Role.IsValid() {
		logger.WarnContext(ctx, "Invalid role in access token", slog.Group("auth", "accessTokenID", accessTokenID, "role", verifiedClaims.Role))
		return authentication.AuthContextWithError(api.ErrAccessTokenInvalid)
	}

	authCtx.Authenticated = true
	authCtx.AccessTokenID = accessToken.ID
	authCtx.AuthSessionID = accessToken.AuthSession.ID
	authCtx.AccountID = accessToken.AuthSession.AccountID
	authCtx.AccessScope = authentication.AccessScopeFull // Normal cookie-based auth has full scope
	if accessToken.AuthSession.Account != nil && accessToken.AuthSession.Account.OrganisationID.Valid {
		authCtx.OrganisationID = &accessToken.AuthSession.Account.OrganisationID.UUID
	}

	if verifiedClaims.IssuedAt != nil {
		authCtx.IssuedAt = verifiedClaims.IssuedAt.Time()
	}
	if verifiedClaims.Expiry != nil {
		authCtx.Expiry = verifiedClaims.Expiry.Time()
	}

	return authCtx
}

func checkCsrfToken(ctx context.Context, deps api.ResolverDependencies, authCtx authentication.AuthContext, csrfTokenValue string) error {
	logger := slogutils.FromContext(ctx)

	if csrfTokenValue == "" {
		return api.ErrCsrfTokenMissing
	}

	csrfToken, err := jwt.ParseSigned(csrfTokenValue, []jose.SignatureAlgorithm{jose.HS256})
	if err != nil {
		logger.WarnContext(ctx, "Could not parse signed CSRF token", slogutils.Err(errors.WithStack(err)))
		return api.ErrCsrfTokenInvalid
	}

	var verifiedClaims jwt.Claims
	if err := csrfToken.Claims([]byte(deps.Config.JWTSecret), &verifiedClaims); err != nil {
		logger.WarnContext(ctx, "Could not verify claims in CSRF token", slog.Group("auth", "accountID", authCtx.AccountID), slogutils.Err(errors.WithStack(err)))
		return api.ErrCsrfTokenInvalid
	}

	err = verifiedClaims.Validate(jwt.Expected{
		ID: authCtx.AuthSessionID.String(),
	}.WithTime(deps.TimeSource.Now()))
	if err != nil {
		logger.WarnContext(ctx, "Could not validate claims in CSRF token", slog.Group("auth", "accountID", authCtx.AccountID), slogutils.Err(errors.WithStack(err)))
		if errors.Is(err, jwt.ErrExpired) {
			return api.ErrCsrfTokenExpired
		}
		return api.ErrCsrfTokenInvalid
	}

	return nil
}
