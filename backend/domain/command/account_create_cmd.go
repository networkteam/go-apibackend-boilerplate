package command

import (
	"strings"

	"github.com/friendsofgo/errors"
	"github.com/gofrs/uuid"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/domain/model"
	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/security/helper"
)

type AccountCreateCmd struct {
	AccountID      uuid.UUID
	EmailAddress   string
	Role           types.Role
	OrganisationID uuid.NullUUID
	password       string
}

func NewAccountCreateCmd(emailAddress string, role types.Role, password string) (cmd AccountCreateCmd, err error) {
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

func (c AccountCreateCmd) Validate(_ domain.Config) error {
	if isBlank(c.EmailAddress) {
		return types.FieldError{
			Field: "emailAddress",
			Code:  types.ErrorCodeRequired,
		}
	}
	if isBlank(c.password) {
		return types.FieldError{
			Field: "password",
			Code:  types.ErrorCodeRequired,
		}
	}
	if err := helper.ValidatePassword(c.password); err != nil {
		return types.FieldError{
			Field: "password",
			Code:  err.Error(),
		}
	}
	if !c.Role.IsValid() {
		return types.FieldError{
			Field: "role",
			Code:  types.ErrorCodeInvalid,
		}
	}
	// organisationID must be set iff role is not SystemAdministrator
	if !((c.Role != types.RoleSystemAdministrator) == c.OrganisationID.Valid) {
		return types.FieldError{
			Field: "organisationId",
			Code:  types.ErrorCodeRequired,
		}
	}
	return nil
}

func (c AccountCreateCmd) NewAccount(config domain.Config) (model.Account, error) {
	accountSecret, err := model.NewAccountSecret()
	if err != nil {
		return model.Account{}, errors.Wrap(err, "generate account secret")
	}
	passwordHash, err := helper.GenerateHashFromPassword([]byte(c.password), config.HashCost)
	if err != nil {
		return model.Account{}, errors.Wrap(err, "hashing password")
	}
	account := model.Account{
		ID:             c.AccountID,
		EmailAddress:   c.EmailAddress,
		Secret:         accountSecret,
		PasswordHash:   passwordHash,
		Role:           c.Role,
		OrganisationID: c.OrganisationID,
	}
	return account, nil
}
