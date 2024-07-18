package domain

import (
	"strings"

	"github.com/gofrs/uuid"

	"myvendor.mytld/myproject/backend/security/helper"
)

type AccountUpdateCmd struct {
	AccountID             uuid.UUID
	EmailAddress          string
	Role                  Role
	CurrentOrganisationID uuid.NullUUID
	NewOrganisationID     uuid.NullUUID
	// Will be nil if not changed
	PasswordHash []byte
	Secret       []byte
	password     string
}

func NewAccountUpdateCmd(config Config, currentOrganisationID uuid.NullUUID, accountID uuid.UUID, emailAddress string, role Role, password string) (cmd AccountUpdateCmd, err error) {
	cmd = AccountUpdateCmd{
		CurrentOrganisationID: currentOrganisationID,
		AccountID:             accountID,
		EmailAddress:          strings.ToLower(strings.TrimSpace(emailAddress)),
		Role:                  role,
		password:              strings.TrimSpace(password),
	}
	if cmd.password != "" {
		cmd.PasswordHash, err = helper.GenerateHashFromPassword([]byte(cmd.password), config.HashCost)
		if err != nil {
			return cmd, err
		}
		cmd.Secret, err = newAccountSecret()
		if err != nil {
			return cmd, err
		}
	}
	return cmd, nil
}

func (c AccountUpdateCmd) Validate(_ Config) error {
	if IsBlank(c.EmailAddress) {
		return FieldError{
			Field: "emailAddress",
			Code:  ErrorCodeRequired,
		}
	}
	if c.password != "" {
		if err := helper.ValidatePassword(c.password); err != nil {
			return FieldError{
				Field: "password",
				Code:  err.Error(),
			}
		}
	}
	if !c.Role.IsValid() {
		return FieldError{
			Field: "role",
			Code:  ErrorCodeInvalid,
		}
	}
	// organisationID must be set iff role is not SystemAdministrator
	if !((c.Role != RoleSystemAdministrator) == c.NewOrganisationID.Valid) {
		return FieldError{
			Field: "organisationId",
			Code:  ErrorCodeRequired,
		}
	}
	return nil
}
