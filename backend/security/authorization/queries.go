package authorization

import (
	"myvendor.mytld/myproject/backend/domain/model"
	"myvendor.mytld/myproject/backend/domain/query"
	"myvendor.mytld/myproject/backend/domain/types"
)

func (a *Authorizer) AllowsOrganisationQuery(query query.OrganisationQuery) error {
	return a.check(
		satisfyAny(
			requireRole(types.RoleSystemAdministrator),
			requireSameOrganisation(&query.OrganisationID),
		),
	)
}

func (a *Authorizer) AllowsAndFilterAllOrganisationsQuery(query *query.OrganisationsQuery) error {
	return a.check(
		satisfyAny(
			requireRole(types.RoleSystemAdministrator),
			requireAll(
				requireRole(types.RoleOrganisationAdministrator),
				setOrganisationID(query),
			),
		),
	)
}

func (a *Authorizer) AllowsAccountView(record model.Account) error {
	return a.check(
		satisfyAny(
			// System administrators can view any account
			requireRole(types.RoleSystemAdministrator),
			// Accounts can view their own account
			requireSameAccount(&record.ID),
			// Organisation administrators can view accounts in their organisation
			requireAll(
				requireRole(types.RoleOrganisationAdministrator),
				requireSameOrganisation(uuidOrNil(record.OrganisationID)),
			),
		),
	)
}

func (a *Authorizer) AllowsAndFilterAllAccountsQuery(query *query.AccountsQuery) error {
	return a.check(
		satisfyAny(
			requireRole(types.RoleSystemAdministrator),
			requireAll(
				requireRole(types.RoleOrganisationAdministrator),
				setOrganisationID(query),
			),
		),
	)
}

func (a *Authorizer) AllowsAllAccountsQuery() error {
	return a.check(
		requireRole(types.RoleSystemAdministrator),
	)
}
