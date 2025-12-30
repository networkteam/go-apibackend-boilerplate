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

const allOrganisationsGQL = `
	query AllOrganisations($filter: OrganisationFilter) {
		result: allOrganisations(filter: $filter) {
			id
			name
		}
		meta: _allOrganisationsMeta(filter: $filter) {
			count
		}
	}
`

func TestQueryResolver_AllOrganisations(t *testing.T) {
	type result struct {
		Data struct {
			Result []struct {
				ID   uuid.UUID
				Name string
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

				require.Len(t, res.Data.Result, 2, "result")
				assert.Equal(t, 2, res.Data.Meta.Count, "meta.count")
			},
		},
		{
			name:          "with OrganisationAdministrator",
			applyAuthFunc: test_auth.ApplyFixedAuthValuesOrganisationAdministrator,
			fixtures:      []string{"base"},
			variables:     map[string]interface{}{},
			expects: func(t *testing.T, db *sql.DB, auth test_auth.FixedAuthTokenData, res result) {
				test_graphql.RequireNoErrors(t, res.GraphqlErrors)

				require.Len(t, res.Data.Result, 1, "result")
				assert.Equalf(t, auth.OrganisationID.UUID, res.Data.Result[0].ID, "result.0.id")

				assert.Equal(t, 1, res.Data.Meta.Count, "meta.count")
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			db := test_db.CreateTestDatabase(t)
			timeSource := test.FixedTime()

			test_db.ExecFixtures(t, db, tc.fixtures...)

			query := test_graphql.GraphqlQuery{
				Query:     allOrganisationsGQL,
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
