package validators

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type validateIntFunc func (value int64, tagValue string) error

func getValidateIntFunc(validatorName string) (validateIntFunc, error) {
	switch validatorName {
		case "min":
			return validateMin, nil
		case "max":
			return validateMax, nil
		case "in":
			return validateIn, nil
		default:
			return nil, errors.New("unsupported validator: "+validatorName)
		}
}

func validateMin(value int64, tagValue string) error {
	minValue, err := strconv.ParseInt(tagValue, 10, 64)

	if err != nil {
		return err
	}

	if value < minValue {
		return fmt.Errorf("value must be >= %d", value)
	}

	return nil
}

func validateMax(value int64, tagValue string) error {
	maxValue, err := strconv.ParseInt(tagValue, 10, 64)

	if err != nil {
		return err
	}

	if value > maxValue {
		return fmt.Errorf("value must be <= %d", value)
	}

	return nil
}

func validateIn(value int64, tagValue string) error {
	bordersSplited := strings.Split(tagValue, ",")

	if len(bordersSplited) != 2 {
		return errors.New("invalid format for tag 'in': " + tagValue)
	}

	minValue, err := strconv.ParseInt(bordersSplited[0], 10, 64)
	if err != nil {
		return err
	}

	maxValue, err := strconv.ParseInt(bordersSplited[1], 10, 64)
	if err != nil {
		return err
	}

	if value < minValue || value > maxValue {
		return fmt.Errorf("value expected to be in range [%d:%d] but was %d", minValue, maxValue, value)
	}

	return nil
}
