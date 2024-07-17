package admin_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"myvendor.mytld/myproject/backend/api"
	"myvendor.mytld/myproject/backend/persistence/repository"
	"myvendor.mytld/myproject/backend/test"
	test_auth "myvendor.mytld/myproject/backend/test/auth"
	test_db "myvendor.mytld/myproject/backend/test/db"
	test_graphql "myvendor.mytld/myproject/backend/test/graphql"
)

const updateAccountGQL = `
	mutation UpdateAccount($id: UUID!, $role: Role!, $emailAddress: String!, $password: String, $organisationId: UUID) {
		result: updateAccount(
			id: $id,
			role: $role,
			emailAddress: $emailAddress,
			password: $password,
			organisationId: $organisationId,
		) {
			id
		}
	}
`

func TestMutationResolver_UpdateAccount(t *testing.T) {
	type result struct {
		Data struct {
			Result *struct {
				ID uuid.UUID
			}
		}
		test_graphql.GraphqlErrors
	}

	tt := []struct {
		name          string
		applyAuthFunc test_auth.ApplyAuthValuesFunc
		fixtures      []string
		variables     map[string]interface{}
		expects       func(t *testing.T, db *sql.DB, auth test_auth.FixedAuthTokenData, res result)
	}{
		{
			name:          "with SystemAdministrator and valid values",
			applyAuthFunc: test_auth.ApplyFixedAuthValuesSystemAdministrator,
			fixtures:      []string{"base"},
			variables: map[string]interface{}{
				"id":             "d7037ad0-d4bb-4dcc-8759-d82fbb3354e8",
				"role":           "SystemAdministrator",
				"emailAddress":   "test@acme.com",
				"organisationId": nil,
			},
			expects: func(t *testing.T, db *sql.DB, auth test_auth.FixedAuthTokenData, res result) {
				test_graphql.RequireNoErrors(t, res.GraphqlErrors)

				account, err := repository.FindAccountByID(context.Background(), db, uuid.Must(uuid.FromString("d7037ad0-d4bb-4dcc-8759-d82fbb3354e8")))
				require.NoError(t, err)

				assert.Equal(t, "test@acme.com", account.EmailAddress)
			},
		},
		{
			name:          "with OrganisationAdministrator and valid values",
			applyAuthFunc: test_auth.ApplyFixedAuthValuesOrganisationAdministrator,
			fixtures:      []string{"base"},
			variables: map[string]interface{}{
				"id":             "3ad082c7-cbda-49e1-a707-c53e1962be65",
				"role":           "OrganisationAdministrator",
				"emailAddress":   "test@acme.com",
				"password":       "myRandomPassword",
				"organisationId": "6330de58-2761-411e-a243-bec6d0c53876",
			},
			expects: func(t *testing.T, db *sql.DB, auth test_auth.FixedAuthTokenData, res result) {
				test_graphql.RequireNoErrors(t, res.GraphqlErrors)

				account, err := repository.FindAccountByID(context.Background(), db, uuid.Must(uuid.FromString("3ad082c7-cbda-49e1-a707-c53e1962be65")))
				require.NoError(t, err)

				assert.Equal(t, "test@acme.com", account.EmailAddress)
			},
		},
		{
			name:          "with OrganisationAdministrator and other organisation",
			applyAuthFunc: test_auth.ApplyFixedAuthValuesOrganisationAdministrator,
			fixtures:      []string{"base"},
			variables: map[string]interface{}{
				"id":             "3ad082c7-cbda-49e1-a707-c53e1962be65",
				"role":           "OrganisationAdministrator",
				"emailAddress":   "test@acme.com",
				"password":       "myRandomPassword",
				"organisationId": "dba20d09-a3df-4975-9406-2fb6fd8f0940",
			},
			expects: func(t *testing.T, db *sql.DB, auth test_auth.FixedAuthTokenData, res result) {
				test_graphql.RequireNotAuthorizedError(t, res.GraphqlErrors)
			},
		},
		{
			name:          "with OrganisationAdministrator and account in other organisation",
			applyAuthFunc: test_auth.ApplyFixedAuthValuesOrganisationAdministrator,
			fixtures:      []string{"base"},
			variables: map[string]interface{}{
				"id":             "2035f4da-f385-42c4-a609-02d9aa7290e5",
				"role":           "OrganisationAdministrator",
				"emailAddress":   "test@acme.com",
				"password":       "myRandomPassword",
				"organisationId": "dba20d09-a3df-4975-9406-2fb6fd8f0940",
			},
			expects: func(t *testing.T, db *sql.DB, auth test_auth.FixedAuthTokenData, res result) {
				test_graphql.RequireNotAuthorizedError(t, res.GraphqlErrors)
			},
		},
		{
			name:          "with OrganisationAdministrator and invalid role",
			applyAuthFunc: test_auth.ApplyFixedAuthValuesOrganisationAdministrator,
			fixtures:      []string{"base"},
			variables: map[string]interface{}{
				"id":           "3ad082c7-cbda-49e1-a707-c53e1962be65",
				"role":         "SystemAdministrator",
				"emailAddress": "test@acme.com",
				"password":     "myRandomPassword",
			},
			expects: func(t *testing.T, db *sql.DB, auth test_auth.FixedAuthTokenData, res result) {
				test_graphql.RequireNotAuthorizedError(t, res.GraphqlErrors)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			db := test_db.CreateTestDatabase(t)
			timeSource := test.FixedTime()

			test_db.ExecFixtures(t, db, tc.fixtures...)

			query := test_graphql.GraphqlQuery{
				Query:     updateAccountGQL,
				Variables: tc.variables,
			}

			var res result

			req := test_graphql.NewRequest(t, query)
			auth := tc.applyAuthFunc(t, timeSource, req)
			test_graphql.Handle(t, api.ResolverDependencies{DB: db, TimeSource: timeSource}, req, &res)
			tc.expects(t, db, auth, res)
		})
	}
}
