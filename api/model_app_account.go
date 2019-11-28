package api

import (
	"github.com/gofrs/uuid"
)

type AppAccount struct {
	ID             uuid.UUID `json:"id"`
	OrganisationID uuid.UUID `json:"organisation"`
	DeviceLabel    string    `json:"deviceLabel"`
	Role           Role      `json:"role"`
}
