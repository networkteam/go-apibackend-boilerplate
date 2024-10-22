package authentication

import (
	"context"
	"time"

	"github.com/apex/log"
	"github.com/gofrs/uuid"

	"myvendor.mytld/myproject/backend/domain/types"
)

type ctxKey string

const (
	authContextKey ctxKey = "authContext"
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

// AuthContext stores authentication information
type AuthContext struct {
	Authenticated             bool
	IgnoreAuthenticationState bool
	SkipCsrfCheck             bool
	Error                     error
	AccountID                 uuid.UUID
	OrganisationID            *uuid.UUID
	Role                      types.Role
	Secret                    []byte
	IssuedAt                  time.Time
	Expiry                    time.Time
}

func (authCtx AuthContext) Fields() log.Fields {
	return map[string]any{
		"authenticated":             authCtx.Authenticated,
		"role":                      authCtx.Role,
		"ignoreAuthenticationState": authCtx.IgnoreAuthenticationState,
		"authenticationError":       authCtx.Error,
		"skipCsrfCheck":             authCtx.SkipCsrfCheck,
		"accountID":                 authCtx.AccountID,
		"organisationID":            authCtx.OrganisationID,
	}
}

// AuthContextWithError builds an auth context with an error
func AuthContextWithError(err error) AuthContext {
	return AuthContext{
		Error: err,
	}
}

func (authCtx AuthContext) OrganisationIDorNil() uuid.UUID {
	if authCtx.OrganisationID == nil {
		return uuid.Nil
	}

	return *authCtx.OrganisationID
}

func (authCtx AuthContext) HasExtendedExpiry() bool {
	if !authCtx.Authenticated {
		return false
	}

	return authCtx.Expiry.Sub(authCtx.IssuedAt) >= AuthTokenExpiryExtended
}

func (authCtx AuthContext) IsOrganisation() bool {
	return authCtx.OrganisationID != nil
}
