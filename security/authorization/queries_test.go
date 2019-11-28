package authorization_test

import (
	"testing"

	"github.com/gofrs/uuid"

	"myvendor/myproject/backend/domain"
	"myvendor/myproject/backend/security/authentication"
	"myvendor/myproject/backend/security/authorization"
)

func TestAuthorizer_AllowsOrganisationsQuery(t *testing.T) {
	ownUUID := uuid.Must(uuid.NewV4())

	tt := []struct {
		description string
		authCtx     authentication.AuthContext
		query       domain.OrganisationsQuery
		wantErr     bool
	}{
		{
			description: "allows SystemAdministrator to query organisations",
			authCtx:     authentication.AuthContext{Role: domain.RoleSystemAdministrator, Authenticated: true},
			query:       domain.OrganisationsQuery{},
		},
		{
			description: "denies OrganisationAdministrator to query organisations",
			authCtx:     authentication.AuthContext{Role: domain.RoleOrganisationAdministrator, Authenticated: true, OrganisationID: &ownUUID},
			query:       domain.OrganisationsQuery{},
			wantErr:     true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.description, func(t *testing.T) {
			err := authorization.NewAuthorizer(tc.authCtx).AllowsOrganisationsQuery(tc.query)
			if (err != nil) != tc.wantErr {
				t.Errorf("Authorizer.AllowsOrganisationsQuery(query) error = %v, expected error: %v", err, tc.wantErr)
			}
		})
	}
}

func TestAuthorizer_AllowsOrganisationQuery(t *testing.T) {
	ownUUID := uuid.Must(uuid.NewV4())
	otherUUID := uuid.Must(uuid.NewV4())

	tt := []struct {
		description string
		authCtx     authentication.AuthContext
		query       domain.OrganisationQuery
		wantErr     bool
	}{
		{
			description: "allows OrganisationAdministrator to query own organisation",
			authCtx:     authentication.AuthContext{Role: domain.RoleOrganisationAdministrator, Authenticated: true, OrganisationID: &ownUUID},
			query:       domain.OrganisationQuery{OrganisationID: ownUUID},
		},
		{
			description: "denies OrganisationAdministrator to query other organisation",
			authCtx:     authentication.AuthContext{Role: domain.RoleOrganisationAdministrator, Authenticated: true, OrganisationID: &ownUUID},
			query:       domain.OrganisationQuery{OrganisationID: otherUUID},
			wantErr:     true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.description, func(t *testing.T) {
			err := authorization.NewAuthorizer(tc.authCtx).AllowsOrganisationQuery(tc.query)
			if (err != nil) != tc.wantErr {
				t.Errorf("Authorizer.AllowsOrganisationQuery(query) error = %v, expected error: %v", err, tc.wantErr)
			}
		})
	}
}
