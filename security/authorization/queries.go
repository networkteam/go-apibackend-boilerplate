package authorization

import "myvendor.mytld/myproject/backend/domain"

func (a *Authorizer) AllowsOrganisationsQuery(query domain.OrganisationsQuery) error {
	return a.RequireRole(domain.RoleSystemAdministrator)
}

func (a *Authorizer) AllowsOrganisationQuery(query domain.OrganisationQuery) error {
	return a.RequireUserForOrganisation(&query.OrganisationID)
}
