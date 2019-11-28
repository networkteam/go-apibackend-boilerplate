package api

import (
	"fmt"
	"io"
	"strconv"

	"github.com/pkg/errors"

	"myvendor/myproject/backend/domain"
)

// TODO Check if external marshal of domain.Role type is better here
type Role domain.Role

func (e Role) String() string {
	return string(e)
}

func (e *Role) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return errors.Errorf("enums must be strings")
	}

	domainRole, err := domain.RoleByIdentifier(str)
	if err != nil {
		return err
	}

	*e = Role(domainRole)
	return nil
}

func (e Role) MarshalGQL(w io.Writer) {
	_, _ = fmt.Fprint(w, strconv.Quote(e.String()))
}
