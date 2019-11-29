package authentication

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid"

	"myvendor.mytld/myproject/backend/domain"
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
	Role                      domain.Role
	Secret                    []byte
	IssuedAt                  time.Time
}

func (authCtx AuthContext) String() string {
	return fmt.Sprintf("%+v", struct {
		Authenticated             bool
		IgnoreAuthenticationState bool
		SkipCsrfCheck             bool
		Error                     error
		AccountID                 uuid.UUID
		OrganisationID            *uuid.UUID
		Role                      domain.Role
		IssuedAt                  time.Time
	}{
		Authenticated:             authCtx.Authenticated,
		IgnoreAuthenticationState: authCtx.IgnoreAuthenticationState,
		SkipCsrfCheck:             authCtx.SkipCsrfCheck,
		Error:                     authCtx.Error,
		AccountID:                 authCtx.AccountID,
		OrganisationID:            authCtx.OrganisationID,
		Role:                      authCtx.Role,
		IssuedAt:                  authCtx.IssuedAt,
	})
}

// AuthContextWithError builds an auth context with an error
func AuthContextWithError(err error) AuthContext {
	return AuthContext{
		Error: err,
	}
}
