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

const updateOrganisationGQL = `
	mutation UpdateOrganisation($id: UUID!, $name: String!) {
		result: updateOrganisation(
			id: $id,
			name: $name,
		) {
			id
		}
	}
`

func TestMutationResolver_UpdateOrganisation(t *testing.T) {
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
		expects       func(t *testing.T, db *sql.DB, res result)
	}{
		{
			name:          "with SystemAdministrator and valid values",
			applyAuthFunc: test_auth.ApplyAuthValuesFuncSystemAdministrator("admin@example.com"),
			fixtures:      []string{"base"},
			variables: map[string]interface{}{
				"id":   "6330de58-2761-411e-a243-bec6d0c53876",
				"name": "Acme Ltd.",
			},
			expects: func(t *testing.T, db *sql.DB, res result) {
				test_graphql.RequireNoErrors(t, res.GraphqlErrors)

				require.NotNil(t, res.Data.Result)
				organisation, err := repository.FindOrganisationByID(context.Background(), db, res.Data.Result.ID, nil)
				require.NoError(t, err)

				assert.Equal(t, "Acme Ltd.", organisation.Name)
			},
		},
		{
			name:          "with OrganisationAdministrator",
			applyAuthFunc: test_auth.ApplyAuthValuesFuncOrganisationAdministrator("admin+acmeinc@example.com"),
			fixtures:      []string{"base"},
			variables: map[string]interface{}{
				"id":   "6330de58-2761-411e-a243-bec6d0c53876",
				"name": "Test corp",
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
				Query:     updateOrganisationGQL,
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
