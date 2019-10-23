package authorization

import "myvendor/myproject/backend/domain"

func (a *Authorizer) AllowsOrganisationCreateCmd(cmd domain.OrganisationCreateCmd) error {
	return a.RequireRole(domain.RoleSystemAdministrator)
}
