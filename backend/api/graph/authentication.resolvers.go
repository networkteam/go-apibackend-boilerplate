package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.49

import (
	"context"

	logger "github.com/apex/log"
	fog_errors "github.com/friendsofgo/errors"
	"myvendor.mytld/myproject/backend/api"
	"myvendor.mytld/myproject/backend/api/graph/helper"
	"myvendor.mytld/myproject/backend/api/graph/model"
	"myvendor.mytld/myproject/backend/domain/command"
	"myvendor.mytld/myproject/backend/domain/query"
	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/handler"
	"myvendor.mytld/myproject/backend/persistence/repository"
	"myvendor.mytld/myproject/backend/security/authentication"
)

// Login is the resolver for the login field.
func (r *mutationResolver) Login(ctx context.Context, credentials model.LoginCredentials) (*model.LoginResult, error) {
	defer helper.ConstantTime(r.SensitiveOperationConstantTime).Wait(ctx)

	cmd := command.NewLoginCmd(credentials.EmailAddress, credentials.Password)
	if credentials.KeepMeLoggedIn != nil && *credentials.KeepMeLoggedIn {
		cmd.ExtendedExpiry = true
	}

	account, err := r.finder.QueryAccountNotAuthorized(ctx, query.AccountQueryNotAuthorized{
		Opts:         helper.AccountQueryOptsFromSelection(ctx, "account"),
		EmailAddress: &cmd.EmailAddress,
	})
	switch {
	case fog_errors.Is(err, repository.ErrNotFound):
		// No op
	case err != nil:
		return nil, fog_errors.Wrap(err, "finding account")
	default:
		cmd.Account = account
	}

	err = r.handler.Login(ctx, cmd)
	if err != nil {
		if fog_errors.Is(err, handler.ErrLoginInvalidCredentials) {
			return &model.LoginResult{
				Error: &model.Error{
					Code: types.ErrorCodeInvalidCredentials,
				},
			}, nil
		}

		return nil, err
	}

	authToken, csrfToken, err := helper.SetAuthTokenCookieForAccount(ctx, account, r.TimeSource, cmd.ExtendedExpiry)
	if err != nil {
		return nil, err
	}

	return &model.LoginResult{
		Account:   helper.MapToAccount(account),
		AuthToken: authToken,
		CsrfToken: csrfToken,
	}, nil
}

// Logout is the resolver for the logout field.
func (r *mutationResolver) Logout(ctx context.Context) (*model.Error, error) {
	log := logger.
		FromContext(ctx).
		WithField("handler", "logout")

	log.
		Debug("Handling logout")

	w := api.GetHTTPResponse(ctx)
	req := api.GetHTTPRequest(ctx)
	authentication.DeleteAuthTokenCookie(w, req)

	return nil, nil
}

// LoginStatus is the resolver for the loginStatus field.
func (r *queryResolver) LoginStatus(ctx context.Context) (bool, error) {
	authCtx := authentication.GetAuthContext(ctx)
	return authCtx.Authenticated, nil
}

// CurrentAccount is the resolver for the currentAccount field.
func (r *queryResolver) CurrentAccount(ctx context.Context) (*model.Account, error) {
	authCtx := authentication.GetAuthContext(ctx)
	account, err := r.finder.QueryAccount(ctx, query.AccountQuery{
		AccountID: authCtx.AccountID,
		Opts:      helper.AccountQueryOptsFromSelection(ctx),
	})
	if err != nil {
		return nil, fog_errors.Wrap(err, "finding account")
	}

	return helper.MapToAccount(account), nil
}
