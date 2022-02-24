package graphql

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func RequireNoErrors(t *testing.T, errs GraphqlErrors) {
	t.Helper()

	if len(errs.Errors) > 0 {
		t.Error("Unexpected GraphQL errors:")
		for _, err := range errs.Errors {
			var typeSuffix string
			if err.Extensions.Type != "" {
				typeSuffix = fmt.Sprintf(" (type: %q)", err.Extensions.Type)
			}
			t.Errorf("%v %s%s", err.Path, err.Message, typeSuffix)
		}
		t.FailNow()
	}
}

func RequireErrors(t *testing.T, errs GraphqlErrors) {
	t.Helper()

	if len(errs.Errors) == 0 {
		t.Error("Expected GraphQL errors, but got none")
		t.FailNow()
	}
}

func RequireNotAuthorizedError(t *testing.T, errs GraphqlErrors) {
	t.Helper()

	if len(errs.Errors) != 1 {
		t.Errorf("Expected 1 GraphQL error, but got %d", len(errs.Errors))
		t.FailNow()
	}

	firstError := errs.Errors[0]
	require.Equal(t, "notAuthorized", firstError.Extensions.Type)
}