package authentication

import (
	"context"

	"github.com/apex/log"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/zbyte/go-kallax"

	"myvendor/myproject/backend/api"
	"myvendor/myproject/backend/api/helper"
	"myvendor/myproject/backend/persistence/records"
	"myvendor/myproject/backend/security/authentication"
	security_helper "myvendor/myproject/backend/security/helper"
)

func (r *MutationResolver) Login(ctx context.Context, credentials api.LoginCredentials) (api.LoginResult, error) {
	log.
		WithField("emailAddress", credentials.EmailAddress).
		Debug("Handling login")

	accountStore := records.NewAccountStore(r.Db)

	query := records.NewAccountQuery().
		Where(kallax.Eq(records.Schema.Account.EmailAddress, credentials.EmailAddress))
	account, err := accountStore.FindOne(query)
	accountNotFound := err == kallax.ErrNotFound
	if accountNotFound {
		// Use an empty user to have constant password compare times
		account = &records.Account{
			PasswordHash: security_helper.DefaultHashForComparison(),
		}
	} else if err != nil {
		return api.LoginResult{}, errors.Wrap(err, "could not query accounts")
	}

	err = security_helper.CompareHashAndPassword(account.PasswordHash, []byte(credentials.Password))
	if err != nil || accountNotFound {
		log.
			WithField("emailAddress", credentials.EmailAddress).
			Info("Login failed")
		return api.LoginResult{
			Error: &api.Error{
				Code: "invalidCredentials",
			},
		}, nil
	}

	accountID := uuid.UUID(account.ID)

	serializedAuthToken, err := authentication.GenerateAuthToken(account, r.TimeSource, authentication.TokenOpts{Expiry: authentication.AuthTokenExpiry})
	if err != nil {
		return api.LoginResult{}, errors.Wrap(err, "could not generate auth token")
	}

	serializedCsrfToken, err := authentication.GenerateCsrfToken(account, r.TimeSource)
	if err != nil {
		return api.LoginResult{}, errors.Wrap(err, "could not generate CSRF token")
	}

	req := api.GetHTTPRequest(ctx)
	w := api.GetHTTPResponse(ctx)

	authentication.SetAuthTokenCookie(w, req, serializedAuthToken)

	log.
		WithField("emailAddress", credentials.EmailAddress).
		WithField("accountID", accountID).
		Debug("Login success")

	userAccount, err := helper.MapToUserAccount(account)
	if err != nil {
		return api.LoginResult{}, errors.Wrap(err, "could not map account to user account")
	}

	organisation, err := helper.GetOrganisationForAccount(account, r.Db)
	if err != nil {
		return api.LoginResult{}, errors.Wrap(err, "failed to fetch associated organisation")
	}

	return api.LoginResult{
		CsrfToken: serializedCsrfToken,
		Account:   userAccount,
		// Explicitly return the organisation, since the sub-root of UserAccount->Organisation does not authorize access
		// to the organisation (yet) without the OrganisationID in AuthContext
		Organisation: organisation,
	}, nil
}
