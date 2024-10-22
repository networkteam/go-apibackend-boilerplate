package authentication

import (
	"github.com/friendsofgo/errors"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"

	"myvendor.mytld/myproject/backend/domain/types"
)

func GenerateCsrfToken(account TokenSecretProvider, timeSource types.TimeSource, opts TokenOpts) (string, error) {
	key := account.GetTokenSecret()
	sig, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.HS256, Key: key}, (&jose.SignerOptions{}).WithType("JWT"))
	if err != nil {
		return "", errors.Wrap(err, "creating signer for JWT")
	}

	now := timeSource.Now()

	cl := jwt.Claims{
		Expiry: jwt.NewNumericDate(now.Add(opts.Expiry)),
	}
	raw, err := jwt.Signed(sig).Claims(cl).Serialize()
	if err != nil {
		return "", errors.Wrap(err, "signing and serializing JWT")
	}

	return raw, nil
}
