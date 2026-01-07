package model

import (
	"maps"
	"time"

	"github.com/gofrs/uuid"
	"github.com/networkteam/construct/v2"
)

type AuthSession struct {
	construct.Table `table_name:"auth_sessions"`

	ID         uuid.UUID   `read_col:"auth_sessions.auth_session_id" write_col:"auth_session_id"`
	AccountID  uuid.UUID   `read_col:"auth_sessions.account_id" write_col:"account_id"`
	DeviceInfo *DeviceInfo `read_col:"auth_sessions.device_info" write_col:"device_info"`
	CreatedAt  time.Time   `read_col:"auth_sessions.created_at" write_col:"created_at"`

	// AuthAccessTokens is side-loaded if IncludeAuthAccessTokens is set in query options
	AuthAccessTokens []AuthAccessToken
	// Account is side-loaded if IncludeAccount is set in query options
	Account *Account
}

type DeviceInfo map[string]any

func (i DeviceInfo) Equal(other DeviceInfo) bool {
	return maps.Equal(i, other)
}
