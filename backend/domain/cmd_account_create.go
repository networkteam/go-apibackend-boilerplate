package domain

import (
	"strings"

	"github.com/friendsofgo/errors"
	"github.com/gofrs/uuid"

	"myvendor.mytld/myproject/backend/security/helper"
)

type AccountCreateCmd struct {
	AccountID      uuid.UUID
	EmailAddress   string
	Role           Role
	OrganisationID uuid.NullUUID
	password       string
}

func NewAccountCreateCmd(emailAddress string, role Role, password string) (cmd AccountCreateCmd, err error) {
	accountID, err := uuid.NewV4()
	if err != nil {
		return cmd, errors.Wrap(err, "generate account id")
	}

	return AccountCreateCmd{
		AccountID:    accountID,
		EmailAddress: strings.ToLower(strings.TrimSpace(emailAddress)),
		Role:         role,
		password:     strings.TrimSpace(password),
	}, nil
}

func (c AccountCreateCmd) Validate(_ Config) error {
	if IsBlank(c.EmailAddress) {
		return FieldError{
			Field: "emailAddress",
			Code:  ErrorCodeRequired,
		}
	}
	if IsBlank(c.password) {
		return FieldError{
			Field: "password",
			Code:  ErrorCodeRequired,
		}
	}
	if err := helper.ValidatePassword(c.password); err != nil {
		return FieldError{
			Field: "password",
			Code:  err.Error(),
		}
	}
	if !c.Role.IsValid() {
		return FieldError{
			Field: "role",
			Code:  ErrorCodeInvalid,
		}
	}
	// organisationID must be set iff role is not SystemAdministrator
	if !((c.Role != RoleSystemAdministrator) == c.OrganisationID.Valid) {
		return FieldError{
			Field: "organisationId",
			Code:  ErrorCodeRequired,
		}
	}
	return nil
}

func (c AccountCreateCmd) NewAccount(config Config) (Account, error) {
	accountSecret, err := newAccountSecret()
	if err != nil {
		return Account{}, errors.Wrap(err, "generate account secret")
	}
	passwordHash, err := helper.GenerateHashFromPassword([]byte(c.password), config.HashCost)
	if err != nil {
		return Account{}, errors.Wrap(err, "hashing password")
	}
	account := Account{
		ID:             c.AccountID,
		EmailAddress:   c.EmailAddress,
		Secret:         accountSecret,
		PasswordHash:   passwordHash,
		Role:           c.Role,
		OrganisationID: c.OrganisationID,
	}
	return account, nil
}
