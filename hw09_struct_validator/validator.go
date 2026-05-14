package hw09structvalidator

import (
	"reflect"
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	panic("implement me")
}

func Validate(v any) ValidationErrors {
	if v == nil {
		return nil
	}

	rootStruct := reflect.ValueOf(v)
	rootType := rootStruct.Type()

	for i := 0; i < rootStruct.NumField(); i++ {
		field := rootStruct.Field(i)
		fieldType := rootType.Field(i)

		tagValue := fieldType.Tag.Get("validate")
		if tagValue == "" {
			continue
		}

		switch field.Kind() {
		case reflect.Struct:
			continue
		case reflect.String:
			continue
		case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
			continue
		default:
			continue
		}

	}

	// validatingErrors := make(ValidationErrors, 0)

	//if err != nil {
	//	validatingErrors = append(validatingErrors, ValidationError{Field: "v", Err: err})
	//	return validatingErrors
	//}
	return nil
}
