package authorization_test

import (
	"testing"

	"github.com/gofrs/uuid"

	"myvendor/myproject/backend/domain"
	"myvendor/myproject/backend/security/authentication"
	"myvendor/myproject/backend/security/authorization"
)

func TestAuthorizer_RequireAdminForOrganisation(t *testing.T) {
	ownUUID := uuid.Must(uuid.NewV4())
	otherUUID := uuid.Must(uuid.NewV4())

	tt := []struct {
		description    string
		authCtx        authentication.AuthContext
		organisationID *uuid.UUID
		wantErr        bool
	}{
		{
			description:    "allows SystemAdministrator without organisation",
			authCtx:        authentication.AuthContext{Role: domain.RoleSystemAdministrator, Authenticated: true},
			organisationID: nil,
		},
		{
			description:    "allows SystemAdministrator for every organisation",
			authCtx:        authentication.AuthContext{Role: domain.RoleSystemAdministrator, Authenticated: true},
			organisationID: &ownUUID,
		},
		{
			description:    "allows OrganisationAdministrator for own organisation",
			authCtx:        authentication.AuthContext{Role: domain.RoleOrganisationAdministrator, Authenticated: true, OrganisationID: &ownUUID},
			organisationID: &ownUUID,
		},
		{
			description:    "denies OrganisationAdministrator for other organisation",
			authCtx:        authentication.AuthContext{Role: domain.RoleOrganisationAdministrator, Authenticated: true, OrganisationID: &ownUUID},
			organisationID: &otherUUID,
			wantErr:        true,
		},
		{
			description:    "denies OrganisationAdministrator without organisation",
			authCtx:        authentication.AuthContext{Role: domain.RoleOrganisationAdministrator, Authenticated: true, OrganisationID: &ownUUID},
			organisationID: nil,
			wantErr:        true,
		},
		{
			description:    "denies User for own organisation",
			authCtx:        authentication.AuthContext{Role: domain.RoleUser, Authenticated: true, OrganisationID: &ownUUID},
			organisationID: &ownUUID,
			wantErr:        true,
		},
		{
			description:    "denies App for own organisation",
			authCtx:        authentication.AuthContext{Role: domain.RoleApp, Authenticated: true, OrganisationID: &ownUUID},
			organisationID: &ownUUID,
			wantErr:        true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.description, func(t *testing.T) {
			err := authorization.NewAuthorizer(tc.authCtx).RequireAdminForOrganisation(tc.organisationID)
			if (err != nil) != tc.wantErr {
				t.Errorf("Authorizer.RequireAdminForOrganisation(organisationID) error = %v, expected error: %v", err, tc.wantErr)
			}
		})
	}
}

func TestAuthorizer_RequireUserForOrganisation(t *testing.T) {
	ownUUID := uuid.Must(uuid.NewV4())
	otherUUID := uuid.Must(uuid.NewV4())

	tt := []struct {
		description    string
		authCtx        authentication.AuthContext
		organisationID *uuid.UUID
		wantErr        bool
	}{
		{
			description:    "allows SystemAdministrator without organisation",
			authCtx:        authentication.AuthContext{Role: domain.RoleSystemAdministrator, Authenticated: true},
			organisationID: nil,
		},
		{
			description:    "allows SystemAdministrator for every organisation",
			authCtx:        authentication.AuthContext{Role: domain.RoleSystemAdministrator, Authenticated: true},
			organisationID: &ownUUID,
		},
		{
			description:    "allows OrganisationAdministrator for own organisation",
			authCtx:        authentication.AuthContext{Role: domain.RoleOrganisationAdministrator, Authenticated: true, OrganisationID: &ownUUID},
			organisationID: &ownUUID,
		},
		{
			description:    "denies OrganisationAdministrator for other organisation",
			authCtx:        authentication.AuthContext{Role: domain.RoleOrganisationAdministrator, Authenticated: true, OrganisationID: &ownUUID},
			organisationID: &otherUUID,
			wantErr:        true,
		},
		{
			description:    "denies OrganisationAdministrator without organisation",
			authCtx:        authentication.AuthContext{Role: domain.RoleOrganisationAdministrator, Authenticated: true, OrganisationID: &ownUUID},
			organisationID: nil,
			wantErr:        true,
		},
		{
			description:    "allows User for own organisation",
			authCtx:        authentication.AuthContext{Role: domain.RoleUser, Authenticated: true, OrganisationID: &ownUUID},
			organisationID: &ownUUID,
		},
		{
			description:    "denies App for own organisation",
			authCtx:        authentication.AuthContext{Role: domain.RoleApp, Authenticated: true, OrganisationID: &ownUUID},
			organisationID: &ownUUID,
			wantErr:        true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.description, func(t *testing.T) {
			err := authorization.NewAuthorizer(tc.authCtx).RequireUserForOrganisation(tc.organisationID)
			if (err != nil) != tc.wantErr {
				t.Errorf("Authorizer.RequireUserForOrganisation(organisationID) error = %v, expected error: %v", err, tc.wantErr)
			}
		})
	}
}
