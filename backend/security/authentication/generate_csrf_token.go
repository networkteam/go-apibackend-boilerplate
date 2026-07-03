package authentication

import (
	"github.com/friendsofgo/errors"
	"github.com/go-jose/go-jose/v4/jwt"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/domain/types"
)

func GenerateCsrfToken(config domain.Config, timeSource types.TimeSource, opts TokenOpts) (string, error) {
	sig, err := getSigner(config)
	if err != nil {
		return "", err
	}

	now := timeSource.Now()

	cl := jwt.Claims{
		ID:       opts.AuthSessionID.String(),
		IssuedAt: jwt.NewNumericDate(now),
		Expiry:   jwt.NewNumericDate(now.Add(opts.Expiry)),
	}
	raw, err := jwt.Signed(sig).Claims(cl).Serialize()
	if err != nil {
		return "", errors.Wrap(err, "signing and serializing JWT")
	}

	return raw, nil
}
