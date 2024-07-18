package admin_test

import (
	"database/sql"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"

	"myvendor.mytld/myproject/backend/api"
	"myvendor.mytld/myproject/backend/test"
	test_auth "myvendor.mytld/myproject/backend/test/auth"
	test_db "myvendor.mytld/myproject/backend/test/db"
	test_graphql "myvendor.mytld/myproject/backend/test/graphql"
)

const allAccountsGQL = `
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
			name:          "with SystemAdministrator and ids filter",
			applyAuthFunc: test_auth.ApplyFixedAuthValuesSystemAdministrator,
			fixtures:      []string{"base"},
			variables: map[string]interface{}{
				"filter": map[string]interface{}{
					"ids": []any{
						"3ad082c7-cbda-49e1-a707-c53e1962be65",
						"f045e5d1-cdad-4964-a7e2-139c8a87346c",
					},
				},
			},
			expects: func(t *testing.T, db *sql.DB, auth test_auth.FixedAuthTokenData, res result) {
				test_graphql.RequireNoErrors(t, res.GraphqlErrors)

				assert.Len(t, res.Data.Result, 2, "result")

				for _, entry := range res.Data.Result {
					assert.Contains(t, []uuid.UUID{
						uuid.FromStringOrNil("3ad082c7-cbda-49e1-a707-c53e1962be65"),
						uuid.FromStringOrNil("f045e5d1-cdad-4964-a7e2-139c8a87346c"),
					}, entry.ID)
				}

				assert.Equal(t, 2, res.Data.Meta.Count, "meta.count")
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
				Query:     allAccountsGQL,
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
