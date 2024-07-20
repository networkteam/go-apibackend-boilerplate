package authorization

import (
	"fmt"
	"strings"

	"github.com/gofrs/uuid"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/security/authentication"
)

func NewAuthorizer(authCtx authentication.AuthContext) *Authorizer {
	return &Authorizer{
		authCtx: authCtx,
	}
}

type Authorizer struct {
	authCtx authentication.AuthContext
}

func (a *Authorizer) hasRole(anyOfRoles ...domain.Role) bool {
	if len(anyOfRoles) == 0 {
		panic("must specify at least one role")
	}

	currentRole := a.authCtx.Role
	for _, role := range anyOfRoles {
		if currentRole == role {
			return true
		}
	}

	return false
}

func (a *Authorizer) requireRole(anyOfRoles ...domain.Role) error {
	if a.hasRole(anyOfRoles...) {
		return nil
	}

	if len(anyOfRoles) == 1 {
		return authorizationError{fmt.Sprintf("requires role %s", anyOfRoles[0])}
	}

	roleIdentifiers := make([]string, len(anyOfRoles))
	for i, role := range anyOfRoles {
		roleIdentifiers[i] = string(role)
	}
	return authorizationError{fmt.Sprintf("requires role in %s", strings.Join(roleIdentifiers, ","))}
}

//nolint:unused // Prepared for future use
func (a *Authorizer) requireSameAccount(accountID *uuid.UUID) error {
	if !a.isSameAccount(accountID) {
		return authorizationError{"requires to be same account"}
	}

	return nil
}

//nolint:unused // Prepared for future use
func (a *Authorizer) isSameAccount(accountID *uuid.UUID) bool {
	if a.authCtx.AccountID == uuid.Nil {
		return false
	}

	if accountID == nil {
		return false
	}

	if a.authCtx.AccountID != *accountID {
		return false
	}

	return true
}

func (a *Authorizer) requireSameOrganisation(organisationID *uuid.UUID) error {
	if !a.isSameOrganisation(organisationID) {
		return authorizationError{"requires to be same organisation"}
	}

	return nil
}

func (a *Authorizer) isSameOrganisation(organisationID *uuid.UUID) bool {
	if a.authCtx.OrganisationID == nil || organisationID == nil {
		return false
	}

	return *a.authCtx.OrganisationID == *organisationID
}

func (a *Authorizer) requireOrganisationAdministrator(organisationID *uuid.UUID) error {
	if a.requireRole(domain.RoleSystemAdministrator) == nil {
		return nil
	}

	if err := a.requireRole(domain.RoleOrganisationAdministrator); err != nil {
		return err
	}

	return a.requireSameOrganisation(organisationID)
}

//nolint:unused // Prepared for future use
func (a *Authorizer) requireOrganisationRole(organisationID *uuid.UUID) error {
	if err := a.requireRole(domain.OrganisationRoles...); err != nil {
		return err
	}

	return a.requireSameOrganisation(organisationID)
}

func (a *Authorizer) requireOrganisationAdministratorAndFilterByOrganisationID(filter domain.OrganisationIDSetter) error {
	if a.requireRole(domain.RoleSystemAdministrator) == nil {
		return nil
	}

	if err := a.requireRole(domain.RoleOrganisationAdministrator); err != nil {
		return err
	}

	// Force a filter by organisation for role OrganisationAdministrator
	if organisationID := a.authCtx.OrganisationID; organisationID != nil {
		filter.SetOrganisationID(organisationID)
	} else {
		return authorizationError{cause: "requires organisation id"}
	}

	return nil
}

func uuidOrNil(id uuid.NullUUID) *uuid.UUID {
	if id.Valid {
		return &id.UUID
	}

	return nil
}
