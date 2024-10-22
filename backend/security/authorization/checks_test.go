package authorization //nolint:testpackage // We want to test the internal check functions

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"

	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/security/authentication"
)

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name      string
		authRole  types.Role
		required  []types.Role
		expectErr bool
	}{
		{
			name:      "Role matches",
			authRole:  types.RoleOrganisationAdministrator,
			required:  []types.Role{types.RoleOrganisationAdministrator},
			expectErr: false,
		},
		{
			name:      "Role does not match",
			authRole:  types.RoleSystemAdministrator,
			required:  []types.Role{types.RoleOrganisationAdministrator},
			expectErr: true,
		},
		{
			name:      "Role matches one of multiple",
			authRole:  types.RoleOrganisationAdministrator,
			required:  []types.Role{types.RoleSystemAdministrator, types.RoleOrganisationAdministrator},
			expectErr: false,
		},
		{
			name:      "No roles required",
			authRole:  types.RoleSystemAdministrator,
			required:  []types.Role{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authCtx := authentication.AuthContext{
				Role: tt.authRole,
			}
			check := requireRole(tt.required...)
			err := check(authCtx)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "requires role")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRequireSameAccount(t *testing.T) {
	accountID := uuid.Must(uuid.NewV4())
	differentAccountID := uuid.Must(uuid.NewV4())

	tests := []struct {
		name      string
		authID    uuid.UUID
		inputID   *uuid.UUID
		expectErr bool
	}{
		{
			name:      "Same account ID",
			authID:    accountID,
			inputID:   &accountID,
			expectErr: false,
		},
		{
			name:      "Different account ID",
			authID:    accountID,
			inputID:   &differentAccountID,
			expectErr: true,
		},
		{
			name:      "Nil auth ID",
			authID:    uuid.Nil,
			inputID:   &accountID,
			expectErr: true,
		},
		{
			name:      "Nil input ID",
			authID:    accountID,
			inputID:   nil,
			expectErr: true,
		},
		{
			name:      "Both IDs nil",
			authID:    uuid.Nil,
			inputID:   nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authCtx := authentication.AuthContext{
				AccountID: tt.authID,
			}
			check := requireSameAccount(tt.inputID)
			err := check(authCtx)
			if tt.expectErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, authorizationError{"requires same account"})
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRequireOrganisationID(t *testing.T) {
	organisationID := uuid.Must(uuid.NewV4())
	differentOrganisationID := uuid.Must(uuid.NewV4())

	tests := []struct {
		name            string
		authOrgID       *uuid.UUID
		inputOrgID      *uuid.UUID
		expectedError   bool
		expectedMessage string
	}{
		{
			name:          "Same organisation ID",
			authOrgID:     &organisationID,
			inputOrgID:    &organisationID,
			expectedError: false,
		},
		{
			name:            "Different organisation ID",
			authOrgID:       &organisationID,
			inputOrgID:      &differentOrganisationID,
			expectedError:   true,
			expectedMessage: "requires same organisation",
		},
		{
			name:            "Nil auth organisation ID",
			authOrgID:       nil,
			inputOrgID:      &organisationID,
			expectedError:   true,
			expectedMessage: "requires same organisation",
		},
		{
			name:            "Nil input organisation ID",
			authOrgID:       &organisationID,
			inputOrgID:      nil,
			expectedError:   true,
			expectedMessage: "requires same organisation",
		},
		{
			name:            "Both organisation IDs nil",
			authOrgID:       nil,
			inputOrgID:      nil,
			expectedError:   true,
			expectedMessage: "requires same organisation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authCtx := authentication.AuthContext{
				OrganisationID: tt.authOrgID,
			}
			check := requireOrganisationID(tt.inputOrgID)
			err := check(authCtx)
			if tt.expectedError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, authorizationError{tt.expectedMessage})
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRequireNotSameAccount(t *testing.T) {
	accountID := uuid.Must(uuid.NewV4())
	differentAccountID := uuid.Must(uuid.NewV4())

	tests := []struct {
		name      string
		authID    uuid.UUID
		inputID   *uuid.UUID
		expectErr bool
	}{
		{
			name:      "Different account ID",
			authID:    accountID,
			inputID:   &differentAccountID,
			expectErr: false,
		},
		{
			name:      "Same account ID",
			authID:    accountID,
			inputID:   &accountID,
			expectErr: true,
		},
		{
			name:      "Nil auth ID",
			authID:    uuid.Nil,
			inputID:   &accountID,
			expectErr: false,
		},
		{
			name:      "Nil input ID",
			authID:    accountID,
			inputID:   nil,
			expectErr: false,
		},
		{
			name:      "Both IDs nil",
			authID:    uuid.Nil,
			inputID:   nil,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authCtx := authentication.AuthContext{
				AccountID: tt.authID,
			}
			check := requireNotSameAccount(tt.inputID)
			err := check(authCtx)
			if tt.expectErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, authorizationError{"cannot perform action on own account"})
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRequireAll(t *testing.T) {
	passCheck := func(authCtx authentication.AuthContext) error {
		return nil
	}
	failCheck := func(authCtx authentication.AuthContext) error {
		return authorizationError{"fail"}
	}

	tests := []struct {
		name      string
		checks    []authorizationCheck
		expectErr bool
	}{
		{
			name:      "All checks pass",
			checks:    []authorizationCheck{passCheck, passCheck},
			expectErr: false,
		},
		{
			name:      "One check fails",
			checks:    []authorizationCheck{passCheck, failCheck},
			expectErr: true,
		},
		{
			name:      "All checks fail",
			checks:    []authorizationCheck{failCheck, failCheck},
			expectErr: true,
		},
		{
			name:      "No checks",
			checks:    []authorizationCheck{},
			expectErr: false,
		},
	}

	authCtx := authentication.AuthContext{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := requireAll(tt.checks...)
			err := check(authCtx)
			if tt.expectErr {
				assert.Error(t, err)
				// Since the first failing check's error is returned, we can assert the error message
				assert.ErrorIs(t, err, authorizationError{"fail"})
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSatisfyAny(t *testing.T) {
	passCheck := func(authCtx authentication.AuthContext) error {
		return nil
	}
	failCheck := func(authCtx authentication.AuthContext) error {
		return authorizationError{"fail"}
	}

	tests := []struct {
		name      string
		checks    []authorizationCheck
		expectErr bool
	}{
		{
			name:      "All checks pass",
			checks:    []authorizationCheck{passCheck, passCheck},
			expectErr: false,
		},
		{
			name:      "One check passes",
			checks:    []authorizationCheck{failCheck, passCheck},
			expectErr: false,
		},
		{
			name:      "All checks fail",
			checks:    []authorizationCheck{failCheck, failCheck},
			expectErr: true,
		},
		{
			name:      "No checks",
			checks:    []authorizationCheck{},
			expectErr: true, // Should fail because no checks are satisfied
		},
	}

	authCtx := authentication.AuthContext{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := satisfyAny(tt.checks...)
			err := check(authCtx)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "any of the following required")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRequireSameOrganisationAdministrator(t *testing.T) {
	organisationID := uuid.Must(uuid.NewV4())
	differentOrganisationID := uuid.Must(uuid.NewV4())

	tests := []struct {
		name          string
		authRole      types.Role
		authOrgID     *uuid.UUID
		inputOrgID    *uuid.UUID
		expectedError bool
	}{
		{
			name:          "Role and organisation match",
			authRole:      types.RoleOrganisationAdministrator,
			authOrgID:     &organisationID,
			inputOrgID:    &organisationID,
			expectedError: false,
		},
		{
			name:          "Role matches, organisation does not",
			authRole:      types.RoleOrganisationAdministrator,
			authOrgID:     &organisationID,
			inputOrgID:    &differentOrganisationID,
			expectedError: true,
		},
		{
			name:          "Role does not match",
			authRole:      types.RoleSystemAdministrator,
			authOrgID:     &organisationID,
			inputOrgID:    &organisationID,
			expectedError: true,
		},
		{
			name:          "Nil organisation ID",
			authRole:      types.RoleOrganisationAdministrator,
			authOrgID:     nil,
			inputOrgID:    &organisationID,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authCtx := authentication.AuthContext{
				Role:           tt.authRole,
				OrganisationID: tt.authOrgID,
			}
			check := requireSameOrganisationAdministrator(tt.inputOrgID)
			err := check(authCtx)
			if tt.expectedError {
				assert.Error(t, err)
				// The error message depends on which check failed
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRequireSameOrganisation(t *testing.T) {
	organisationID := uuid.Must(uuid.NewV4())
	differentOrganisationID := uuid.Must(uuid.NewV4())

	tests := []struct {
		name          string
		authRole      types.Role
		authOrgID     *uuid.UUID
		inputOrgID    *uuid.UUID
		expectedError bool
	}{
		{
			name:          "Organisation Administrator with matching ID",
			authRole:      types.RoleOrganisationAdministrator,
			authOrgID:     &organisationID,
			inputOrgID:    &organisationID,
			expectedError: false,
		},
		{
			name:          "Role does not match",
			authRole:      types.RoleSystemAdministrator,
			authOrgID:     &organisationID,
			inputOrgID:    &organisationID,
			expectedError: true,
		},
		{
			name:          "Organisation IDs do not match",
			authRole:      types.RoleOrganisationAdministrator,
			authOrgID:     &organisationID,
			inputOrgID:    &differentOrganisationID,
			expectedError: true,
		},
		{
			name:          "Nil organisation ID",
			authRole:      types.RoleOrganisationAdministrator,
			authOrgID:     nil,
			inputOrgID:    &organisationID,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authCtx := authentication.AuthContext{
				Role:           tt.authRole,
				OrganisationID: tt.authOrgID,
			}
			check := requireSameOrganisation(tt.inputOrgID)
			err := check(authCtx)
			if tt.expectedError {
				assert.Error(t, err)
				// The error message may vary
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRequireNotAuthenticated(t *testing.T) {
	tests := []struct {
		name          string
		authenticated bool
		expectedError bool
	}{
		{
			name:          "Not authenticated",
			authenticated: false,
			expectedError: false,
		},
		{
			name:          "Authenticated",
			authenticated: true,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authCtx := authentication.AuthContext{
				Authenticated: tt.authenticated,
			}
			check := requireNotAuthenticated()
			err := check(authCtx)
			if tt.expectedError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, authorizationError{"must not be authenticated"})
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type mockOrganisationIDSetter struct {
	organisationID *uuid.UUID
}

func (m *mockOrganisationIDSetter) SetOrganisationID(organisationID *uuid.UUID) {
	m.organisationID = organisationID
}

func TestSetOrganisationID(t *testing.T) {
	organisationID := uuid.Must(uuid.NewV4())

	tests := []struct {
		name      string
		authOrgID *uuid.UUID
		expectErr bool
	}{
		{
			name:      "Auth context has organisation ID",
			authOrgID: &organisationID,
			expectErr: false,
		},
		{
			name:      "Auth context missing organisation ID",
			authOrgID: nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authCtx := authentication.AuthContext{
				OrganisationID: tt.authOrgID,
			}
			query := &mockOrganisationIDSetter{}
			check := setOrganisationID(query)
			err := check(authCtx)
			if tt.expectErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, authorizationError{"organisation ID is required"})
				assert.Nil(t, query.organisationID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.authOrgID, query.organisationID)
			}
		})
	}
}
