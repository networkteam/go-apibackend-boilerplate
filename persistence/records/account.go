package records

import (
	"github.com/gofrs/uuid"
	"github.com/friendsofgo/errors"
	"github.com/zbyte/go-kallax"

	"myvendor.mytld/myproject/backend/domain"
	security_helper "myvendor.mytld/myproject/backend/security/helper"
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
	EmailAddress   *string `unique:"true"`
	PasswordHash   []byte
	DeviceToken    *string
	DeviceOs       *string
	DeviceLabel    *string

	Organisation *Organisation `fk:",inverse"`
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

func (a *Account) GetTokenSecret() []byte {
	return a.Secret
}

func (a *Account) GetAccountID() uuid.UUID {
	return uuid.UUID(a.ID)
}

func (a *Account) GetOrganisationID() uuid.UUID {
	organisationID := valueToUUID(a.Value(Schema.Account.OrganisationFK.String()))
	if organisationID != nil {
		return uuid.UUID(*organisationID)
	}
	return uuid.Nil
}

func (a *Account) GetRoleIdentifier() string {
	return a.RoleIdentifier
}
