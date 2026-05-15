package hw09structvalidator

import (
	"errors"
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
		return nil, errors.New("unsupported validator: " + validatorName)
	}
}

func validateLen(value string, tagValue string) error {
	expectedLength, err := strconv.ParseInt(tagValue, 10, 64)
	if err != nil {
		return err
	}

	if len(value) != int(expectedLength) {
		return fmt.Errorf("expected length of string %d but was %d", expectedLength, len(value))
	}

	return nil
}

func validateRegExp(value string, tagValue string) error {
	matched, err := regexp.MatchString(tagValue, value)
	if err != nil {
		return err
	}

	if !matched {
		return errors.New("regexp is not matched")
	}

	return nil
}

func validateInString(value string, tagValue string) error {
	expectedValues := strings.Split(tagValue, ",")

	if !slices.Contains(expectedValues, value) {
		return errors.New("forbidden value")
	}

	return nil
}
