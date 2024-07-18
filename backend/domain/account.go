package domain

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/networkteam/construct/v2"

	security_helper "myvendor.mytld/myproject/backend/security/helper"
)

const accountSecretLength = 16

type Account struct {
	construct.Table `table_name:"accounts"`

	ID             uuid.UUID     `read_col:"accounts.account_id" write_col:"account_id"`
	EmailAddress   string        `read_col:"accounts.email_address,sortable" write_col:"email_address"`
	Secret         []byte        `read_col:"accounts.secret" write_col:"secret"`
	PasswordHash   []byte        `read_col:"accounts.password_hash" write_col:"password_hash"`
	Role           Role          `read_col:"accounts.role_identifier,sortable" write_col:"role_identifier"`
	LastLogin      *time.Time    `read_col:"accounts.last_login,sortable" write_col:"last_login"`
	OrganisationID uuid.NullUUID `read_col:"accounts.organisation_id" write_col:"organisation_id"`

	CreatedAt time.Time `read_col:"accounts.created_at,sortable"`
	UpdatedAt time.Time `read_col:"accounts.updated_at,sortable"`

	// Organisation that the account is assigned to (if not system administrator), is side-loaded
	Organisation *Organisation
}

// Methods to implement authentication.AuthTokenDataProvider:

// GetTokenSecret implements authentication.TokenSecretProvider
func (a Account) GetTokenSecret() []byte {
	return a.Secret
}

// GetAccountID implements authentication.AccountIDProvider
func (a Account) GetAccountID() uuid.UUID {
	return a.ID
}

// GetOrganisationID implements authentication.OrganisationIDProvider
func (a Account) GetOrganisationID() uuid.NullUUID {
	return a.OrganisationID
}

// GetRoleIdentifier implements authentication.RoleIdentifierProvider
func (a Account) GetRoleIdentifier() string {
	return string(a.Role)
}

// GetPasswordHash implements LoginDataProvider
func (a Account) GetPasswordHash() []byte {
	return a.PasswordHash
}

func newAccountSecret() ([]byte, error) {
	return security_helper.GenerateRandomBytes(accountSecretLength)
}
