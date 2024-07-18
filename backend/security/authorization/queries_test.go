package authorization_test

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/security/authentication"
	"myvendor.mytld/myproject/backend/security/authorization"
)

func TestAuthorizer_AllowsAccountView(t *testing.T) {
	fixtureAccountID := uuid.Must(uuid.FromString("04086bfe-4f22-4aa3-9ed7-f85b15a83efd"))
	fixtureOrganisationID := uuid.Must(uuid.FromString("2bf9eab6-c592-4c9c-99d6-20339c845ea8"))

	tests := []struct {
		name    string
		authCtx authentication.AuthContext
		record  domain.Account
		wantErr bool
	}{
		{
			name:    "unauthenticated",
			authCtx: authentication.AuthContext{},
			record: domain.Account{
				ID:             fixtureAccountID,
				OrganisationID: uuid.NullUUID{UUID: fixtureOrganisationID, Valid: true},
				Role:           domain.RoleOrganisationAdministrator,
			},
			wantErr: true,
		},
		{
			name: "OrganisationAdministrator - account in same organisation",
			authCtx: authentication.AuthContext{
				Authenticated:  true,
				AccountID:      fixtureAccountID,
				OrganisationID: &fixtureOrganisationID,
				Role:           domain.RoleOrganisationAdministrator,
			},
			record: domain.Account{
				ID:             uuid.Must(uuid.FromString("f49c01b7-15a6-48ad-8989-f2fd4e5fa5c1")),
				OrganisationID: uuid.NullUUID{UUID: fixtureOrganisationID, Valid: true},
				Role:           domain.RoleOrganisationAdministrator,
			},
			wantErr: false,
		},
		{
			name: "OrganisationAdministrator - account in other organisation",
			authCtx: authentication.AuthContext{
				Authenticated:  true,
				AccountID:      fixtureAccountID,
				OrganisationID: &fixtureOrganisationID,
				Role:           domain.RoleOrganisationAdministrator,
			},
			record: domain.Account{
				ID:             uuid.Must(uuid.FromString("f49c01b7-15a6-48ad-8989-f2fd4e5fa5c1")),
				OrganisationID: uuid.NullUUID{UUID: uuid.Must(uuid.FromString("f9e84475-45f9-47d1-a58c-e416f1c7f39d")), Valid: true},
				Role:           domain.RoleOrganisationAdministrator,
			},
			wantErr: true,
		},
		{
			name: "SystemAdministrator",
			authCtx: authentication.AuthContext{
				Authenticated: true,
				AccountID:     fixtureAccountID,
				Role:          domain.RoleSystemAdministrator,
			},
			record: domain.Account{
				ID:             uuid.Must(uuid.FromString("f49c01b7-15a6-48ad-8989-f2fd4e5fa5c1")),
				OrganisationID: uuid.NullUUID{UUID: uuid.Must(uuid.FromString("f9e84475-45f9-47d1-a58c-e416f1c7f39d")), Valid: true},
				Role:           domain.RoleOrganisationAdministrator,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := authorization.NewAuthorizer(tt.authCtx)

			err := a.AllowsAccountView(tt.record)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
