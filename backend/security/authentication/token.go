package authentication

import (
	"time"

	"github.com/friendsofgo/errors"
	"github.com/go-jose/go-jose/v4"
	"github.com/gofrs/uuid"

	"myvendor.mytld/myproject/backend/domain"
)

type TokenOpts struct {
	Expiry        time.Duration
	AccessTokenID uuid.UUID
	AuthSessionID uuid.UUID
}

// TokenOptsForAccount will return the token options (expiry) based on the role of an account
func TokenOptsForAccount(accessTokenID uuid.UUID, authSessionID uuid.UUID, tokenExpiryType TokenExpiryType) TokenOpts {
	expiry := tokenExpiryType.GetExpiry()

	return TokenOpts{
		Expiry:        expiry,
		AccessTokenID: accessTokenID,
		AuthSessionID: authSessionID,
	}
}

func getSigner(config domain.Config) (jose.Signer, error) {
	sig, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.HS256, Key: []byte(config.JWTSecret)}, (&jose.SignerOptions{}).WithType("JWT"))
	if err != nil {
		return nil, errors.Wrap(err, "creating signer for JWT")
	}
	return sig, nil
}
