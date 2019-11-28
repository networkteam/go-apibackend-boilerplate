package authentication_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"myvendor/myproject/backend/api/helper"
	test_db "myvendor/myproject/backend/test/db"
	test_graphql "myvendor/myproject/backend/test/graphql"
)

const appUserLoginGql = `
mutation ($emailAddress: String!, $password: String!, $deviceToken: String, $deviceOs: String) {
	result: loginAppUser(credentials: {emailAddress: $emailAddress, password: $password, deviceToken: $deviceToken, deviceOs: $deviceOs}) {
		account {
			id
		}
		authToken
		error {
			code
		}
	}
}
`

type appUserLoginResult struct {
	Data struct {
		AppUserLogin struct {
			Account struct {
				Id string `json:"id"`
			} `json:"account"`
			AuthToken string `json:"authToken"`
			Error     struct {
				Code string `json:"code"`
			} `json:"error"`
		} `json:"result"`
	} `json:"data"`
}

func Test_AppUserLogin_With_Valid_Credentials(t *testing.T) {
	db, cleanup := test_db.CreateTestDatabase(t)
	defer cleanup()

	test_db.ExecFixtures(t, db, "base")

	expectedAccountID := "5c216747-9bdb-4be7-b65e-3249cb5243dc"

	query := test_graphql.GraphqlQuery{
		Query: appUserLoginGql,
		Variables: map[string]interface{}{
			"emailAddress": "supervisor@example.com",
			"password":     "myRandomPassword",
		},
	}
	var loginResult appUserLoginResult

	req := test_graphql.NewRequest(t, query)
	test_graphql.Handle(t, db, helper.FixedTime(), req, &loginResult)

	require.Empty(t, loginResult.Data.AppUserLogin.Error.Code, "data.login.error.code")
	assert.Equal(t, expectedAccountID, loginResult.Data.AppUserLogin.Account.Id, "data.login.account.id")
	assert.NotEmpty(t, loginResult.Data.AppUserLogin.AuthToken, "data.login.authToken")
}

func Test_AppUserLogin_With_Invalid_Password(t *testing.T) {
	db, cleanup := test_db.CreateTestDatabase(t)
	defer cleanup()

	test_db.ExecFixtures(t, db, "base")

	query := test_graphql.GraphqlQuery{
		Query: appUserLoginGql,
		Variables: map[string]interface{}{
			"emailAddress": "test@example.com",
			"password":     "wrongPassword",
		},
	}
	var loginResult appUserLoginResult

	req := test_graphql.NewRequest(t, query)
	test_graphql.Handle(t, db, helper.FixedTime(), req, &loginResult)

	require.Equal(t, "invalidCredentials", loginResult.Data.AppUserLogin.Error.Code, "data.login.error.code")
	assert.Empty(t, loginResult.Data.AppUserLogin.Account.Id, "data.login.account.id")
	assert.Empty(t, loginResult.Data.AppUserLogin.AuthToken, "data.login.authToken")
}

func Test_AppUserLogin_With_Unknown_EmailAddress(t *testing.T) {
	db, cleanup := test_db.CreateTestDatabase(t)
	defer cleanup()

	test_db.ExecFixtures(t, db, "base")

	query := test_graphql.GraphqlQuery{
		Query: appUserLoginGql,
		Variables: map[string]interface{}{
			"emailAddress": "not-exists@example.com",
			"password":     "somePassword",
		},
	}
	var loginResult appUserLoginResult

	req := test_graphql.NewRequest(t, query)
	test_graphql.Handle(t, db, helper.FixedTime(), req, &loginResult)

	require.Equal(t, "invalidCredentials", loginResult.Data.AppUserLogin.Error.Code, "data.login.error.code")
	assert.Empty(t, loginResult.Data.AppUserLogin.Account.Id, "data.login.account.id")
	assert.Empty(t, loginResult.Data.AppUserLogin.AuthToken, "data.login.authToken")
}

func Test_AppUserLogin_Has_No_Expiration(t *testing.T) {
	db, cleanup := test_db.CreateTestDatabase(t)
	defer cleanup()

	test_db.ExecFixtures(t, db, "base")

	query := test_graphql.GraphqlQuery{
		Query: appUserLoginGql,
		Variables: map[string]interface{}{
			"emailAddress": "supervisor@example.com",
			"password":     "myRandomPassword",
		},
	}
	loginTime := helper.FixedTime()

	// Login app user

	var loginResult appUserLoginResult

	req := test_graphql.NewRequest(t, query)
	test_graphql.Handle(t, db, loginTime, req, &loginResult)
	authToken := loginResult.Data.AppUserLogin.AuthToken

	require.Empty(t, loginResult.Data.AppUserLogin.Error.Code, "data.login.error.code")

	// Check login status a few minutes later

	firstCheckTime := loginTime.Add(7 * time.Minute)

	var loginStatusResult loginStatusResult

	req = test_graphql.NewRequest(t, test_graphql.GraphqlQuery{
		Query: loginStatusGql,
	})
	req.Header.Set("Authorization", authToken)
	test_graphql.Handle(t, db, firstCheckTime, req, &loginStatusResult)

	require.Equal(t, true, loginStatusResult.Data.LoginStatus, "data.loginStatus")

	// Check login status 6 hours and some minutes later

	laterCheckTime := loginTime.Add(6*time.Hour + 17*time.Minute)

	req = test_graphql.NewRequest(t, test_graphql.GraphqlQuery{
		Query: loginStatusGql,
	})
	req.Header.Set("Authorization", authToken)
	test_graphql.Handle(t, db, laterCheckTime, req, &loginStatusResult)

	require.Equal(t, true, loginStatusResult.Data.LoginStatus, "data.loginStatus")
}
