package authorization

import (
	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/security/authentication"
)

func (a *Authorizer) AllowsAccountCreateCmd(cmd domain.AccountCreateCmd) error {
	return a.check(
		satisfyAny(
			requireRole(domain.RoleSystemAdministrator),
			requireAll(
				requireSameOrganisationAdministrator(uuidOrNil(cmd.OrganisationID)),
				func(_ authentication.AuthContext) error {
					if cmd.Role != domain.RoleOrganisationAdministrator {
						return authorizationError{cause: "role not allowed"}
					}
					return nil
				},
			),
		),
	)
}

func (a *Authorizer) AllowsAccountUpdateCmd(cmd domain.AccountUpdateCmd) error {
	return a.check(
		satisfyAny(
			requireRole(domain.RoleSystemAdministrator),
			requireAll(
				requireSameOrganisationAdministrator(uuidOrNil(cmd.CurrentOrganisationID)),
				func(_ authentication.AuthContext) error {
					if cmd.CurrentOrganisationID != cmd.NewOrganisationID {
						return authorizationError{cause: "organisation may not be changed"}
					}
					if cmd.Role != domain.RoleOrganisationAdministrator {
						return authorizationError{cause: "role not allowed"}
					}
					return nil
				},
			),
		),
	)
}

func (a *Authorizer) AllowsAccountDeleteCmd(cmd domain.AccountDeleteCmd) error {
	return a.check(
		requireAll(
			requireNotSameAccount(&cmd.AccountID),
			satisfyAny(
				requireRole(domain.RoleSystemAdministrator),
				requireSameOrganisationAdministrator(uuidOrNil(cmd.OrganisationID)),
			),
		),
	)
}

func (a *Authorizer) AllowsOrganisationCreateCmd(domain.OrganisationCreateCmd) error {
	return a.check(
		requireRole(domain.RoleSystemAdministrator),
	)
}

func (a *Authorizer) AllowsOrganisationUpdateCmd(domain.OrganisationUpdateCmd) error {
	return a.check(
		requireRole(domain.RoleSystemAdministrator),
	)
}

func (a *Authorizer) AllowsOrganisationDeleteCmd(domain.OrganisationDeleteCmd) error {
	return a.check(
		requireRole(domain.RoleSystemAdministrator),
	)
}
