package api

import (
	"errors"

	"myvendor.mytld/myproject/backend/api/graph/model"
	"myvendor.mytld/myproject/backend/domain"
)

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

func (err TypedError) Code() string {
	return err.errorType
}

// Extensions implements graphql.ExtendedError
func (err TypedError) Extensions() map[string]interface{} {
	return map[string]interface{}{
		"type": err.errorType,
	}
}

func ResultFromErr(err error) (*model.Result, error) {
	if fieldsError := FieldsErrorFromErr(err); fieldsError != nil {
		return &model.Result{
			Error: fieldsError,
		}, nil
	}
	return nil, err
}

func FieldsErrorFromErr(err error) *model.FieldsError {
	var fieldErr domain.FieldResolvableError
	if errors.As(err, &fieldErr) {
		return &model.FieldsError{
			Errors: []*model.FieldError{
				{
					Path:      fieldErr.FieldPath(),
					Code:      fieldErr.ErrorCode(),
					Arguments: fieldErr.ErrorArguments(),
				},
			},
		}
	}
	return nil
}
