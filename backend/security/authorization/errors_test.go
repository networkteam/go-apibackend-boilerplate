//nolint:testpackage
package authorization

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	err := authorizationError{
		cause: "test",
		err:   nil,
	}
	if err.Error() != "not authorized: test" {
		t.Errorf("Expected 'not authorized: test', got '%s'", err.Error())
	}
	if err.AuthorizationCause() != "test" {
		t.Errorf("Expected 'test', got '%s'", err.AuthorizationCause())
	}
	if err.Unwrap() != nil {
		t.Errorf("Expected nil, got '%v'", err.Unwrap())
	}

	{
		var authErr Error
		if !errors.As(err, &authErr) {
			t.Errorf("Expected to be able to assert to Error interface")
		}
		if !errors.Is(err, authErr) {
			t.Errorf("AuthorizationError should be equal to itself")
		}
	}

	{
		var emptyErr Error
		if errors.Is(err, emptyErr) {
			t.Errorf("AuthorizationError should not be equal to empty error")
		}
	}

	{
		authErr := authorizationError{cause: "requires role [User]"}
		nestedAuthErr := authorizationError{cause: "no check satisfied", err: authErr}

		assert.Equal(t, "not authorized: no check satisfied: requires role [User]", nestedAuthErr.Error())
	}
}
