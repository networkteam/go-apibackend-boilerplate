package authentication

import (
	"time"

	"github.com/friendsofgo/errors"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"

	"myvendor.mytld/myproject/backend/domain"
)

const AuthTokenExpiryDefault = 6 * time.Hour

type TokenOpts struct {
	Expiry time.Duration
}

// TokenOptsForAccount will return the token options (expiry) based on the role of an account
func TokenOptsForAccount(account RoleIdentifierProvider) TokenOpts {
	expiry := AuthTokenExpiryDefault

	return TokenOpts{
		Expiry: expiry,
	}
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

	organisationIDValue := ""
	if account.GetOrganisationID().Valid {
		organisationIDValue = account.GetOrganisationID().UUID.String()
	}
	privateCl := struct {
		Role           string `json:"role"`
		OrganisationID string `json:"organisationId,omitempty"`
	}{
		Role:           account.GetRoleIdentifier(),
		OrganisationID: organisationIDValue,
	}

	raw, err := jwt.Signed(sig).Claims(cl).Claims(privateCl).CompactSerialize()
	if err != nil {
		return "", errors.Wrap(err, "signing and serializing JWT")
	}

	return raw, nil
}
