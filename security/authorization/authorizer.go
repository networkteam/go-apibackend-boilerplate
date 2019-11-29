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

func (a *Authorizer) RequireRole(anyOfRoles ...domain.Role) error {
	if len(anyOfRoles) == 0 {
		panic("must specify at least one role")
	}

	roleIdentifiers := make([]string, len(anyOfRoles))
	for i, role := range anyOfRoles {
		if a.authCtx.Role == role {
			return nil
		}
		roleIdentifiers[i] = string(role)
	}
	if len(anyOfRoles) == 1 {
		return authorizationError{fmt.Sprintf("requires role %s", anyOfRoles[0])}
	}
	return authorizationError{fmt.Sprintf("requires role in %s", strings.Join(roleIdentifiers, ","))}
}

func (a *Authorizer) RequireAdminForOrganisation(organisationID *uuid.UUID) error {
	if err := a.RequireRole(domain.RoleSystemAdministrator, domain.RoleOrganisationAdministrator); err != nil {
		return err
	}

	return a.requireOrganisation(organisationID)
}

func (a *Authorizer) RequireUserForOrganisation(organisationID *uuid.UUID) error {
	if err := a.RequireRole(domain.RoleSystemAdministrator, domain.RoleOrganisationAdministrator, domain.RoleUser); err != nil {
		return err
	}
	return a.requireOrganisation(organisationID)
}

func (a *Authorizer) requireOrganisation(organisationID *uuid.UUID) error {
	if a.authCtx.Role == domain.RoleSystemAdministrator {
		return nil
	}

	if a.authCtx.OrganisationID == nil {
		return NewAuthorizationError("requires organisation in auth context")
	}
	if organisationID == nil {
		return NewAuthorizationError("requires organisationId")
	}
	if *organisationID == uuid.Nil {
		return NewAuthorizationError("organisationId must not be empty")
	}
	if *organisationID != *a.authCtx.OrganisationID {
		return NewAuthorizationError("access to other organisation not allowed")
	}

	return nil
}
