package domain

import "fmt"

type FieldResolvableError interface {
	error
	FieldPath() []string
	ErrorArguments() []string
	ErrorCode() string
}

type FieldError struct {
	Field     string
	Code      string
	Arguments []string
}

func (n FieldError) FieldPath() []string {
	return []string{n.Field}
}

func (n FieldError) ErrorArguments() []string {
	return n.Arguments
}

func (n FieldError) Error() string {
	return fmt.Sprintf("for field: %s: %s", n.Field, n.Code)
}

func (n FieldError) ErrorCode() string {
	return n.Code
}

var _ FieldResolvableError = new(FieldError)
