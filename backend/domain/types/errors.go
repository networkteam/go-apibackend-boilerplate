package types

import (
	"errors"
	"fmt"
)

var ErrEnumsMustBeStrings = errors.New("enums must be strings")

type FieldResolvableError interface {
	error
	FieldPath() []string
	ErrorArguments() []string
	ErrorCode() string
	Is(error) bool
}

type FieldError struct {
	// Field name to identify the field for the error.
	// Can be omitted for anonymous errors that can be wrapped later with a specific field name.
	Field     string
	Code      string
	Arguments []string
}

func (n FieldError) FieldPath() []string {
	if n.Field == "" {
		return nil
	}
	return []string{n.Field}
}

func (n FieldError) ErrorArguments() []string {
	return n.Arguments
}

func (n FieldError) Error() string {
	if n.Field == "" {
		return n.Code
	}
	return fmt.Sprintf("field %s: %s", n.Field, n.Code)
}

func (n FieldError) ErrorCode() string {
	return n.Code
}

func (n FieldError) Is(err error) bool {
	return fieldResolvableErrorIs(n, err)
}

func fieldResolvableErrorIs(fieldErr FieldResolvableError, err error) bool {
	newErr, ok := err.(FieldResolvableError) //nolint:errorlint
	if !ok {
		return false
	}
	if !stringsEqual(fieldErr.FieldPath(), newErr.FieldPath()) {
		return false
	}
	if fieldErr.ErrorCode() != newErr.ErrorCode() {
		return false
	}
	if !stringsEqual(fieldErr.ErrorArguments(), newErr.ErrorArguments()) {
		return false
	}
	return true
}

func (n FieldError) Extensions() map[string]any {
	return map[string]any{
		"type":      "validationFailed",
		"code":      n.Code,
		"field":     n.Field,
		"arguments": n.Arguments,
	}
}

var _ FieldResolvableError = &FieldError{}

type nestedFieldError struct {
	field string
	err   FieldResolvableError
}

func (n nestedFieldError) Error() string {
	if n.field == "" {
		return n.err.Error()
	}
	return fmt.Sprintf("field %s: %s", n.field, n.err.Error())
}

func (n nestedFieldError) FieldPath() []string {
	return append([]string{n.field}, n.err.FieldPath()...)
}

func (n nestedFieldError) ErrorArguments() []string {
	return n.err.ErrorArguments()
}

func (n nestedFieldError) ErrorCode() string {
	return n.err.ErrorCode()
}

func (n nestedFieldError) Is(err error) bool {
	return fieldResolvableErrorIs(n, err)
}

var _ FieldResolvableError = &nestedFieldError{}

func WrapFieldError(err FieldResolvableError, fieldName string) FieldResolvableError {
	return &nestedFieldError{
		field: fieldName,
		err:   err,
	}
}

func stringsEqual(s1 []string, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i, s := range s1 {
		if s != s2[i] {
			return false
		}
	}
	return true
}
