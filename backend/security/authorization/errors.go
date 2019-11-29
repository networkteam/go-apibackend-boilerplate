package authorization

import "fmt"

type AuthorizationError interface {
	error
	AuthorizationCause() string
}

type authorizationError struct {
	cause string
}

var _ AuthorizationError = authorizationError{}

func (e authorizationError) Error() string {
	return fmt.Sprintf("not authorized: %v", e.cause)
}

func (e authorizationError) AuthorizationCause() string {
	return e.cause
}

// Implements graphql.ExtendedError
func (e authorizationError) Extensions() map[string]interface{} {
	return map[string]interface{}{
		"type":  "notAuthorized",
		"cause": e.cause,
	}
}

func NewAuthorizationError(cause string) AuthorizationError {
	return authorizationError{cause}
}
