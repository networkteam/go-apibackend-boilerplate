package authentication

import (
	"time"

	"github.com/friendsofgo/errors"
	jose "gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"

	"myvendor.mytld/myproject/backend/domain"
)

const AuthTokenExpiry = 6 * time.Hour

type TokenOpts struct {
	Expiry time.Duration
}

func GenerateAuthToken(account AuthTokenDataProvider, timeSource domain.TimeSource, opts TokenOpts) (string, error) {
	key := account.GetTokenSecret()
	sig, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.HS256, Key: key}, (&jose.SignerOptions{}).WithType("JWT"))
	if err != nil {
		return "", errors.Wrap(err, "creating signer for JWT")
	}

	now := timeSource.Now()

	cl := jwt.Claims{
		Subject:  account.GetAccountID().String(),
		IssuedAt: jwt.NewNumericDate(now),
		Expiry:   jwt.NewNumericDate(now.Add(opts.Expiry)),
	}

	privateCl := struct {
		Role           string `json:"role"`
		OrganisationID string `json:"organisationId"`
	}{
		Role:           account.GetRoleIdentifier(),
		OrganisationID: account.GetOrganisationID().String(),
	}

	raw, err := jwt.Signed(sig).Claims(cl).Claims(privateCl).CompactSerialize()
	if err != nil {
		return "", errors.Wrap(err, "signing and serializing JWT")
	}

	return raw, nil
}
