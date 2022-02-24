package authentication

import (
	"github.com/friendsofgo/errors"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"

	"myvendor.mytld/myproject/backend/domain"
)

func GenerateCsrfToken(account TokenSecretProvider, timeSource domain.TimeSource, opts TokenOpts) (string, error) {
	key := account.GetTokenSecret()
	sig, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.HS256, Key: key}, (&jose.SignerOptions{}).WithType("JWT"))
	if err != nil {
		return "", errors.Wrap(err, "creating signer for JWT")
	}

	now := timeSource.Now()

	cl := jwt.Claims{
		Expiry: jwt.NewNumericDate(now.Add(opts.Expiry)),
	}
	raw, err := jwt.Signed(sig).Claims(cl).CompactSerialize()
	if err != nil {
		return "", errors.Wrap(err, "signing and serializing JWT")
	}

	return raw, nil
}
