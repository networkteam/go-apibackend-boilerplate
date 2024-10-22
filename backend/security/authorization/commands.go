package authorization

import (
	"myvendor.mytld/myproject/backend/domain/command"
	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/security/authentication"
)

func (a *Authorizer) AllowsAccountCreateCmd(cmd command.AccountCreateCmd) error {
	return a.check(
		satisfyAny(
			requireRole(types.RoleSystemAdministrator),
			requireAll(
				requireSameOrganisationAdministrator(uuidOrNil(cmd.OrganisationID)),
				func(_ authentication.AuthContext) error {
					if cmd.Role != types.RoleOrganisationAdministrator {
						return authorizationError{cause: "role not allowed"}
					}
					return nil
				},
			),
		),
	)
}

func (a *Authorizer) AllowsAccountUpdateCmd(cmd command.AccountUpdateCmd) error {
	return a.check(
		satisfyAny(
			requireRole(types.RoleSystemAdministrator),
			requireAll(
				requireSameOrganisationAdministrator(uuidOrNil(cmd.CurrentOrganisationID)),
				func(_ authentication.AuthContext) error {
					if cmd.CurrentOrganisationID != cmd.NewOrganisationID {
						return authorizationError{cause: "organisation may not be changed"}
					}
					if cmd.Role != types.RoleOrganisationAdministrator {
						return authorizationError{cause: "role not allowed"}
					}
					return nil
				},
			),
		),
	)
}

func (a *Authorizer) AllowsAccountDeleteCmd(cmd command.AccountDeleteCmd) error {
	return a.check(
		requireAll(
			requireNotSameAccount(&cmd.AccountID),
			satisfyAny(
				requireRole(types.RoleSystemAdministrator),
				requireSameOrganisationAdministrator(uuidOrNil(cmd.OrganisationID)),
			),
		),
	)
}

func (a *Authorizer) AllowsOrganisationCreateCmd(command.OrganisationCreateCmd) error {
	return a.check(
		requireRole(types.RoleSystemAdministrator),
	)
}

func (a *Authorizer) AllowsOrganisationUpdateCmd(command.OrganisationUpdateCmd) error {
	return a.check(
		requireRole(types.RoleSystemAdministrator),
	)
}

func (a *Authorizer) AllowsOrganisationDeleteCmd(command.OrganisationDeleteCmd) error {
	return a.check(
		requireRole(types.RoleSystemAdministrator),
	)
}
