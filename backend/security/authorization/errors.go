package authorization

import "fmt"

type Error interface {
	error
	AuthorizationCause() string
}

type authorizationError struct {
	cause string
}

var _ Error = authorizationError{}

func (e authorizationError) Error() string {
	return fmt.Sprintf("not authorized: %v", e.cause)
}

func (e authorizationError) AuthorizationCause() string {
	return e.cause
}

// Extensions implements graphql.ExtendedError
func (e authorizationError) Extensions() map[string]any {
	return map[string]any{
		"type":  "notAuthorized",
		"cause": e.cause,
	}
}
