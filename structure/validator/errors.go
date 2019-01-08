package validator

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tidepool-org/platform/errors"
)

const (
	ErrorCodeValueNotExists       = "value-not-exists"
	ErrorCodeValueExists          = "value-exists"
	ErrorCodeValueNotEmpty        = "value-not-empty"
	ErrorCodeValueEmpty           = "value-empty"
	ErrorCodeValueDuplicate       = "value-duplicate"
	ErrorCodeValueNotTrue         = "value-not-true"
	ErrorCodeValueNotFalse        = "value-not-false"
	ErrorCodeValueOutOfRange      = "value-out-of-range"
	ErrorCodeValueDisallowed      = "value-disallowed"
	ErrorCodeValueNotAllowed      = "value-not-allowed"
	ErrorCodeValueMatches         = "value-matches"
	ErrorCodeValueNotMatches      = "value-not-matches"
	ErrorCodeValueNotAfter        = "value-not-after"
	ErrorCodeValueNotBefore       = "value-not-before"
	ErrorCodeValueNotValid        = "value-not-valid"
	ErrorCodeValuesNotExistForAny = "values-not-exist-for-any"
	ErrorCodeValuesNotExistForOne = "values-not-exist-for-one"
	ErrorCodeLengthOutOfRange     = "length-out-of-range"
)

func ErrorValueNotExists() error {
	return errors.Prepared(ErrorCodeValueNotExists, "value does not exist", "value does not exist")
}

func ErrorValueExists() error {
	return errors.Prepared(ErrorCodeValueExists, "value exists", "value exists")
}

func ErrorValueNotEmpty() error {
	return errors.Prepared(ErrorCodeValueNotEmpty, "value is not empty", "value is not empty")
}

func ErrorValueEmpty() error {
	return errors.Prepared(ErrorCodeValueEmpty, "value is empty", "value is empty")
}

func ErrorValueDuplicate() error {
	return errors.Preparedf(ErrorCodeValueDuplicate, "value is a duplicate", "value is a duplicate")
}

func ErrorValueBoolNotTrue() error {
	return errors.Prepared(ErrorCodeValueNotTrue, "value is not true", "value is not true")
}

func ErrorValueBoolNotFalse() error {
	return errors.Prepared(ErrorCodeValueNotFalse, "value is not false", "value is not false")
}

func ErrorValueNotEqualTo(value interface{}, limit interface{}) error {
	return errors.Preparedf(ErrorCodeValueOutOfRange, "value is out of range", "value %s is not equal to %s", stringify(value), stringify(limit))
}

func ErrorValueEqualTo(value interface{}, limit interface{}) error {
	return errors.Preparedf(ErrorCodeValueOutOfRange, "value is out of range", "value %s is equal to %s", stringify(value), stringify(limit))
}

func ErrorValueNotLessThan(value interface{}, limit interface{}) error {
	return errors.Preparedf(ErrorCodeValueOutOfRange, "value is out of range", "value %s is not less than %s", stringify(value), stringify(limit))
}

func ErrorValueNotLessThanOrEqualTo(value interface{}, limit interface{}) error {
	return errors.Preparedf(ErrorCodeValueOutOfRange, "value is out of range", "value %s is not less than or equal to %s", stringify(value), stringify(limit))
}

func ErrorValueNotGreaterThan(value interface{}, limit interface{}) error {
	return errors.Preparedf(ErrorCodeValueOutOfRange, "value is out of range", "value %s is not greater than %s", stringify(value), stringify(limit))
}

func ErrorValueNotGreaterThanOrEqualTo(value interface{}, limit interface{}) error {
	return errors.Preparedf(ErrorCodeValueOutOfRange, "value is out of range", "value %s is not greater than or equal to %s", stringify(value), stringify(limit))
}

func ErrorValueNotInRange(value interface{}, lowerLimit interface{}, upperLimit interface{}) error {
	return errors.Preparedf(ErrorCodeValueOutOfRange, "value is out of range", "value %s is not between %s and %s", stringify(value), stringify(lowerLimit), stringify(upperLimit))
}

func ErrorValueFloat64OneOf(value float64, disallowedValues []float64) error {
	return errors.Preparedf(ErrorCodeValueDisallowed, "value is one of the disallowed values", "value %s is one of %s", stringify(value), stringify(disallowedValues))
}

func ErrorValueFloat64NotOneOf(value float64, allowedValues []float64) error {
	return errors.Preparedf(ErrorCodeValueNotAllowed, "value is not one of the allowed values", "value %s is not one of %s", stringify(value), stringify(allowedValues))
}

func ErrorValueIntOneOf(value int, disallowedValues []int) error {
	return errors.Preparedf(ErrorCodeValueDisallowed, "value is one of the disallowed values", "value %s is one of %s", stringify(value), stringify(disallowedValues))
}

func ErrorValueIntNotOneOf(value int, allowedValues []int) error {
	return errors.Preparedf(ErrorCodeValueNotAllowed, "value is not one of the allowed values", "value %s is not one of %s", stringify(value), stringify(allowedValues))
}

func ErrorValueStringOneOf(value string, disallowedValues []string) error {
	return errors.Preparedf(ErrorCodeValueDisallowed, "value is one of the disallowed values", "value %s is one of %s", stringify(value), stringify(disallowedValues))
}

func ErrorValueStringNotOneOf(value string, allowedValues []string) error {
	return errors.Preparedf(ErrorCodeValueNotAllowed, "value is not one of the allowed values", "value %s is not one of %s", stringify(value), stringify(allowedValues))
}

func ErrorValueStringMatches(value string, expression *regexp.Regexp) error {
	return errors.Preparedf(ErrorCodeValueMatches, "value matches expression", "value %s matches expression %s", stringify(value), stringify(expression))
}

func ErrorValueStringNotMatches(value string, expression *regexp.Regexp) error {
	return errors.Preparedf(ErrorCodeValueNotMatches, "value does not match expression", "value %s does not match expression %s", stringify(value), stringify(expression))
}

func ErrorValueStringAsTimeNotValid(value string, layout string) error {
	return errors.Preparedf(ErrorCodeValueNotValid, "value is not valid", "value %s is not valid as time with layout %s", stringify(value), stringify(layout))
}

func ErrorValueTimeNotAfter(value time.Time, limit time.Time) error {
	return errors.Preparedf(ErrorCodeValueNotAfter, "value is not after the specified time", "value %s is not after %s", stringify(value), stringify(limit))
}

func ErrorValueTimeNotAfterNow(value time.Time) error {
	return errors.Preparedf(ErrorCodeValueNotAfter, "value is not after the specified time", "value %s is not after now", stringify(value))
}

func ErrorValueTimeNotBefore(value time.Time, limit time.Time) error {
	return errors.Preparedf(ErrorCodeValueNotBefore, "value is not before the specified time", "value %s is not before %s", stringify(value), stringify(limit))
}

func ErrorValueTimeNotBeforeNow(value time.Time) error {
	return errors.Preparedf(ErrorCodeValueNotBefore, "value is not before the specified time", "value %s is not before now", stringify(value))
}

func ErrorValuesNotExistForAny(references ...string) error {
	return errors.Preparedf(ErrorCodeValuesNotExistForAny, "values do not exist for any", "values do not exist for any of %s", stringify(references))
}

func ErrorValuesNotExistForOne(references ...string) error {
	return errors.Preparedf(ErrorCodeValuesNotExistForOne, "values do not exist for one", "values do not exist for one of %s", stringify(references))
}

func ErrorLengthNotEqualTo(length int, limit int) error {
	return errors.Preparedf(ErrorCodeLengthOutOfRange, "length is out of range", "length %d is not equal to %d", length, limit)
}

func ErrorLengthEqualTo(length int, limit int) error {
	return errors.Preparedf(ErrorCodeLengthOutOfRange, "length is out of range", "length %d is equal to %d", length, limit)
}

func ErrorLengthNotLessThan(length int, limit int) error {
	return errors.Preparedf(ErrorCodeLengthOutOfRange, "length is out of range", "length %d is not less than %d", length, limit)
}

func ErrorLengthNotLessThanOrEqualTo(length int, limit int) error {
	return errors.Preparedf(ErrorCodeLengthOutOfRange, "length is out of range", "length %d is not less than or equal to %d", length, limit)
}

func ErrorLengthNotGreaterThan(length int, limit int) error {
	return errors.Preparedf(ErrorCodeLengthOutOfRange, "length is out of range", "length %d is not greater than %d", length, limit)
}

func ErrorLengthNotGreaterThanOrEqualTo(length int, limit int) error {
	return errors.Preparedf(ErrorCodeLengthOutOfRange, "length is out of range", "length %d is not greater than or equal to %d", length, limit)
}

func ErrorLengthNotInRange(length int, lowerLimit int, upperLimit int) error {
	return errors.Preparedf(ErrorCodeLengthOutOfRange, "length is out of range", "length %d is not between %d and %d", length, lowerLimit, upperLimit)
}

func stringify(interfaceValue interface{}) string {
	switch typeValue := interfaceValue.(type) {
	case float64:
		return strconv.FormatFloat(typeValue, 'f', -1, 64)
	case []float64:
		values := []string{}
		for _, value := range typeValue {
			values = append(values, stringify(value))
		}
		return fmt.Sprintf("[%s]", strings.Join(values, ", "))
	case []int:
		values := []string{}
		for _, value := range typeValue {
			values = append(values, stringify(value))
		}
		return fmt.Sprintf("[%s]", strings.Join(values, ", "))
	case string:
		return strconv.Quote(typeValue)
	case []string:
		values := []string{}
		for _, value := range typeValue {
			values = append(values, stringify(value))
		}
		return fmt.Sprintf("[%s]", strings.Join(values, ", "))
	case time.Time:
		return strconv.Quote(typeValue.Format(time.RFC3339Nano))
	case *regexp.Regexp:
		if typeValue == nil {
			return strconv.Quote("<MISSING>")
		}
		return strconv.Quote(typeValue.String())
	default:
		return fmt.Sprintf("%v", interfaceValue)
	}
}
