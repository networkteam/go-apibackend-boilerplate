package model

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/networkteam/construct/v2"
)

type AuthAccessToken struct {
	construct.Table `table_name:"auth_access_tokens"`

	ID            uuid.UUID `read_col:"auth_access_tokens.auth_access_token_id" write_col:"auth_access_token_id"`
	AuthSessionID uuid.UUID `read_col:"auth_access_tokens.auth_session_id" write_col:"auth_session_id"`
	ExpiresAt     time.Time `read_col:"auth_access_tokens.expires_at" write_col:"expires_at"`
	CreatedAt     time.Time `read_col:"auth_access_tokens.created_at" write_col:"created_at"`

	// AuthSession is side-loaded if IncludeAuthSession is set in query options
	AuthSession *AuthSession
}
