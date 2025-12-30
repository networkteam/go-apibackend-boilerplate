package authentication

import (
	"context"
	"log/slog"
	"time"

	"github.com/gofrs/uuid"

	"myvendor.mytld/myproject/backend/domain/types"
)

type ctxKey string

const (
	authContextKey ctxKey = "authContext"
)

// AccessScope defines the scope of access for an authenticated session
type AccessScope string

const (
	// AccessScopeFull is the default access scope with full access using cookie-based authentication with CSRF
	AccessScopeFull AccessScope = "full"
)

func WithAuthContext(ctx context.Context, authCtx AuthContext) context.Context {
	return context.WithValue(ctx, authContextKey, authCtx)
}

// GetAuthContext gets stored authentication information (passed in by http middlewares) from context
func GetAuthContext(ctx context.Context) AuthContext {
	if authCtx, ok := ctx.Value(authContextKey).(AuthContext); ok {
		return authCtx
	}

	panic("no AuthContext given in context")
}

// GetAuthContextOptional gets stored authentication information (passed in by http middlewares) from context, or nil if not set
func GetAuthContextOptional(ctx context.Context) *AuthContext {
	if authCtx, ok := ctx.Value(authContextKey).(AuthContext); ok {
		return &authCtx
	}
	return nil
}

// AuthContext stores authentication information
type AuthContext struct {
	Authenticated             bool
	IgnoreAuthenticationState bool
	SkipCsrfCheck             bool
	Error                     error
	AuthSessionID             uuid.UUID
	AccessTokenID             uuid.UUID
	AccountID                 uuid.UUID
	OrganisationID            *uuid.UUID
	Role                      types.Role
	IssuedAt                  time.Time
	Expiry                    time.Time
	AccessScope               AccessScope
	ImpersonatorAuthSessionID *uuid.UUID
}

func (authCtx AuthContext) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.Bool("authenticated", authCtx.Authenticated),
		slog.String("role", string(authCtx.Role)),
		slog.Bool("ignoreAuthenticationState", authCtx.IgnoreAuthenticationState),
		slog.Any("authenticationError", authCtx.Error),
		slog.Bool("skipCsrfCheck", authCtx.SkipCsrfCheck),
		slog.String("accountID", authCtx.AccountID.String()),
		slog.String("accessScope", string(authCtx.AccessScope)),
	}
	if authCtx.OrganisationID != nil {
		attrs = append(attrs, slog.String("organisationID", authCtx.OrganisationID.String()))
	}
	return slog.GroupValue(attrs...)
}

// AuthContextWithError builds an auth context with an error
func AuthContextWithError(err error) AuthContext {
	return AuthContext{
		Error: err,
	}
}

func (authCtx AuthContext) GetTokenExpiryType() TokenExpiryType {
	if !authCtx.Authenticated {
		return TokenExpiryTypeDefault
	}

	expiry := authCtx.Expiry.Sub(authCtx.IssuedAt)

	if expiry >= TokenExpiryExtended {
		return TokenExpiryTypeExtended
	}

	if expiry >= TokenExpiryDefault {
		return TokenExpiryTypeDefault
	}

	return TokenExpiryTypeImpersonate
}
