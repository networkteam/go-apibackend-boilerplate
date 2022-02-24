package graph_test

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

const createAccountGql = `
	mutation CreateAccount($role: Role!, $emailAddress: String!, $password: String!, $organisationId: UUID) {
		result: createAccount(
			role: $role,
			emailAddress: $emailAddress,
			password: $password,
			organisationId: $organisationId,
		) {
			id
		}
	}
`

func TestMutationResolver_CreateAccount(t *testing.T) {
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
				"role":         "SystemAdministrator",
				"emailAddress": "test@acme.com",
				"password":     "myRandomPassword",
			},
			expects: func(t *testing.T, db *sql.DB, auth test_auth.FixedAuthTokenData, res result) {
				test_graphql.RequireNoErrors(t, res.GraphqlErrors)

				require.NotNil(t, res.Data.Result)
				account, err := repository.FindAccountByID(context.Background(), db, res.Data.Result.ID)
				require.NoError(t, err)

				assert.Equal(t, "test@acme.com", account.EmailAddress)
			},
		},
		{
			name:          "with OrganisationAdministrator and valid values",
			applyAuthFunc: test_auth.ApplyFixedAuthValuesOrganisationAdministrator,
			fixtures:      []string{"base"},
			variables: map[string]interface{}{
				"role":           "OrganisationAdministrator",
				"emailAddress":   "test@acme.com",
				"password":       "myRandomPassword",
				"organisationId": "6330de58-2761-411e-a243-bec6d0c53876",
			},
			expects: func(t *testing.T, db *sql.DB, auth test_auth.FixedAuthTokenData, res result) {
				test_graphql.RequireNoErrors(t, res.GraphqlErrors)

				require.NotNil(t, res.Data.Result)
				account, err := repository.FindAccountByID(context.Background(), db, res.Data.Result.ID)
				require.NoError(t, err)

				assert.Equal(t, "test@acme.com", account.EmailAddress)
			},
		},
		{
			name:          "with OrganisationAdministrator and other organisation",
			applyAuthFunc: test_auth.ApplyFixedAuthValuesOrganisationAdministrator,
			fixtures:      []string{"base"},
			variables: map[string]interface{}{
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
				Query:     createAccountGql,
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

const updateAccountGql = `
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
				Query:     updateAccountGql,
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

const deleteAccountGql = `
	mutation DeleteAccount($id: UUID!) {
		result: deleteAccount(
			id: $id,
		) {
			id
		}
	}
`

func TestMutationResolver_DeleteAccount(t *testing.T) {
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
			name:          "with SystemAdministrator and own account",
			applyAuthFunc: test_auth.ApplyFixedAuthValuesSystemAdministrator,
			fixtures:      []string{"base"},
			variables: map[string]interface{}{
				"id": "d7037ad0-d4bb-4dcc-8759-d82fbb3354e8",
			},
			expects: func(t *testing.T, db *sql.DB, auth test_auth.FixedAuthTokenData, res result) {
				test_graphql.RequireNotAuthorizedError(t, res.GraphqlErrors)
			},
		},
		{
			name:          "with SystemAdministrator and other account",
			applyAuthFunc: test_auth.ApplyFixedAuthValuesSystemAdministrator,
			fixtures:      []string{"base"},
			variables: map[string]interface{}{
				"id": "3ad082c7-cbda-49e1-a707-c53e1962be65",
			},
			expects: func(t *testing.T, db *sql.DB, auth test_auth.FixedAuthTokenData, res result) {
				test_graphql.RequireNoErrors(t, res.GraphqlErrors)

				_, err := repository.FindAccountByID(context.Background(), db, uuid.Must(uuid.FromString("3ad082c7-cbda-49e1-a707-c53e1962be65")))
				require.ErrorIs(t, err, repository.ErrNotFound)
			},
		},
		{
			name:          "with OrganisationAdministrator and own account",
			applyAuthFunc: test_auth.ApplyFixedAuthValuesOrganisationAdministrator,
			fixtures:      []string{"base"},
			variables: map[string]interface{}{
				"id": "d7037ad0-d4bb-4dcc-8759-d82fbb3354e8",
			},
			expects: func(t *testing.T, db *sql.DB, auth test_auth.FixedAuthTokenData, res result) {
				test_graphql.RequireNotAuthorizedError(t, res.GraphqlErrors)
			},
		},
		{
			name:          "with OrganisationAdministrator and other account",
			applyAuthFunc: test_auth.ApplyFixedAuthValuesOrganisationAdministrator,
			fixtures:      []string{"base"},
			variables: map[string]interface{}{
				"id": "f045e5d1-cdad-4964-a7e2-139c8a87346c",
			},
			expects: func(t *testing.T, db *sql.DB, auth test_auth.FixedAuthTokenData, res result) {
				test_graphql.RequireNoErrors(t, res.GraphqlErrors)

				_, err := repository.FindAccountByID(context.Background(), db, uuid.Must(uuid.FromString("f045e5d1-cdad-4964-a7e2-139c8a87346c")))
				require.ErrorIs(t, err, repository.ErrNotFound)
			},
		},
		{
			name:          "with OrganisationAdministrator and account in other organisation",
			applyAuthFunc: test_auth.ApplyFixedAuthValuesOrganisationAdministrator,
			fixtures:      []string{"base"},
			variables: map[string]interface{}{
				"id": "2035f4da-f385-42c4-a609-02d9aa7290e5",
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
				Query:     deleteAccountGql,
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

const allAccountsGql = `
	query AllAccounts($filter: AccountFilter) {
		result: allAccounts(filter: $filter) {
			id
			organisationId
		}
		meta: _allAccountsMeta(filter: $filter) {
			count
		}
	}
`

func TestQueryResolver_AllAccounts(t *testing.T) {
	type result struct {
		Data struct {
			Result []struct {
				ID             uuid.UUID
				OrganisationID *uuid.UUID
				emailAddress   string
			}
			Meta *struct {
				Count int
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
			name:          "with SystemAdministrator and no filter",
			applyAuthFunc: test_auth.ApplyFixedAuthValuesSystemAdministrator,
			fixtures:      []string{"base"},
			variables:     map[string]interface{}{},
			expects: func(t *testing.T, db *sql.DB, auth test_auth.FixedAuthTokenData, res result) {
				test_graphql.RequireNoErrors(t, res.GraphqlErrors)

				assert.Len(t, res.Data.Result, 4, "result")
				assert.Equal(t, 4, res.Data.Meta.Count, "meta.count")
			},
		},
		{
			name:          "with SystemAdministrator and q filter",
			applyAuthFunc: test_auth.ApplyFixedAuthValuesSystemAdministrator,
			fixtures:      []string{"base"},
			variables: map[string]interface{}{
				"filter": map[string]interface{}{
					"q": "othercorp",
				},
			},
			expects: func(t *testing.T, db *sql.DB, auth test_auth.FixedAuthTokenData, res result) {
				test_graphql.RequireNoErrors(t, res.GraphqlErrors)

				assert.Len(t, res.Data.Result, 1, "result")
				assert.Equal(t, 1, res.Data.Meta.Count, "meta.count")
			},
		},
		{
			name:          "with OrganisationAdministrator and no filter",
			applyAuthFunc: test_auth.ApplyFixedAuthValuesOrganisationAdministrator,
			fixtures:      []string{"base"},
			variables:     map[string]interface{}{},
			expects: func(t *testing.T, db *sql.DB, auth test_auth.FixedAuthTokenData, res result) {
				test_graphql.RequireNoErrors(t, res.GraphqlErrors)

				assert.Len(t, res.Data.Result, 2, "result")
				assert.Equal(t, 2, res.Data.Meta.Count, "meta.count")

				for i, entry := range res.Data.Result {
					if assert.NotNil(t, entry.OrganisationID, "result.%d.organisationId", i) {
						assert.Equal(t, auth.OrganisationID.UUID, *entry.OrganisationID, "result.%d.organisationId", i)
					}
				}
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			db := test_db.CreateTestDatabase(t)
			timeSource := test.FixedTime()

			test_db.ExecFixtures(t, db, tc.fixtures...)

			query := test_graphql.GraphqlQuery{
				Query:     allAccountsGql,
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

const accountGql = `
	query Account($id: UUID!) {
		result: Account(id: $id) {
			id
			emailAddress
		}
	}
`

func TestQueryResolver_Account(t *testing.T) {
	type result struct {
		Data struct {
			Result *struct {
				ID           uuid.UUID
				EmailAddress string
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
			name:          "with SystemAdministrator",
			applyAuthFunc: test_auth.ApplyFixedAuthValuesSystemAdministrator,
			fixtures:      []string{"base"},
			variables: map[string]interface{}{
				"id": "d7037ad0-d4bb-4dcc-8759-d82fbb3354e8",
			},
			expects: func(t *testing.T, db *sql.DB, auth test_auth.FixedAuthTokenData, res result) {
				test_graphql.RequireNoErrors(t, res.GraphqlErrors)

				require.NotNil(t, res.Data.Result, "result")
				assert.Equal(t, "admin@example.com", res.Data.Result.EmailAddress)
			},
		},
		{
			name:          "with OrganisationAdministrator and global account",
			applyAuthFunc: test_auth.ApplyFixedAuthValuesOrganisationAdministrator,
			fixtures:      []string{"base"},
			variables: map[string]interface{}{
				"id": "d7037ad0-d4bb-4dcc-8759-d82fbb3354e8",
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
				"id": "2035f4da-f385-42c4-a609-02d9aa7290e5",
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
				Query:     accountGql,
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
