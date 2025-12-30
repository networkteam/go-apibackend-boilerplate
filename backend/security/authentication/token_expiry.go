package authentication

import (
	"errors"
	"time"

	"myvendor.mytld/myproject/backend/domain/types"
)

const (
	TokenExpiryDefault     = 6 * time.Hour
	TokenExpiryExtended    = 30 * 24 * time.Hour
	TokenExpiryImpersonate = 15 * time.Minute
)

type TokenExpiryType string

const TokenExpiryTypeDefault = TokenExpiryType("Default")
const TokenExpiryTypeExtended = TokenExpiryType("Extended")
const TokenExpiryTypeImpersonate = TokenExpiryType("Impersonate")

var ErrUnknownTokenTypeExpiry = errors.New("unknown token expiry")

func TokenExpiryTypeByIdentifier(tokenExpiryTypeIdentifier string) (TokenExpiryType, error) {
	r := TokenExpiryType(tokenExpiryTypeIdentifier)
	if !r.IsValid() {
		return r, ErrUnknownTokenTypeExpiry
	}
	return r, nil
}

func (r TokenExpiryType) IsValid() bool {
	switch r {
	case TokenExpiryTypeDefault:
	case TokenExpiryTypeExtended:
	case TokenExpiryTypeImpersonate:
	default:
		return false
	}
	return true
}

func (r TokenExpiryType) GetExpiry() time.Duration {
	switch r {
	case TokenExpiryTypeDefault:
		return TokenExpiryDefault
	case TokenExpiryTypeExtended:
		return TokenExpiryExtended
	case TokenExpiryTypeImpersonate:
		return TokenExpiryImpersonate
	default:
		return TokenExpiryDefault
	}
}

func (r TokenExpiryType) GetExpiresAt(timeSource types.TimeSource) time.Time {
	expiry := r.GetExpiry()

	return timeSource.Now().Add(expiry)
}
