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

const createOrganisationGQL = `
	mutation CreateOrganisation($name: String!) {
		result: createOrganisation(
			name: $name,
		) {
			id
		}
	}
`

func TestMutationResolver_CreateOrganisation(t *testing.T) {
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
				"name": "Next big thing",
			},
			expects: func(t *testing.T, db *sql.DB, auth test_auth.FixedAuthTokenData, res result) {
				test_graphql.RequireNoErrors(t, res.GraphqlErrors)

				require.NotNil(t, res.Data.Result)
				organisation, err := repository.FindOrganisationByID(context.Background(), db, res.Data.Result.ID, nil)
				require.NoError(t, err)

				assert.Equal(t, "Next big thing", organisation.Name)
			},
		},
		{
			name:          "with OrganisationAdministrator",
			applyAuthFunc: test_auth.ApplyFixedAuthValuesOrganisationAdministrator,
			fixtures:      []string{"base"},
			variables: map[string]interface{}{
				"name": "My new corp",
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
				Query:     createOrganisationGQL,
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
