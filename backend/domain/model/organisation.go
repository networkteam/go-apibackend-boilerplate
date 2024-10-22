package model

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/networkteam/construct/v2"
)

type Organisation struct {
	construct.Table `table_name:"organisations"`

	ID   uuid.UUID `read_col:"organisations.organisation_id" write_col:"organisation_id"`
	Name string    `read_col:"organisations.name,sortable" write_col:"name"`

	CreatedAt time.Time `read_col:"organisations.created_at,sortable"`
	UpdatedAt time.Time `read_col:"organisations.updated_at,sortable"`
}
