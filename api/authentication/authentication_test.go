package authentication_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"myvendor.mytld/myproject/backend/api"
	api_handler "myvendor.mytld/myproject/backend/api/handler"
	"myvendor.mytld/myproject/backend/api/helper"
	test_auth "myvendor.mytld/myproject/backend/test/auth"
	test_db "myvendor.mytld/myproject/backend/test/db"
	test_graphql "myvendor.mytld/myproject/backend/test/graphql"
)

const loginGql = `
mutation ($emailAddress: String!, $password: String!) {
	result: login(credentials: {emailAddress: $emailAddress, password: $password}) {
		account {
			id
		}
		csrfToken
		error {
			code
		}
	}
}
`

type loginResult struct {
	Data struct {
		Login struct {
			Account struct {
				Id string `json:"id"`
			} `json:"account"`
			CsrfToken string `json:"csrfToken"`
			Error     struct {
				Code string `json:"code"`
			} `json:"error"`
		} `json:"result"`
	} `json:"data"`
	test_graphql.GraphqlErrors
}

const loginStatusGql = `
{
	loginStatus
}
`

type loginStatusResult struct {
	Data struct {
		LoginStatus bool `json:"loginStatus"`
	}
	test_graphql.GraphqlErrors
}

const logoutGql = `
mutation {
	logout {
		code
	}
}
`

func Test_Login_With_Valid_Credentials(t *testing.T) {
	db, cleanup := test_db.CreateTestDatabase(t)
	defer cleanup()

	test_db.ExecFixtures(t, db, "base")

	expectedAccountID := "0352324c-25aa-4def-935d-0eed999f1f99"

	query := test_graphql.GraphqlQuery{
		Query: loginGql,
		Variables: map[string]interface{}{
			"emailAddress": "test@example.com",
			"password":     "myRandomPassword",
		},
	}
	var loginResult loginResult

	req := test_graphql.NewRequest(t, query)
	w := test_graphql.Handle(t, api.ResolverDependencies{Db: db}, req, &loginResult)
	test_graphql.RequireNoErrors(t, loginResult.GraphqlErrors)

	cookie := readCookie(t, w, "authToken")

	require.Empty(t, loginResult.Data.Login.Error.Code, "data.login.error.code")
	assert.Equal(t, expectedAccountID, loginResult.Data.Login.Account.Id, "data.login.account.id")
	assert.NotEmpty(t, loginResult.Data.Login.CsrfToken, "data.login.csrfToken")
	assert.NotEmpty(t, cookie.Value, "cookie value is not empty")
}

func Test_Login_With_Invalid_Password(t *testing.T) {
	db, cleanup := test_db.CreateTestDatabase(t)
	defer cleanup()

	test_db.ExecFixtures(t, db, "base")

	query := test_graphql.GraphqlQuery{
		Query: loginGql,
		Variables: map[string]interface{}{
			"emailAddress": "test@example.com",
			"password":     "wrongPassword",
		},
	}
	var loginResult loginResult

	req := test_graphql.NewRequest(t, query)
	test_graphql.Handle(t, api.ResolverDependencies{Db: db}, req, &loginResult)
	test_graphql.RequireNoErrors(t, loginResult.GraphqlErrors)

	require.Equal(t, "invalidCredentials", loginResult.Data.Login.Error.Code, "data.login.error.code")
	assert.Empty(t, loginResult.Data.Login.Account.Id, "data.login.account.id")
	assert.Empty(t, loginResult.Data.Login.CsrfToken, "data.login.csrfToken")
}

func Test_Login_With_Unknown_EmailAddress(t *testing.T) {
	db, cleanup := test_db.CreateTestDatabase(t)
	defer cleanup()

	test_db.ExecFixtures(t, db, "base")

	query := test_graphql.GraphqlQuery{
		Query: loginGql,
		Variables: map[string]interface{}{
			"emailAddress": "not-exists@example.com",
			"password":     "somePassword",
		},
	}
	var loginResult loginResult

	req := test_graphql.NewRequest(t, query)
	test_graphql.Handle(t, api.ResolverDependencies{Db: db}, req, &loginResult)
	test_graphql.RequireNoErrors(t, loginResult.GraphqlErrors)

	require.Equal(t, "invalidCredentials", loginResult.Data.Login.Error.Code, "data.login.error.code")
	assert.Empty(t, loginResult.Data.Login.Account.Id, "data.login.account.id")
	assert.Empty(t, loginResult.Data.Login.CsrfToken, "data.login.csrfToken")
}

func Test_LoginStatus_With_Valid_Authentication(t *testing.T) {
	db, cleanup := test_db.CreateTestDatabase(t)
	defer cleanup()

	test_db.ExecFixtures(t, db, "base")

	query := test_graphql.GraphqlQuery{
		Query: loginGql,
		Variables: map[string]interface{}{
			"emailAddress": "test@example.com",
			"password":     "myRandomPassword",
		},
	}

	var loginResult loginResult

	req := test_graphql.NewRequest(t, query)
	w := test_graphql.Handle(t, api.ResolverDependencies{Db: db}, req, &loginResult)
	test_graphql.RequireNoErrors(t, loginResult.GraphqlErrors)

	cookie := readCookie(t, w, "authToken")

	query = test_graphql.GraphqlQuery{
		Query: loginStatusGql,
	}

	var loginStatusResult loginStatusResult

	req = test_graphql.NewRequest(t, query)
	req.Header.Set("X-CSRF-Token", loginResult.Data.Login.CsrfToken)
	req.AddCookie(cookie)
	test_graphql.Handle(t, api.ResolverDependencies{Db: db}, req, &loginStatusResult)
	test_graphql.RequireNoErrors(t, loginStatusResult.GraphqlErrors)

	require.True(t, loginStatusResult.Data.LoginStatus, "data.loginStatus")
}

func Test_LoginStatus_With_Invalid_Authentication(t *testing.T) {
	db, cleanup := test_db.CreateTestDatabase(t)
	defer cleanup()

	test_db.ExecFixtures(t, db, "base")

	query := test_graphql.GraphqlQuery{
		Query: loginStatusGql,
	}

	var loginStatusResult loginStatusResult

	req := test_graphql.NewRequest(t, query)
	// Invalid CSRF token generated with another secret
	req.Header.Set("X-CSRF-Token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NDUwNzcyNTR9.cZX4qrzVpbKJSoxBdlFsgnAq3fc8CwweD2cmITyNT9U")
	req.AddCookie(&http.Cookie{Name: "authToken", Value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NDUwNzcyNTQsImlhdCI6MTU0NTA1NTY1NCwicm9sZSI6IlN5c3RlbUFkbWluaXN0cmF0b3IiLCJzdWIiOiJiMmNlNDYwMi04ODI2LTQ5M2MtOWZkMS00OTI0MzMyNWEyY2UifQ.7rDHvp9W6aEc2rvylhCXiy_eA8kJkvT_FxY9UD8LxmM"})
	test_graphql.Handle(t, api.ResolverDependencies{Db: db}, req, &loginStatusResult)
	test_graphql.RequireNoErrors(t, loginStatusResult.GraphqlErrors)

	require.False(t, loginStatusResult.Data.LoginStatus, "data.loginStatus")
}

func Test_Token_Refresh(t *testing.T) {
	db, cleanup := test_db.CreateTestDatabase(t)
	defer cleanup()
	fixedTime := helper.FixedTime()
	test_db.ExecFixtures(t, db, "base")

	//
	// Login
	//

	query := test_graphql.GraphqlQuery{
		Query: loginGql,
		Variables: map[string]interface{}{
			"emailAddress": "test@example.com",
			"password":     "myRandomPassword",
		},
	}

	var loginResult loginResult

	req := test_graphql.NewRequest(t, query)
	w := test_graphql.Handle(t, api.ResolverDependencies{Db: db, TimeSource: fixedTime}, req, &loginResult)
	test_graphql.RequireNoErrors(t, loginResult.GraphqlErrors)

	cookie := readCookie(t, w, "authToken")
	csrfToken := loginResult.Data.Login.CsrfToken

	//
	// 30 Minutes later...
	//

	// Add time after login so the refresh threshold gets triggered
	fixedTime = fixedTime.Add(30 * time.Minute)

	query = test_graphql.GraphqlQuery{
		Query: loginStatusGql,
	}

	var loginStatusResult loginStatusResult

	req = test_graphql.NewRequest(t, query)
	req.Header.Set("X-CSRF-Token", csrfToken)
	req.AddCookie(cookie)
	w = test_graphql.Handle(t, api.ResolverDependencies{Db: db, TimeSource: fixedTime}, req, &loginStatusResult)
	test_graphql.RequireNoErrors(t, loginStatusResult.GraphqlErrors)

	require.True(t, loginStatusResult.Data.LoginStatus, "data.loginStatus")

	//
	// Refresh tokens should be "pushed"
	//

	refreshedCsrfToken := w.Header().Get("X-Refresh-CSRF-Token")
	assert.NotEmpty(t, refreshedCsrfToken, "X-Refresh-CSRF-Token header")
	assert.NotEqual(t, csrfToken, refreshedCsrfToken, "refreshed CSRF token")
	refreshedAuthTokenCookie := readCookie(t, w, "authToken")
	assert.NotEmpty(t, refreshedAuthTokenCookie.Value, "'authToken' cookie value")
	assert.NotEqual(t, cookie.Value, refreshedAuthTokenCookie.Value, " refreshed 'authToken' cookie")

	//
	// Test if refreshed tokens could be used
	//

	req = test_graphql.NewRequest(t, query)
	req.Header.Set("X-CSRF-Token", refreshedCsrfToken)
	req.AddCookie(refreshedAuthTokenCookie)
	test_graphql.Handle(t, api.ResolverDependencies{Db: db, TimeSource: fixedTime}, req, &loginStatusResult)
	test_graphql.RequireNoErrors(t, loginStatusResult.GraphqlErrors)

	assert.True(t, loginStatusResult.Data.LoginStatus, "data.loginStatus")
}

// This is important to verify, since CSRF is skipped for GET requests by default
func Test_Mutation_With_Get_Fails(t *testing.T) {
	db, cleanup := test_db.CreateTestDatabase(t)
	defer cleanup()

	test_db.ExecFixtures(t, db, "base")

	query := test_graphql.GraphqlQuery{
		Query: logoutGql,
	}

	req := test_graphql.NewRequest(t, query)
	req.Method = http.MethodGet
	test_auth.ApplyFixedAuthValuesSystemAdministrator(req)

	graphqlHandler := api_handler.NewGraphqlHandler(api.ResolverDependencies{
		Db:         db,
		TimeSource: helper.FixedTime(),
	}, api_handler.HandlerConfig{})
	w := httptest.NewRecorder()
	graphqlHandler.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnprocessableEntity, w.Code, "response status")
}

func readCookie(t *testing.T, w *httptest.ResponseRecorder, cookieName string) *http.Cookie {
	request := &http.Request{Header: http.Header{"Cookie": w.HeaderMap["Set-Cookie"]}}
	cookie, err := request.Cookie("authToken")
	require.NoError(t, err, "failed to read %q cookie: %v", cookieName, err)

	return cookie
}
