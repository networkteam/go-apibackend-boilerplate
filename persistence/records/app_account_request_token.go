package records

import (
	"time"

	"github.com/zbyte/go-kallax"
)

type AppAccountRequestToken struct {
	kallax.Model `table:"app_account_request_tokens" pk:"id"`

	ID             kallax.UUID
	ConnectToken   string
	RoleIdentifier string
	Expiry         time.Time
	Organisation   *Organisation `fk:",inverse"`
	DeviceLabel    string
}
