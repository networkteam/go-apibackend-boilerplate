package authorization_test

import (
	"testing"

	"github.com/gofrs/uuid"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/security/authentication"
	"myvendor.mytld/myproject/backend/security/authorization"
)

func TestAuthorizer_AllowsOrganisationCreateCmd(t *testing.T) {
	ownUUID := uuid.Must(uuid.NewV4())
	otherUUID := uuid.Must(uuid.NewV4())

	tt := []struct {
		description string
		authCtx     authentication.AuthContext
		cmd         domain.OrganisationCreateCmd
		wantErr     bool
	}{
		{
			description: "allows SystemAdministrator to create organisation",
			authCtx:     authentication.AuthContext{Role: domain.RoleSystemAdministrator, Authenticated: true},
			cmd:         domain.OrganisationCreateCmd{OrganisationID: otherUUID},
		},
		{
			description: "denies OrganisationAdministrator to create organisation",
			authCtx:     authentication.AuthContext{Role: domain.RoleOrganisationAdministrator, Authenticated: true, OrganisationID: &ownUUID},
			cmd:         domain.OrganisationCreateCmd{OrganisationID: otherUUID},
			wantErr:     true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.description, func(t *testing.T) {
			err := authorization.NewAuthorizer(tc.authCtx).AllowsOrganisationCreateCmd(tc.cmd)
			if (err != nil) != tc.wantErr {
				t.Errorf("Authorizer.AllowsOrganisationCreateCmd(cmd) error = %v, expected error: %v", err, tc.wantErr)
			}
		})
	}
}

func TestAuthorizer_AllowsOrganisationUpdateCmd(t *testing.T) {
	ownUUID := uuid.Must(uuid.NewV4())
	otherUUID := uuid.Must(uuid.NewV4())

	tt := []struct {
		description string
		authCtx     authentication.AuthContext
		cmd         domain.OrganisationUpdateCmd
		wantErr     bool
	}{
		{
			description: "allows SystemAdministrator to update organisation",
			authCtx:     authentication.AuthContext{Role: domain.RoleSystemAdministrator, Authenticated: true},
			cmd:         domain.OrganisationUpdateCmd{OrganisationID: otherUUID},
		},
		{
			description: "denies OrganisationAdministrator to update organisation",
			authCtx:     authentication.AuthContext{Role: domain.RoleOrganisationAdministrator, Authenticated: true, OrganisationID: &ownUUID},
			cmd:         domain.OrganisationUpdateCmd{OrganisationID: otherUUID},
			wantErr:     true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.description, func(t *testing.T) {
			err := authorization.NewAuthorizer(tc.authCtx).AllowsOrganisationUpdateCmd(tc.cmd)
			if (err != nil) != tc.wantErr {
				t.Errorf("Authorizer.AllowsOrganisationUpdateCmd(cmd) error = %v, expected error: %v", err, tc.wantErr)
			}
		})
	}
}

func TestAuthorizer_AllowsOrganisationDeleteCmd(t *testing.T) {
	ownUUID := uuid.Must(uuid.NewV4())
	otherUUID := uuid.Must(uuid.NewV4())

	tt := []struct {
		description string
		authCtx     authentication.AuthContext
		cmd         domain.OrganisationDeleteCmd
		wantErr     bool
	}{
		{
			description: "allows SystemAdministrator to delete organisation",
			authCtx:     authentication.AuthContext{Role: domain.RoleSystemAdministrator, Authenticated: true},
			cmd:         domain.OrganisationDeleteCmd{OrganisationID: otherUUID},
		},
		{
			description: "denies OrganisationAdministrator to delete organisation",
			authCtx:     authentication.AuthContext{Role: domain.RoleOrganisationAdministrator, Authenticated: true, OrganisationID: &ownUUID},
			cmd:         domain.OrganisationDeleteCmd{OrganisationID: otherUUID},
			wantErr:     true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.description, func(t *testing.T) {
			err := authorization.NewAuthorizer(tc.authCtx).AllowsOrganisationDeleteCmd(tc.cmd)
			if (err != nil) != tc.wantErr {
				t.Errorf("Authorizer.AllowsOrganisationDeleteCmd(cmd) error = %v, expected error: %v", err, tc.wantErr)
			}
		})
	}
}
