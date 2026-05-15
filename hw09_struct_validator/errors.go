package hw09structvalidator

import (
	"fmt"
	"strings"
)

type ValidationError struct {
	Field string
	Err   error
}

type (
	ValidationErrors []ValidationError
)

func (v ValidationErrors) Error() string {
	stringBuilder := strings.Builder{}

	for _, err := range v {
		fmt.Fprintf(&stringBuilder, "field: %s, error: %s", err.Field, err.Err.Error())
	}

	return stringBuilder.String()
}

func NewValidationError(field string, err error) ValidationError {
	return ValidationError{
		Field: field,
		Err:   err,
	}
}

type TagDeclarationError struct {
	Msg string
	Err error
}

func (tagDeclarationError *TagDeclarationError) Error() string { return tagDeclarationError.Msg }

func (tagDeclarationError *TagDeclarationError) Unwrap() error { return tagDeclarationError.Err }

type InvalidValueError struct {
	Msg string
}

func (invalidValueError *InvalidValueError) Error() string { return invalidValueError.Msg }
