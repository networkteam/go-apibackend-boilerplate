package types

import (
	"errors"
	"fmt"
	"io"
	"strconv"
)

type Role string

const RoleSystemAdministrator = Role("SystemAdministrator")
const RoleOrganisationAdministrator = Role("OrganisationAdministrator")

//nolint:gochecknoglobals
var OrganisationRoles = []Role{
	RoleOrganisationAdministrator,
}

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

func (r *Role) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return ErrEnumsMustBeStrings
	}

	domainRole, err := RoleByIdentifier(str)
	if err != nil {
		return err
	}

	*r = domainRole
	return nil
}

func (r Role) MarshalGQL(w io.Writer) {
	_, _ = fmt.Fprint(w, strconv.Quote(string(r)))
}
