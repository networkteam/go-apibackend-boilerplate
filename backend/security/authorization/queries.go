package authorization

import (
	"myvendor.mytld/myproject/backend/domain"
)

func (a *Authorizer) AllowsOrganisationQuery(query domain.OrganisationQuery) error {
	return a.check(
		satisfyAny(
			requireRole(domain.RoleSystemAdministrator),
			requireSameOrganisation(&query.OrganisationID),
		),
	)
}

func (a *Authorizer) AllowsAndFilterAllOrganisationsQuery(query *domain.OrganisationsQuery) error {
	return a.check(
		satisfyAny(
			requireRole(domain.RoleSystemAdministrator),
			requireAll(
				requireRole(domain.RoleOrganisationAdministrator),
				setOrganisationID(query),
			),
		),
	)
}

func (a *Authorizer) AllowsAccountView(record domain.Account) error {
	return a.check(
		satisfyAny(
			requireRole(domain.RoleSystemAdministrator),
			requireSameOrganisation(uuidOrNil(record.OrganisationID)),
		),
	)
}

func (a *Authorizer) AllowsAndFilterAllAccountsQuery(query *domain.AccountsQuery) error {
	return a.check(
		satisfyAny(
			requireRole(domain.RoleSystemAdministrator),
			requireAll(
				requireRole(domain.RoleOrganisationAdministrator),
				setOrganisationID(query),
			),
		),
	)
}

func (a *Authorizer) AllowsAllAccountsQuery() error {
	return a.check(
		requireRole(domain.RoleSystemAdministrator),
	)
}
