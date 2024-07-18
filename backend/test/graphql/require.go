package graphql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func RequireNoErrors(t *testing.T, errs GraphqlErrors) {
	t.Helper()

	if len(errs.Errors) > 0 {
		t.Errorf("Unexpected GraphQL errors:\n%s", errs)
		t.FailNow()
	}
}

func RequireErrors(t *testing.T, errs GraphqlErrors, expectedErrors ...GraphqlError) {
	t.Helper()

	if len(errs.Errors) == 0 {
		t.Error("Expected GraphQL errors, but got none")
		t.FailNow()
	}

	if expectedErrors != nil && len(expectedErrors) != len(errs.Errors) {
		t.Errorf("Expected %d GraphQL errors, but got %d:\n%s", len(expectedErrors), len(errs.Errors), errs)
		t.FailNow()
	}

	for _, expectedError := range expectedErrors {
		if !graphqlErrorMatches(expectedError, errs.Errors) {
			t.Errorf("Expected GraphQL error %s, but got:\n%s", expectedError, errs)
			t.FailNow()
		}
	}
}

func graphqlErrorMatches(expectedErr GraphqlError, errs []GraphqlError) bool {
	for _, err := range errs {
		if expectedErr.Message != "" && err.Message != expectedErr.Message {
			continue
		}

		expectedExtensions := expectedErr.Extensions
		if expectedExtensions.Field != "" && err.Extensions.Field != expectedExtensions.Field {
			continue
		}
		if expectedExtensions.Code != "" && err.Extensions.Code != expectedExtensions.Code {
			continue
		}
		if expectedExtensions.Type != "" && err.Extensions.Type != expectedExtensions.Type {
			continue
		}

		return true
	}

	return false
}

func RequireNotAuthorizedError(t *testing.T, errs GraphqlErrors) {
	t.Helper()
	requireGraphqlErrorType(t, errs, "notAuthorized")
}

func RequireAuthTokenExpiredError(t *testing.T, errs GraphqlErrors) {
	t.Helper()
	requireGraphqlErrorType(t, errs, "authTokenExpired")
}

func RequireAuthenticationRequiredError(t *testing.T, errs GraphqlErrors) {
	t.Helper()
	requireGraphqlErrorType(t, errs, "authenticationRequired")
}

func requireGraphqlErrorType(t *testing.T, errs GraphqlErrors, errorType string) {
	t.Helper()

	if len(errs.Errors) != 1 {
		t.Errorf("Expected 1 GraphQL error, but got %d:\n%s", len(errs.Errors), errs)
		t.FailNow()
	}

	firstError := errs.Errors[0]
	if firstError.Extensions.Type == "" {
		t.Errorf("Expected GraphQL error extension type %s, but got generic error: %s", errorType, firstError)
		t.FailNow()
	}
	if firstError.Extensions.Type != errorType {
		t.Errorf("Expected GraphQL error extension type %s, but got: %s", errorType, firstError.Extensions.Type)
		t.FailNow()
	}
}

func AssertFieldError(t *testing.T, err *FieldsError, expectedCode string, expectedPath []string) bool {
	t.Helper()

	if assert.NotNil(t, err, "result.error") {
		if assert.Len(t, err.Errors, 1, "result.error.errors") {
			return assert.Equal(t, expectedCode, err.Errors[0].Code) &&
				assert.Equal(t, expectedPath, err.Errors[0].Path)
		}
	}
	return false
}
