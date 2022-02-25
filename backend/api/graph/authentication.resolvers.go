package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"database/sql"

	logger "github.com/apex/log"
	fog_errors "github.com/friendsofgo/errors"
	"myvendor.mytld/myproject/backend/api"
	"myvendor.mytld/myproject/backend/api/graph/helper"
	"myvendor.mytld/myproject/backend/api/graph/model"
	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/persistence/repository"
	"myvendor.mytld/myproject/backend/security/authentication"
	security_helper "myvendor.mytld/myproject/backend/security/helper"
)

func (r *mutationResolver) Login(ctx context.Context, credentials model.LoginCredentials) (*model.LoginResult, error) {
	log := logger.FromContext(ctx).
		WithField("handler", "login")

	log.
		WithField("emailAddress", credentials.EmailAddress).
		Debug("Handling login")

	var (
		serializedAuthToken string
		serializedCsrfToken string
		account             domain.Account
	)
	err := repository.Transactional(ctx, r.DB, func(tx *sql.Tx) error {
		var err error
		account, err = repository.FindAccountByEmailAddress(ctx, r.DB, credentials.EmailAddress)
		accountNotFound := err == repository.ErrNotFound
		if accountNotFound {
			// Use an empty user to have constant password compare times
			account = domain.Account{
				PasswordHash: security_helper.DefaultHashForComparison(r.Config.HashCost),
			}
		} else if err != nil {
			return fog_errors.Wrap(err, "finding account")
		}

		err = security_helper.CompareHashAndPassword(account.PasswordHash, []byte(credentials.Password))
		if err != nil || accountNotFound {
			if accountNotFound {
				log.
					WithField("username", credentials.EmailAddress).
					Warn("Login failed, account not found")
			} else {
				log.
					WithField("username", credentials.EmailAddress).
					WithError(err).
					Warn("Login failed, invalid password")
			}

			return api.ErrInvalidCredentials
		}

		now := r.TimeSource.Now()
		ptrNow := &now
		err = repository.UpdateAccount(ctx, r.DB, account.ID, repository.AccountChangeSet{LastLogin: &ptrNow})
		if err != nil {
			return fog_errors.Wrap(err, "updating account last login")
		}

		tokenOpts := authentication.TokenOptsForAccount(account)
		serializedAuthToken, err = authentication.GenerateAuthToken(account, r.TimeSource, tokenOpts)
		if err != nil {
			return fog_errors.Wrap(err, "generating auth token")
		}

		serializedCsrfToken, err = authentication.GenerateCsrfToken(account, r.TimeSource, tokenOpts)
		if err != nil {
			return fog_errors.Wrap(err, "generating CSRF token")
		}

		return nil
	})
	if err != nil {
		// Check if typed error is returned, then we output a LoginResult with Error and Code
		var typedError api.TypedError
		if fog_errors.As(err, &typedError) {
			return &model.LoginResult{
				Error: &model.Error{
					Code: typedError.Code(),
				},
			}, nil
		}

		return nil, fog_errors.Wrap(err, "running transaction")
	}

	req := api.GetHTTPRequest(ctx)
	w := api.GetHTTPResponse(ctx)

	authentication.SetAuthTokenCookie(w, req, serializedAuthToken)

	log.
		WithField("emailAddress", credentials.EmailAddress).
		WithField("accountID", account.ID).
		Info("Login success")

	apiAccount := helper.MapToAccount(account)

	return &model.LoginResult{
		Account:   apiAccount,
		AuthToken: serializedAuthToken,
		CsrfToken: serializedCsrfToken,
	}, nil
}

func (r *mutationResolver) Logout(ctx context.Context) (*model.Error, error) {
	log := logger.FromContext(ctx).
		WithField("handler", "logout")

	log.
		Debug("Handling logout")

	w := api.GetHTTPResponse(ctx)
	req := api.GetHTTPRequest(ctx)

	authentication.DeleteAuthTokenCookie(w, req)

	return nil, nil
}

func (r *queryResolver) LoginStatus(ctx context.Context) (bool, error) {
	authCtx := authentication.GetAuthContext(ctx)
	return authCtx.Authenticated, nil
}
