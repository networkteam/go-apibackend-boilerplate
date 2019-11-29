package records

import (
	"github.com/zbyte/go-kallax"
)

type Organisation struct {
	kallax.Model `table:"organisations" pk:"id"`

	ID             kallax.UUID
	Name           string `kallax:"organisation_name",unique:"true"`
}
