package api

import (
	"github.com/gofrs/uuid"
)

type UserAccount struct {
	ID             uuid.UUID  `json:"id"`
	EmailAddress   string     `json:"emailAddress"`
	OrganisationID *uuid.UUID `json:"organisation"`
	FirstName      string     `json:"firstName"`
	LastName       string     `json:"lastName"`
	Role           Role       `json:"role"`
}
