package command

import (
	"github.com/gofrs/uuid"
)

type AccountDeleteCmd struct {
	AccountID      uuid.UUID
	OrganisationID uuid.NullUUID
}

func NewAccountDeleteCmd(accountID uuid.UUID, organisationID uuid.NullUUID) AccountDeleteCmd {
	return AccountDeleteCmd{
		AccountID:      accountID,
		OrganisationID: organisationID,
	}
}
