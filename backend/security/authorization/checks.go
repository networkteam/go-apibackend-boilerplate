package authorization

import (
	"fmt"
	"strings"

	"github.com/gofrs/uuid"

	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/security/authentication"
)

type authorizationCheck func(authCtx authentication.AuthContext) error

func requireRole(roles ...types.Role) authorizationCheck {
	return func(authCtx authentication.AuthContext) error {
		currentRole := authCtx.Role
		for _, role := range roles {
			if currentRole == role {
				return nil
			}
		}
		return authorizationError{fmt.Sprintf("requires role %v", roles)}
	}
}

func requireSameAccount(accountID *uuid.UUID) authorizationCheck {
	return func(authCtx authentication.AuthContext) error {
		if authCtx.AccountID == uuid.Nil || accountID == nil {
			return authorizationError{"requires same account"}
		}
		if authCtx.AccountID != *accountID {
			return authorizationError{"requires same account"}
		}
		return nil
	}
}

func requireOrganisationID(organisationID *uuid.UUID) authorizationCheck {
	return func(authCtx authentication.AuthContext) error {
		if authCtx.OrganisationID == nil || organisationID == nil {
			return authorizationError{"requires same organisation"}
		}
		if *authCtx.OrganisationID != *organisationID {
			return authorizationError{"requires same organisation"}
		}
		return nil
	}
}

func requireNotSameAccount(accountID *uuid.UUID) authorizationCheck {
	return func(authCtx authentication.AuthContext) error {
		if authCtx.AccountID == uuid.Nil || accountID == nil {
			return nil
		}
		if authCtx.AccountID == *accountID {
			return authorizationError{"cannot perform action on own account"}
		}
		return nil
	}
}

func requireAll(checks ...authorizationCheck) authorizationCheck {
	return func(authCtx authentication.AuthContext) error {
		for _, check := range checks {
			if err := check(authCtx); err != nil {
				return err
			}
		}
		return nil
	}
}

func satisfyAny(checks ...authorizationCheck) authorizationCheck {
	return func(authCtx authentication.AuthContext) error {
		var errors []string
		for _, check := range checks {
			err := check(authCtx)
			if err == nil {
				return nil
			}
			errors = append(errors, err.Error())
		}
		return authorizationError{fmt.Sprintf("any of the following required: %v", strings.Join(errors, "; "))}
	}
}

func requireSameOrganisationAdministrator(organisationID *uuid.UUID) authorizationCheck {
	return requireAll(
		requireRole(types.RoleOrganisationAdministrator),
		requireOrganisationID(organisationID),
	)
}

func requireSameOrganisation(organisationID *uuid.UUID) authorizationCheck {
	return requireAll(
		requireRole(types.OrganisationRoles...),
		requireOrganisationID(organisationID),
	)
}

func requireNotAuthenticated() authorizationCheck {
	return func(authCtx authentication.AuthContext) error {
		if authCtx.Authenticated {
			return authorizationError{"must not be authenticated"}
		}
		return nil
	}
}

func setOrganisationID(query OrganisationIDSetter) authorizationCheck {
	return func(authCtx authentication.AuthContext) error {
		if authCtx.OrganisationID == nil {
			return authorizationError{"organisation ID is required"}
		}
		query.SetOrganisationID(authCtx.OrganisationID)
		return nil
	}
}

func uuidOrNil(id uuid.NullUUID) *uuid.UUID {
	if id.Valid {
		return &id.UUID
	}

	return nil
}
