package hw09structvalidator

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type validateStringFunc func(value string, tagValue string) error

func getValidateStringFunc(validatorName string) (validateStringFunc, error) {
	switch validatorName {
	case "len":
		return validateLen, nil
	case "regexp":
		return validateRegExp, nil
	case "in":
		return validateInString, nil
	default:
		return nil, &TagDeclarationError{Msg: "unsupported validator: " + validatorName}
	}
}

func validateLen(value string, tagValue string) error {
	expectedLength, err := strconv.ParseInt(tagValue, 10, 64)
	if err != nil {
		return &TagDeclarationError{Msg: "failed to parse length value", Err: err}
	}

	if len(value) != int(expectedLength) {
		return &InvalidValueError{Msg: fmt.Sprintf("expected length of string %d but was %d", expectedLength, len(value))}
	}

	return nil
}

func validateRegExp(value string, tagValue string) error {
	matched, err := regexp.MatchString(tagValue, value)
	if err != nil {
		return &TagDeclarationError{Msg: "failed to compile regexp", Err: err}
	}

	if !matched {
		return &InvalidValueError{Msg: "regexp is not matched"}
	}

	return nil
}

func validateInString(value string, tagValue string) error {
	expectedValues := strings.Split(tagValue, ",")

	if !slices.Contains(expectedValues, value) {
		return &InvalidValueError{Msg: "forbidden value"}
	}

	return nil
}
