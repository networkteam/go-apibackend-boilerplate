package domain

import "errors"

type Role string

const RoleSystemAdministrator = Role("SystemAdministrator")
const RoleOrganisationAdministrator = Role("OrganisationAdministrator")
const RoleUser = Role("User")
const RoleApp = Role("App")

var ErrUnknownRole = errors.New("unknown role")

func RoleByIdentifier(roleIdentifier string) (Role, error) {
	r := Role(roleIdentifier)
	if !r.IsValid() {
		return r, ErrUnknownRole
	}
	return r, nil
}

func (r Role) IsValid() bool {
	switch r {
	case RoleSystemAdministrator:
	case RoleOrganisationAdministrator:
	case RoleUser:
	case RoleApp:
	default:
		return false
	}
	return true
}

func IsValidUserRole(role Role) bool {
	return role == RoleSystemAdministrator || role == RoleOrganisationAdministrator || role == RoleUser
}

func isValidAppAccountRole(role Role) bool {
	return role == RoleApp
}
