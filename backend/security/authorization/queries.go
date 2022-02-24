package authorization

import (
	"myvendor.mytld/myproject/backend/domain"
)

func (a *Authorizer) AllowsOrganisationQuery(query domain.OrganisationQuery) error {
	return a.requireOrganisationAdministrator(&query.OrganisationID)
}

func (a *Authorizer) AllowsAndFilterAllOrganisationsQuery(query *domain.OrganisationsQuery) error {
	return a.requireOrganisationAdministratorAndFilterByOrganisationID(query)
}

func (a *Authorizer) AllowsAccountView(record domain.Account) error {
	return a.requireOrganisationAdministrator(uuidOrNil(record.OrganisationID))
}

func (a *Authorizer) AllowsAndFilterAllAccountsQuery(query *domain.AccountsQuery) error {
	return a.requireOrganisationAdministratorAndFilterByOrganisationID(query)
}

func (a *Authorizer) AllowsAllAccountsQuery() error {
	return a.requireRole(domain.RoleSystemAdministrator)
}
