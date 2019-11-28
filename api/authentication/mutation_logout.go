package authentication

import (
	"context"

	"github.com/apex/log"
	"github.com/pkg/errors"
	"github.com/zbyte/go-kallax"

	"myvendor/myproject/backend/api"
	"myvendor/myproject/backend/persistence/records"
	"myvendor/myproject/backend/security/authentication"
)

func (r *MutationResolver) Logout(ctx context.Context) (*api.Error, error) {
	log.
		Debug("Handling logout")

	authCtx := authentication.GetAuthContext(ctx)
	accountId := authCtx.AccountID

	accountStore := records.NewAccountStore(r.Db)
	query := records.NewAccountQuery().
		Where(kallax.Eq(records.Schema.Account.ID, kallax.UUID(accountId)))

	account, err := accountStore.FindOne(query)

	if err != nil && err != kallax.ErrNotFound {
		return nil, errors.Wrap(err, "could not query accounts")
	}

	if account != nil {
		account.DeviceToken = nil
		account.DeviceOs = nil
		_, err = accountStore.Update(account, records.Schema.Account.DeviceToken, records.Schema.Account.DeviceOs)
		if err != nil {
			return nil, errors.Wrap(err, "could not update account")
		}
	}

	w := api.GetHTTPResponse(ctx)
	req := api.GetHTTPRequest(ctx)

	authentication.DeleteAuthTokenCookie(w, req)

	return nil, nil
}
