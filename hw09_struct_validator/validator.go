package hw09structvalidator

import (
	"errors"
	"reflect"
	"strings"
)

func Validate(v any) ValidationErrors {
	if v == nil {
		return nil
	}

	rootStruct := reflect.ValueOf(v)
	rootType := rootStruct.Type()

	if rootType.Kind() != reflect.Struct {
		return nil
	}

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
	return nil
}

func ValidateFieldValueSlice(values []any, tagValue string) []error {
	validatingErrors := make([]error, 0)

	for _, value := range values {
		if errors := ValidateFieldValue(value, tagValue); errors != nil {
			validatingErrors = append(validatingErrors, errors...)
		}
	}

	return validatingErrors
}

func ValidateFieldValue(value any, tagValue string) []error {
	tagsRaw := strings.Split(tagValue, "|")

	if len(tagsRaw) == 0 {
		return nil
	}

	validatingErrors := make([]error, 0)

	for _, tagRaw := range tagsRaw {
		splitedTag := strings.Split(tagRaw, ":")

		if len(splitedTag) != 2 {
			validatingErrors = append(validatingErrors, errors.New("Too many ':' in tag "+tagRaw))
			continue
		}

		validatorName := splitedTag[0]
		validatorValue := splitedTag[1]

		if err := executeValidation(value, validatorName, validatorValue); err != nil {
			validatingErrors = append(validatingErrors, err)
		}
	}

	return validatingErrors
}

func executeValidation(value any, validatorName, validatorValue string) error {
	switch value := value.(type) {
	case int, int8, int16, int32, int64:
		return executeValidationForInt(value, validatorName, validatorValue)
	case string:
		return executeValidationForString(value, validatorName, validatorValue)
	default:
		return nil
	}
}

func executeValidationForInt(value any, validatorName, validatorValue string) error {
	validateIntFunc, err := getValidateIntFunc(validatorName)
	if err != nil {
		return err
	}
	integer, err := anyIntToInt64(value)
	if err != nil {
		return err
	}

	return validateIntFunc(integer, validatorValue)
}

func executeValidationForString(value string, validatorName, validatorValue string) error {
	validateIntFunc, err := getValidateStringFunc(validatorName)
	if err != nil {
		return err
	}

	return validateIntFunc(value, validatorValue)
}

func anyIntToInt64(value any) (int64, error) {
	switch value := value.(type) {
	case int:
		return int64(value), nil
	case int8:
		return int64(value), nil
	case int16:
		return int64(value), nil
	case int32:
		return int64(value), nil
	case int64:
		return value, nil
	default:
		return 0, errors.New("unsupported type")
	}
}
