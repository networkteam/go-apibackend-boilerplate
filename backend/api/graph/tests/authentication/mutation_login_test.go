package authentication_test

import (
	"context"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"myvendor.mytld/myproject/backend/api"
	"myvendor.mytld/myproject/backend/test"
	test_db "myvendor.mytld/myproject/backend/test/db"
	test_graphql "myvendor.mytld/myproject/backend/test/graphql"
	test_telemetry "myvendor.mytld/myproject/backend/test/telemetry"
)

const loginGQL = `
	mutation Login($emailAddress: String!, $password: String!) {
		result: login(
			credentials: {
				emailAddress: $emailAddress,
				password: $password,
			}
		) {
			account {
				id
				emailAddress
				role
				organisationId
			}
			authToken
			csrfToken
			error {
				code
			}
		}
	}
`

type loginResult struct {
	Data struct {
		Result struct {
			Account *struct {
				ID             uuid.UUID
				EmailAddress   string
				Role           string
				OrganisationID *uuid.UUID
			}
			AuthToken string
			CsrfToken string
			Error     *struct {
				Code string
			}
		}
	}
	test_graphql.GraphqlErrors
}

func TestMutationResolver_Login_WithSystemAdministrator_Valid(t *testing.T) {
	db := test_db.CreateTestDatabase(t)
	timeSource := test.FixedTime()

	test_db.ExecFixtures(t, db, "base")

	query := test_graphql.GraphqlQuery{
		Query: loginGQL,
		Variables: map[string]interface{}{
			"emailAddress": "admin@example.com",
			"password":     "myRandomPassword",
		},
	}

	var result loginResult

	metricsReader, meterProvider := test_telemetry.SetupTestMeter(t)

	req := test_graphql.NewRequest(t, query)
	resp := test_graphql.Handle(t, api.ResolverDependencies{DB: db, TimeSource: timeSource, MeterProvider: meterProvider}, req, &result)
	test_graphql.RequireNoErrors(t, result.GraphqlErrors)

	require.Nil(t, result.Data.Result.Error)

	require.NotNil(t, result.Data.Result.Account, "result.account")
	assert.Equal(t, "admin@example.com", result.Data.Result.Account.EmailAddress, "result.account.emailAddress")
	csrfToken := result.Data.Result.CsrfToken
	assert.NotEmpty(t, csrfToken, "result.csrfToken")

	setCookieHeader := resp.Header().Get("Set-Cookie")
	assert.NotEmpty(t, setCookieHeader, "Set-Cookie header is set")

	{
		metricsData := metricdata.ResourceMetrics{}
		err := metricsReader.Collect(context.Background(), &metricsData)
		require.NoError(t, err)
		spew.Dump(metricsData)
	}

	// Test we can use a restricted field after authentication

	var loginStatusResult struct {
		Data struct {
			Result bool
		}
		test_graphql.GraphqlErrors
	}

	query = test_graphql.GraphqlQuery{
		Query: `query { result: loginStatus }`,
	}

	req = test_graphql.NewRequest(t, query)
	req.Header.Set("Cookie", setCookieHeader)
	req.Header.Set("X-CSRF-Token", csrfToken)
	test_graphql.Handle(t, api.ResolverDependencies{DB: db, TimeSource: timeSource}, req, &loginStatusResult)
	test_graphql.RequireNoErrors(t, loginStatusResult.GraphqlErrors)

	assert.True(t, loginStatusResult.Data.Result, "result")
}

func TestMutationResolver_Login_WithSystemAdministrator_InvalidPassword(t *testing.T) {
	db := test_db.CreateTestDatabase(t)
	timeSource := test.FixedTime()

	test_db.ExecFixtures(t, db, "base")

	query := test_graphql.GraphqlQuery{
		Query: loginGQL,
		Variables: map[string]interface{}{
			"emailAddress": "admin@example.com",
			"password":     "not-my-password",
		},
	}

	var result loginResult

	metricsReader, meterProvider := test_telemetry.SetupTestMeter(t)

	req := test_graphql.NewRequest(t, query)
	test_graphql.Handle(t, api.ResolverDependencies{DB: db, TimeSource: timeSource, MeterProvider: meterProvider}, req, &result)
	test_graphql.RequireNoErrors(t, result.GraphqlErrors)

	require.NotNil(t, result.Data.Result.Error, "result.error")
	assert.Equal(t, "invalidCredentials", result.Data.Result.Error.Code, "result.error.code")

	test_telemetry.AssertMeterCounter(t, metricsReader, "authentication", "login.invalid_credentials", 1)

	{
		metricsData := metricdata.ResourceMetrics{}
		err := metricsReader.Collect(context.Background(), &metricsData)
		require.NoError(t, err)
		spew.Dump(metricsData)
	}
}

func TestMutationResolver_Login_WithOrganisationAdministrator_Valid(t *testing.T) {
	db := test_db.CreateTestDatabase(t)
	timeSource := test.FixedTime()

	test_db.ExecFixtures(t, db, "base")

	query := test_graphql.GraphqlQuery{
		Query: loginGQL,
		Variables: map[string]interface{}{
			"emailAddress": "admin+acmeinc@example.com",
			"password":     "myRandomPassword",
		},
	}

	var result loginResult

	req := test_graphql.NewRequest(t, query)
	test_graphql.Handle(t, api.ResolverDependencies{DB: db, TimeSource: timeSource}, req, &result)
	test_graphql.RequireNoErrors(t, result.GraphqlErrors)

	require.Nil(t, result.Data.Result.Error)

	organisationID := uuid.Must(uuid.FromString("6330de58-2761-411e-a243-bec6d0c53876"))

	require.NotNil(t, result.Data.Result.Account)
	assert.Equal(t, "admin+acmeinc@example.com", result.Data.Result.Account.EmailAddress, "result.account.emailAddress")
	require.NotNil(t, result.Data.Result.Account.OrganisationID, "result.account.organisationId")
	assert.Equal(t, organisationID, *result.Data.Result.Account.OrganisationID, "result.account.organisationId")
}
