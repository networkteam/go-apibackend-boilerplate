package command

import (
	"strings"

	"github.com/gofrs/uuid"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/domain/model"
	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/security/helper"
)

type AccountUpdateCmd struct {
	AccountID             uuid.UUID
	EmailAddress          string
	Role                  types.Role
	CurrentOrganisationID uuid.NullUUID
	NewOrganisationID     uuid.NullUUID
	// Will be nil if not changed
	PasswordHash []byte
	Secret       []byte
	password     string
}

func NewAccountUpdateCmd(config domain.Config, currentOrganisationID uuid.NullUUID, accountID uuid.UUID, emailAddress string, role types.Role, password string) (cmd AccountUpdateCmd, err error) {
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
		cmd.Secret, err = model.NewAccountSecret()
		if err != nil {
			return cmd, err
		}
	}
	return cmd, nil
}

func (c AccountUpdateCmd) Validate(_ domain.Config) error {
	if isBlank(c.EmailAddress) {
		return types.FieldError{
			Field: "emailAddress",
			Code:  types.ErrorCodeRequired,
		}
	}
	if c.password != "" {
		if err := helper.ValidatePassword(c.password); err != nil {
			return types.FieldError{
				Field: "password",
				Code:  err.Error(),
			}
		}
	}
	if !c.Role.IsValid() {
		return types.FieldError{
			Field: "role",
			Code:  types.ErrorCodeInvalid,
		}
	}
	// organisationID must be set iff role is not SystemAdministrator
	if !((c.Role != types.RoleSystemAdministrator) == c.NewOrganisationID.Valid) {
		return types.FieldError{
			Field: "organisationId",
			Code:  types.ErrorCodeRequired,
		}
	}
	return nil
}
