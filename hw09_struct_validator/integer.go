package hw09structvalidator

import (
	"fmt"
	"strconv"
	"strings"
)

type validateIntFunc func(value int64, tagValue string) error

func getValidateIntFunc(validatorName string) (validateIntFunc, error) {
	switch validatorName {
	case "min":
		return validateMin, nil
	case "max":
		return validateMax, nil
	case "in":
		return validateIn, nil
	default:
		return nil, &TagDeclarationError{Msg: "unsupported validator: " + validatorName}
	}
}

func validateMin(value int64, tagValue string) error {
	minValue, err := strconv.ParseInt(tagValue, 10, 64)
	if err != nil {
		return &TagDeclarationError{Msg: "failed to parse min value", Err: err}
	}

	if value < minValue {
		return &InvalidValueError{Msg: fmt.Sprintf("value must be >= %d", minValue)}
	}

	return nil
}

func validateMax(value int64, tagValue string) error {
	maxValue, err := strconv.ParseInt(tagValue, 10, 64)
	if err != nil {
		return &TagDeclarationError{Msg: "failed to parse max value", Err: err}
	}

	if value > maxValue {
		return &InvalidValueError{Msg: fmt.Sprintf("value must be <= %d", maxValue)}
	}

	return nil
}

func validateIn(value int64, tagValue string) error {
	bordersSplited := strings.Split(tagValue, ",")

	if len(bordersSplited) != 2 {
		return &TagDeclarationError{Msg: "invalid format for tag 'in': " + tagValue}
	}

	minValue, err := strconv.ParseInt(bordersSplited[0], 10, 64)
	if err != nil {
		return &TagDeclarationError{Msg: "failed to parse min value", Err: err}
	}

	maxValue, err := strconv.ParseInt(bordersSplited[1], 10, 64)
	if err != nil {
		return &TagDeclarationError{Msg: "failed to parse max value", Err: err}
	}

	if value < minValue || value > maxValue {
		return &InvalidValueError{Msg: fmt.Sprintf("value expected to be in range [%d:%d] but was %d", minValue, maxValue, value)}
	}

	return nil
}
