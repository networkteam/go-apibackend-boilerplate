package api

import "myvendor/myproject/backend/domain"

var ErrAuthTokenInvalid = TypedError{"authTokenInvalid", "auth token invalid"}
var ErrAuthTokenExpired = TypedError{"authTokenExpired", "auth token expired"}
var ErrAuthenticationRequired = TypedError{"authenticationRequired", "authentication required"}
var ErrCsrfTokenMissing = TypedError{"csrfTokenMissing", "CSRF token missing"}
var ErrCsrfTokenInvalid = TypedError{"csrfTokenInvalid", "CSRF token invalid"}

type TypedError struct {
	errorType string
	error     string
}

func (err TypedError) Error() string {
	return err.error
}

// Implements graphql.ExtendedError
func (err TypedError) Extensions() map[string]interface{} {
	return map[string]interface{}{
		"type": err.errorType,
	}
}

func CreateResultFromErr(err error) (CreateResult, error) {
	if fieldsError := FieldsErrorFromErr(err); fieldsError != nil {
		return CreateResult{
			Error: fieldsError,
		}, nil
	}
	return CreateResult{}, err
}

func ResultFromErr(err error) (Result, error) {
	if fieldsError := FieldsErrorFromErr(err); fieldsError != nil {
		return Result{
			Error: fieldsError,
		}, nil
	}
	return Result{}, err
}

func FieldsErrorFromErr(err error) *FieldsError {
	if f, ok := err.(domain.FieldResolvableError); ok {
		return &FieldsError{
			Errors: []*FieldError{
				{
					Path:      f.FieldPath(),
					Code:      f.ErrorCode(),
					Arguments: f.ErrorArguments(),
				},
			},
		}
	}
	return nil
}
