package records

import (
	"github.com/pkg/errors"
	"github.com/zbyte/go-kallax"

	"myvendor/myproject/backend/domain"
	security_helper "myvendor/myproject/backend/security/helper"
)

const (
	AccountTypeUser = "user"
	AccountTypeApp  = "app"

	secretLength = 64
)

// An Account is either a User or an App
type Account struct {
	kallax.Model `table:"accounts" pk:"id"`

	ID             kallax.UUID
	Type           string
	RoleIdentifier string
	FirstName      *string
	LastName       *string
	Secret         []byte
	EmailAddress   *string   `unique:"true"`
	PasswordHash   []byte
	DeviceToken    *string
	DeviceOs       *string

	// A User Account has an optional OrganisationID, an App Account has a mandatory OrganisationID
	// TODO Don't use pointer, see https://github.com/src-d/go-kallax/#primary-keys
	OrganisationID *kallax.UUID

	DeviceLabel *string
}

func newAccount() (*Account, error) {
	a := new(Account)
	if err := a.RegenerateSecret(); err != nil {
		return nil, err
	}
	return a, nil
}

func NewUserAccount() (*Account, error) {
	a, err := newAccount()
	if err != nil {
		return nil, err
	}

	a.Type = AccountTypeUser
	return a, nil
}

func NewAppAccount() (*Account, error) {
	a, err := newAccount()
	if err != nil {
		return nil, err
	}

	a.Type = AccountTypeApp
	return a, nil
}

func (a *Account) Role() (domain.Role, error) {
	return domain.RoleByIdentifier(a.RoleIdentifier)
}

func (a *Account) RegenerateSecret() error {
	generatedSecret, err := security_helper.GenerateRandomBytes(secretLength)
	if err != nil {
		return errors.Wrap(err, "generating secret failed")
	}
	a.Secret = generatedSecret
	return nil
}

func (a *Account) BeforeSave() error {
	// Store organisationID as null in database for foreign key checks with optional id
	if a.OrganisationID != nil && a.OrganisationID.IsEmpty() {
		a.OrganisationID = nil
	}
	return nil
}

func (a *Account) GetTokenSecret() []byte {
	return a.Secret
}

func (a *Account) GetAccountID() string {
	return a.ID.String()
}

func (a *Account) GetOrganisationID() string {
	if a.OrganisationID == nil {
		return ""
	}
	return a.OrganisationID.String()
}

func (a *Account) GetRoleIdentifier() string {
	return a.RoleIdentifier
}
