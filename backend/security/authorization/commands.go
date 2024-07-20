package authorization

import (
	"myvendor.mytld/myproject/backend/domain"
)

func (a *Authorizer) AllowsAccountCreateCmd(cmd domain.AccountCreateCmd) error {
	if a.requireRole(domain.RoleSystemAdministrator) == nil {
		return nil
	}

	if err := a.requireRole(domain.RoleOrganisationAdministrator); err != nil {
		return err
	}
	if err := a.requireSameOrganisation(uuidOrNil(cmd.OrganisationID)); err != nil {
		return err
	}
	if cmd.Role != domain.RoleOrganisationAdministrator {
		return authorizationError{cause: "role not allowed"}
	}
	return nil
}

func (a *Authorizer) AllowsAccountUpdateCmd(cmd domain.AccountUpdateCmd) error {
	if a.requireRole(domain.RoleSystemAdministrator) == nil {
		return nil
	}

	if err := a.requireRole(domain.RoleOrganisationAdministrator); err != nil {
		return err
	}
	if cmd.CurrentOrganisationID != cmd.NewOrganisationID {
		return authorizationError{cause: "organisation may not be changed"}
	}
	if err := a.requireSameOrganisation(uuidOrNil(cmd.CurrentOrganisationID)); err != nil {
		return err
	}
	if cmd.Role != domain.RoleOrganisationAdministrator {
		return authorizationError{cause: "role not allowed"}
	}
	return nil
}

func (a *Authorizer) AllowsAccountDeleteCmd(cmd domain.AccountDeleteCmd) error {
	if cmd.AccountID == a.authCtx.AccountID {
		return authorizationError{cause: "cannot delete own account"}
	}

	return a.requireOrganisationAdministrator(uuidOrNil(cmd.OrganisationID))
}

func (a *Authorizer) AllowsOrganisationCreateCmd(domain.OrganisationCreateCmd) error {
	return a.requireRole(domain.RoleSystemAdministrator)
}

func (a *Authorizer) AllowsOrganisationUpdateCmd(domain.OrganisationUpdateCmd) error {
	return a.requireRole(domain.RoleSystemAdministrator)
}

func (a *Authorizer) AllowsOrganisationDeleteCmd(domain.OrganisationDeleteCmd) error {
	return a.requireRole(domain.RoleSystemAdministrator)
}
