package authentication

import (
	"context"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/gofrs/uuid"
	"github.com/zbyte/go-kallax"

	"myvendor.mytld/myproject/backend/api"
	"myvendor.mytld/myproject/backend/api/helper"
	"myvendor.mytld/myproject/backend/logger"
	"myvendor.mytld/myproject/backend/persistence/records"
	"myvendor.mytld/myproject/backend/security/authentication"
	security_helper "myvendor.mytld/myproject/backend/security/helper"
)

const (
	// 10000 days expiry ("never" for app users)
	appUserAuthTokenExpiry = 10000 * 24 * time.Hour
)

func (r *MutationResolver) LoginAppUser(ctx context.Context, credentials api.AppUserLoginCredentials) (*api.AppUserLoginResult, error) {
	log := logger.GetLogger(ctx).
		WithField("handler", "loginAppUser")

	log.
		WithField("emailAddress", credentials.EmailAddress).
		Debug("Handling login app user")

	accountStore := records.NewAccountStore(r.Db)

	query := records.NewAccountQuery().
		Where(kallax.Eq(records.Schema.Account.EmailAddress, credentials.EmailAddress)).
		Where(kallax.Eq(records.Schema.Account.Type, records.AccountTypeApp))
	account, err := accountStore.FindOne(query)
	accountNotFound := err == kallax.ErrNotFound
	if accountNotFound {
		// Use an empty user to have constant password compare times
		account = &records.Account{
			PasswordHash: security_helper.DefaultHashForComparison(),
		}
	} else if err != nil {
		return nil, errors.Wrap(err, "could not query accounts")
	}

	err = security_helper.CompareHashAndPassword(account.PasswordHash, []byte(credentials.Password))
	if err != nil || accountNotFound {
		log.
			WithField("emailAddress", credentials.EmailAddress).
			WithField("accountNotFound", accountNotFound).
			WithError(err).
			Warn("Login failed")
		return &api.AppUserLoginResult{
			Account: &api.UserAccount{},
			Error: &api.Error{
				Code: "invalidCredentials",
			},
		}, nil
	}

	accountID := uuid.UUID(account.ID)

	serializedAuthToken, err := authentication.GenerateAuthToken(account, r.TimeSource, authentication.TokenOpts{Expiry: appUserAuthTokenExpiry})
	if err != nil {
		return nil, errors.Wrap(err, "could not generate auth token")
	}

	log.
		WithField("emailAddress", credentials.EmailAddress).
		WithField("accountID", accountID).
		WithField("deviceToken", credentials.DeviceToken).
		WithField("deviceOs", credentials.DeviceOs).
		Debug("Login app user success")

	userAccount, err := helper.MapToUserAccount(account)
	if err != nil {
		return nil, errors.Wrap(err, "could not map account to user account")
	}

	organisation, err := helper.GetOrganisationForAccount(account, r.Db)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch associated organisation")
	}

	// If a device token is present, remove it from other accounts to prevent notifications after app user changed
	if credentials.DeviceToken != nil && credentials.DeviceOs != nil {
		err := accountStore.AccountCleanupDeviceTokens(*credentials.DeviceToken, *credentials.DeviceOs)
		if err != nil {
			return nil, errors.Wrap(err, "cleaning up account device tokens")
		}
	}

	account.DeviceToken = credentials.DeviceToken
	account.DeviceOs = credentials.DeviceOs
	_, err = accountStore.Update(account, records.Schema.Account.DeviceToken, records.Schema.Account.DeviceOs)
	if err != nil {
		return nil, errors.Wrap(err, "could not update account")
	}

	return &api.AppUserLoginResult{
		Account:      userAccount,
		AuthToken:    serializedAuthToken,
		Organisation: organisation,
	}, nil
}