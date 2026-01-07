package authorization

import (
	"errors"
	"fmt"
)

type Error interface {
	error
	AuthorizationCause() string
}

func Fail(cause string) Error {
	return authorizationError{cause: cause}
}

type authorizationError struct {
	cause string
	err   error
}

var _ Error = authorizationError{}

func (e authorizationError) Error() string {
	return fmt.Sprintf("not authorized: %s", e.unprefixedError())
}

func (e authorizationError) unprefixedError() string {
	if e.err != nil {
		var authErr authorizationError
		if errors.As(e.err, &authErr) {
			return fmt.Sprintf("%s: %s", e.cause, authErr.unprefixedError())
		}
		return fmt.Sprintf("%s: %v", e.cause, e.err)
	}
	return e.cause
}

func (e authorizationError) Unwrap() error {
	return e.err
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
