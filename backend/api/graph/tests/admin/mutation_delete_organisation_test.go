package admin_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"

	"myvendor.mytld/myproject/backend/api"
	"myvendor.mytld/myproject/backend/persistence/repository"
	"myvendor.mytld/myproject/backend/test"
	test_auth "myvendor.mytld/myproject/backend/test/auth"
	test_db "myvendor.mytld/myproject/backend/test/db"
	test_graphql "myvendor.mytld/myproject/backend/test/graphql"
)

const deleteOrganisationGQL = `
	mutation DeleteOrganisation($id: UUID!) {
		result: deleteOrganisation(
			id: $id,
		) {
			id
		}
	}
`

func TestMutationResolver_DeleteOrganisation(t *testing.T) {
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
			name:          "with SystemAdministrator",
			applyAuthFunc: test_auth.ApplyFixedAuthValuesSystemAdministrator,
			fixtures:      []string{"base"},
			variables: map[string]interface{}{
				"id": "6330de58-2761-411e-a243-bec6d0c53876",
			},
			expects: func(t *testing.T, db *sql.DB, auth test_auth.FixedAuthTokenData, res result) {
				test_graphql.RequireNoErrors(t, res.GraphqlErrors)

				require.NotNil(t, res.Data.Result)
				_, err := repository.FindOrganisationByID(context.Background(), db, res.Data.Result.ID)
				require.ErrorIs(t, err, repository.ErrNotFound)
			},
		},
		{
			name:          "with OrganisationAdministrator",
			applyAuthFunc: test_auth.ApplyFixedAuthValuesOrganisationAdministrator,
			fixtures:      []string{"base"},
			variables: map[string]interface{}{
				"id": "6330de58-2761-411e-a243-bec6d0c53876",
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
				Query:     deleteOrganisationGQL,
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
