package authentication

import (
	"github.com/friendsofgo/errors"
	"github.com/go-jose/go-jose/v4/jwt"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/domain/types"
)

type AccessTokenClaims struct {
	jwt.Claims
	Role string `json:"role"`
}

func GenerateAccessToken(config domain.Config, account AccessTokenDataProvider, timeSource types.TimeSource, opts TokenOpts) (string, error) {
	sig, err := getSigner(config)
	if err != nil {
		return "", err
	}

	now := timeSource.Now()
	claims := AccessTokenClaims{
		Claims: jwt.Claims{
			ID:       opts.AccessTokenID.String(),
			Subject:  account.GetAccountID().String(),
			IssuedAt: jwt.NewNumericDate(now),
			Expiry:   jwt.NewNumericDate(now.Add(opts.Expiry)),
		},
		Role: string(account.GetRole()),
	}

	raw, err := jwt.Signed(sig).Claims(claims).Serialize()
	if err != nil {
		return "", errors.Wrap(err, "signing and serializing JWT")
	}

	return raw, nil
}
