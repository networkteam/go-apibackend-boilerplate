package auth

import (
	"context"
	"net/http"

	"github.com/gofrs/uuid"

	"myvendor/myproject/backend/domain"
	"myvendor/myproject/backend/security/authentication"
)

// Auth token for SystemAdministrator
var FixedCookieValueSystemAdministrator = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NDUwNzkzNzcsImlhdCI6MTU0NTA1Nzc3Nywicm9sZSI6IlN5c3RlbUFkbWluaXN0cmF0b3IiLCJzdWIiOiIwMzUyMzI0Yy0yNWFhLTRkZWYtOTM1ZC0wZWVkOTk5ZjFmOTkifQ.7yE39P_xZGPvOzKDuNXy5qlU9WilfbA_x1xj-a9C7-E"
var FixedCsrfTokenSystemAdministrator = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NDUwNzkzNzd9.lobqm1XgilwvA9kpamBw6l5HiUf1jgpjg8ngWNy7HK4"

// Auth token for OrganisationAdministrator for organisation networkteam (7bc9e6b5-9cc0-435f-9ad5-8c4f482f0c45)
var FixedCookieValueOrganisationAdministrator = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NDUwNzkzNzcsImlhdCI6MTU0NTA1Nzc3Nywicm9sZSI6Ik9yZ2FuaXNhdGlvbkFkbWluaXN0cmF0b3IiLCJzdWIiOiIwMzUyMzI0Yy0yNWFhLTRkZWYtOTM1ZC0wZWVkOTk5ZjFmYTAifQ.QRPKDg46ghSx5vxl1AnZLm9VfuymjLH9sbicL13hxqk"
var FixedCsrfTokenOrganisationAdministrator = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NDUwNzkzNzd9.lobqm1XgilwvA9kpamBw6l5HiUf1jgpjg8ngWNy7HK4"

// app account 1dcec63f-5d0c-4abf-9d4c-f707632a1a73
var FixedAuthTokenOperatorAccount = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MDgxMjk3NzcsImlhdCI6MTU0NTA1Nzc3Nywib3JnYW5pc2F0aW9uSWQiOiI3YmM5ZTZiNS05Y2MwLTQzNWYtOWFkNS04YzRmNDgyZjBjNDUiLCJyb2xlIjoiT3BlcmF0b3IiLCJzdWIiOiIxZGNlYzYzZi01ZDBjLTRhYmYtOWQ0Yy1mNzA3NjMyYTFhNzMifQ.8ijrEfXQjIQIwckW695Zm5cOfqn184WVBJ3TRWitRW0"

func ApplyFixedAuthValuesSystemAdministrator(req *http.Request) {
	req.Header.Set("X-CSRF-Token", FixedCsrfTokenSystemAdministrator)
	cookie := http.Cookie{
		Name:  "authToken",
		Value: FixedCookieValueSystemAdministrator,
	}
	req.AddCookie(&cookie)
}

func ApplyFixedAuthValuesOrganisationAdministrator(req *http.Request) {
	req.Header.Set("X-CSRF-Token", FixedCsrfTokenOrganisationAdministrator)
	cookie := http.Cookie{
		Name:  "authToken",
		Value: FixedCookieValueOrganisationAdministrator,
	}
	req.AddCookie(&cookie)
}

func ApplyFixedAuthValuesOperator(req *http.Request) {
	req.Header.Set("Authorization", FixedAuthTokenOperatorAccount)
}

func GetContextWithOrganisationAdministrator() context.Context {
	ctx := context.Background()

	organisationID := uuid.FromStringOrNil("7bc9e6b5-9cc0-435f-9ad5-8c4f482f0c45")
	return authentication.WithAuthContext(ctx, authentication.AuthContext{
		Authenticated:  true,
		OrganisationID: &organisationID,
		Role:           domain.RoleOrganisationAdministrator,
	})
}
