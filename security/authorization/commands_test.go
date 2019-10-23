package authorization_test

import (
	"testing"

	"github.com/gofrs/uuid"

	"myvendor/myproject/backend/domain"
	"myvendor/myproject/backend/security/authentication"
	"myvendor/myproject/backend/security/authorization"
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
