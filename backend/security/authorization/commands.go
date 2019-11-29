package authorization

import "myvendor.mytld/myproject/backend/domain"

func (a *Authorizer) AllowsOrganisationCreateCmd(cmd domain.OrganisationCreateCmd) error {
	return a.RequireRole(domain.RoleSystemAdministrator)
}

func (a *Authorizer) AllowsOrganisationUpdateCmd(cmd domain.OrganisationUpdateCmd) error {
	return a.RequireRole(domain.RoleSystemAdministrator)
}

func (a *Authorizer) AllowsOrganisationDeleteCmd(cmd domain.OrganisationDeleteCmd) error {
	return a.RequireRole(domain.RoleSystemAdministrator)
}
