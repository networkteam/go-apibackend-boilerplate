package api

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"myvendor.mytld/myproject/backend/api/graph/public/model"
	"myvendor.mytld/myproject/backend/domain/types"
)

var ErrAccessTokenInvalid = TypedError{"accessTokenInvalid", "access token invalid"}
var ErrAccessTokenExpired = TypedError{"accessTokenExpired", "access token expired"}
var ErrAuthenticationFailed = TypedError{"authenticationFailed", "authentication failed"}
var ErrAuthenticationRequired = TypedError{"authenticationRequired", "authentication required"}
var ErrCsrfTokenMissing = TypedError{"csrfTokenMissing", "CSRF token missing"}
var ErrCsrfTokenInvalid = TypedError{"csrfTokenInvalid", "CSRF token invalid"}
var ErrCsrfTokenExpired = TypedError{"csrfTokenExpired", "CSRF token expired"}

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

func FieldsErrorFromErr(err error, opt ...FieldsErrorFromErrOpt) *model.FieldsError {
	var remapFieldPaths []FieldPathMapping
	if len(opt) > 0 {
		remapFieldPaths = opt[0].RemapFieldPaths
	}

	var fieldErr types.FieldResolvableError
	if errors.As(err, &fieldErr) {
		path := fieldErr.FieldPath()
		for _, mapping := range remapFieldPaths {
			if slices.Equal(path, mapping.From) {
				path = mapping.To
				break
			}
		}

		return &model.FieldsError{
			Errors: []*model.FieldError{
				{
					Path:      path,
					Code:      fieldErr.ErrorCode(),
					Arguments: fieldErr.ErrorArguments(),
				},
			},
		}
	}
	return nil
}

type FieldsErrorFromErrOpt struct {
	RemapFieldPaths []FieldPathMapping
}

type FieldPathMapping struct {
	From []string
	To   []string
}

type fieldPathError struct {
	path      []string
	code      string
	arguments []string
}

func (f fieldPathError) Error() string {
	if len(f.path) == 0 {
		return f.code
	}
	return fmt.Sprintf("field %s: %s", strings.Join(f.path, "."), f.code)
}

func (f fieldPathError) FieldPath() []string {
	return f.path
}

func (f fieldPathError) ErrorArguments() []string {
	return f.arguments
}

func (f fieldPathError) ErrorCode() string {
	return f.code
}

func (f fieldPathError) Is(err error) bool {
	newErr, ok := err.(types.FieldResolvableError)
	if !ok {
		return false
	}
	if !slices.Equal(f.path, newErr.FieldPath()) {
		return false
	}
	if f.code != newErr.ErrorCode() {
		return false
	}
	if !slices.Equal(f.arguments, newErr.ErrorArguments()) {
		return false
	}
	return true
}

func (f fieldPathError) Extensions() map[string]any {
	return map[string]any{
		"type":      "validationFailed",
		"code":      f.code,
		"field":     strings.Join(f.FieldPath(), "."),
		"arguments": f.arguments,
	}
}

func RemapFieldResolvableError(err error, mapping FieldPathMapping) error {
	var fieldErr types.FieldResolvableError
	if errors.As(err, &fieldErr) {
		path := fieldErr.FieldPath()
		if slices.Equal(path, mapping.From) {
			return fieldPathError{
				path:      mapping.To,
				code:      fieldErr.ErrorCode(),
				arguments: fieldErr.ErrorArguments(),
			}
		}
	}
	return err
}
