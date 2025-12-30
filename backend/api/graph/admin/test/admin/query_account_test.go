package admin_test

import (
	"database/sql"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"myvendor.mytld/myproject/backend/api"
	"myvendor.mytld/myproject/backend/test"
	test_auth "myvendor.mytld/myproject/backend/test/auth"
	test_db "myvendor.mytld/myproject/backend/test/db"
	test_graphql "myvendor.mytld/myproject/backend/test/graphql"
)

const accountGQL = `
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
		expects       func(t *testing.T, db *sql.DB, res result)
	}{
		{
			name:          "with SystemAdministrator",
			applyAuthFunc: test_auth.ApplyAuthValuesFuncSystemAdministrator("admin@example.com"),
			fixtures:      []string{"base"},
			variables: map[string]interface{}{
				"id": "d7037ad0-d4bb-4dcc-8759-d82fbb3354e8",
			},
			expects: func(t *testing.T, db *sql.DB, res result) {
				test_graphql.RequireNoErrors(t, res.GraphqlErrors)

				require.NotNil(t, res.Data.Result, "result")
				assert.Equal(t, "admin@example.com", res.Data.Result.EmailAddress)
			},
		},
		{
			name:          "with OrganisationAdministrator and global account",
			applyAuthFunc: test_auth.ApplyAuthValuesFuncOrganisationAdministrator("admin+acmeinc@example.com"),
			fixtures:      []string{"base"},
			variables: map[string]interface{}{
				"id": "d7037ad0-d4bb-4dcc-8759-d82fbb3354e8",
			},
			expects: func(t *testing.T, db *sql.DB, res result) {
				test_graphql.RequireNotAuthorizedError(t, res.GraphqlErrors)
			},
		},
		{
			name:          "with OrganisationAdministrator and account in other organisation",
			applyAuthFunc: test_auth.ApplyAuthValuesFuncOrganisationAdministrator("admin+acmeinc@example.com"),
			fixtures:      []string{"base"},
			variables: map[string]interface{}{
				"id": "2035f4da-f385-42c4-a609-02d9aa7290e5",
			},
			expects: func(t *testing.T, db *sql.DB, res result) {
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
				Query:     accountGQL,
				Variables: tc.variables,
			}

			var res result

			deps := api.ResolverDependencies{DB: db, TimeSource: timeSource}
			req := test_graphql.NewRequest(t, query)
			tc.applyAuthFunc(t, deps, req)
			test_graphql.HandleAdmin(t, deps, req, &res)
			tc.expects(t, db, res)
		})
	}
}
