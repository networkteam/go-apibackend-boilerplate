package helper

import (
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"myvendor/myproject/backend/api"
	"myvendor/myproject/backend/persistence/records"
)

func MapResultSetToUserAccounts(res *records.AccountResultSet) (accounts []*api.UserAccount, err error) {
	err = res.ForEach(func(userAccount *records.Account) error {
		account, err := MapToUserAccount(userAccount)
		if err != nil {
			return errors.Wrap(err, "could not map to user account")
		}
		accounts = append(accounts, account)
		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "mapping user accounts")
	}

	return
}

func MapToUserAccount(a *records.Account) (*api.UserAccount, error) {
	role, err := a.Role()
	if err != nil {
		return nil, errors.Wrap(err, "could not get role for user account")
	}

	userAccount := &api.UserAccount{
		ID:           uuid.UUID(a.ID),
		EmailAddress: StringOrEmpty(a.EmailAddress),
		Role:         api.Role(role),
	}
	if a.OrganisationID != nil && !(*a.OrganisationID).IsEmpty() {
		organisationID := uuid.UUID(*a.OrganisationID)
		userAccount.OrganisationID = &organisationID
	}
	userAccount.FirstName = StringOrEmpty(a.FirstName)
	userAccount.LastName = StringOrEmpty(a.LastName)

	return userAccount, nil
}
