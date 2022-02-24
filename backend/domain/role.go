package domain

import "errors"

type Role string

const RoleSystemAdministrator = Role("SystemAdministrator")
const RoleOrganisationAdministrator = Role("OrganisationAdministrator")

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
	default:
		return false
	}
	return true
}
