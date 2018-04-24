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
	ErrorCodeValueNotExists   = "value-not-exists"
	ErrorCodeValueExists      = "value-exists"
	ErrorCodeValueNotEmpty    = "value-not-empty"
	ErrorCodeValueEmpty       = "value-empty"
	ErrorCodeValueNotTrue     = "value-not-true"
	ErrorCodeValueNotFalse    = "value-not-false"
	ErrorCodeValueOutOfRange  = "value-out-of-range"
	ErrorCodeValueDisallowed  = "value-disallowed"
	ErrorCodeValueNotAllowed  = "value-not-allowed"
	ErrorCodeValueMatches     = "value-matches"
	ErrorCodeValueNotMatches  = "value-not-matches"
	ErrorCodeValueZero        = "value-zero"
	ErrorCodeValueNotZero     = "value-not-zero"
	ErrorCodeValueNotAfter    = "value-not-after"
	ErrorCodeValueNotBefore   = "value-not-before"
	ErrorCodeValueNotValid    = "value-not-valid"
	ErrorCodeLengthOutOfRange = "length-out-of-range"
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

func ErrorValueNotTrue() error {
	return errors.Prepared(ErrorCodeValueNotTrue, "value is not true", "value is not true")
}

func ErrorValueNotFalse() error {
	return errors.Prepared(ErrorCodeValueNotFalse, "value is not false", "value is not false")
}

func ErrorValueNotEqualTo(value interface{}, limit interface{}) error {
	return errors.Preparedf(ErrorCodeValueOutOfRange, "value is out of range", "value %v is not equal to %v", QuoteIfString(value), QuoteIfString(limit))
}

func ErrorValueEqualTo(value interface{}, limit interface{}) error {
	return errors.Preparedf(ErrorCodeValueOutOfRange, "value is out of range", "value %v is equal to %v", QuoteIfString(value), QuoteIfString(limit))
}

func ErrorValueNotLessThan(value interface{}, limit interface{}) error {
	return errors.Preparedf(ErrorCodeValueOutOfRange, "value is out of range", "value %v is not less than %v", QuoteIfString(value), QuoteIfString(limit))
}

func ErrorValueNotLessThanOrEqualTo(value interface{}, limit interface{}) error {
	return errors.Preparedf(ErrorCodeValueOutOfRange, "value is out of range", "value %v is not less than or equal to %v", QuoteIfString(value), QuoteIfString(limit))
}

func ErrorValueNotGreaterThan(value interface{}, limit interface{}) error {
	return errors.Preparedf(ErrorCodeValueOutOfRange, "value is out of range", "value %v is not greater than %v", QuoteIfString(value), QuoteIfString(limit))
}

func ErrorValueNotGreaterThanOrEqualTo(value interface{}, limit interface{}) error {
	return errors.Preparedf(ErrorCodeValueOutOfRange, "value is out of range", "value %v is not greater than or equal to %v", QuoteIfString(value), QuoteIfString(limit))
}

func ErrorValueNotInRange(value interface{}, lowerLimit interface{}, upperLimit interface{}) error {
	return errors.Preparedf(ErrorCodeValueOutOfRange, "value is out of range", "value %v is not between %v and %v", QuoteIfString(value), QuoteIfString(lowerLimit), QuoteIfString(upperLimit))
}

func ErrorValueFloat64OneOf(value float64, disallowedValues []float64) error {
	disallowedValuesStrings := []string{}
	for _, disallowedValue := range disallowedValues {
		disallowedValuesStrings = append(disallowedValuesStrings, fmt.Sprintf("%v", disallowedValue))
	}
	disallowedValuesString := strings.Join(disallowedValuesStrings, ", ")
	return errors.Preparedf(ErrorCodeValueDisallowed, "value is one of the disallowed values", "value %v is one of [%s]", value, disallowedValuesString)
}

func ErrorValueFloat64NotOneOf(value float64, allowedValues []float64) error {
	allowedValuesStrings := []string{}
	for _, allowedValue := range allowedValues {
		allowedValuesStrings = append(allowedValuesStrings, fmt.Sprintf("%v", allowedValue))
	}
	allowedValuesString := strings.Join(allowedValuesStrings, ", ")
	return errors.Preparedf(ErrorCodeValueNotAllowed, "value is not one of the allowed values", "value %v is not one of [%s]", value, allowedValuesString)
}

func ErrorValueIntOneOf(value int, disallowedValues []int) error {
	disallowedValuesStrings := []string{}
	for _, disallowedValue := range disallowedValues {
		disallowedValuesStrings = append(disallowedValuesStrings, fmt.Sprintf("%d", disallowedValue))
	}
	disallowedValuesString := strings.Join(disallowedValuesStrings, ", ")
	return errors.Preparedf(ErrorCodeValueDisallowed, "value is one of the disallowed values", "value %d is one of [%s]", value, disallowedValuesString)
}

func ErrorValueIntNotOneOf(value int, allowedValues []int) error {
	allowedValuesStrings := []string{}
	for _, allowedValue := range allowedValues {
		allowedValuesStrings = append(allowedValuesStrings, fmt.Sprintf("%d", allowedValue))
	}
	allowedValuesString := strings.Join(allowedValuesStrings, ", ")
	return errors.Preparedf(ErrorCodeValueNotAllowed, "value is not one of the allowed values", "value %d is not one of [%s]", value, allowedValuesString)
}

func ErrorValueStringOneOf(value string, disallowedValues []string) error {
	disallowedValuesStrings := []string{}
	for _, disallowedValue := range disallowedValues {
		disallowedValuesStrings = append(disallowedValuesStrings, strconv.Quote(disallowedValue))
	}
	disallowedValuesString := strings.Join(disallowedValuesStrings, ", ")
	return errors.Preparedf(ErrorCodeValueDisallowed, "value is one of the disallowed values", "value %q is one of [%v]", value, disallowedValuesString)
}

func ErrorValueStringNotOneOf(value string, allowedValues []string) error {
	allowedValuesStrings := []string{}
	for _, allowedValue := range allowedValues {
		allowedValuesStrings = append(allowedValuesStrings, strconv.Quote(allowedValue))
	}
	allowedValuesString := strings.Join(allowedValuesStrings, ", ")
	return errors.Preparedf(ErrorCodeValueNotAllowed, "value is not one of the allowed values", "value %q is not one of [%v]", value, allowedValuesString)
}

func ErrorValueStringMatches(value string, expression *regexp.Regexp) error {
	return errors.Preparedf(ErrorCodeValueMatches, "value matches expression", "value %q matches expression %q", value, ExpressionAsString(expression))
}

func ErrorValueStringNotMatches(value string, expression *regexp.Regexp) error {
	return errors.Preparedf(ErrorCodeValueNotMatches, "value does not match expression", "value %q does not match expression %q", value, ExpressionAsString(expression))
}

func ErrorValueStringAsTimeNotValid(value string, layout string) error {
	return errors.Preparedf(ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as time with layout %q", value, layout)
}

func ErrorValueTimeZero(value time.Time) error {
	return errors.Preparedf(ErrorCodeValueZero, "value is zero", "value %q is zero", value.Format(time.RFC3339))
}

func ErrorValueTimeNotZero(value time.Time) error {
	return errors.Preparedf(ErrorCodeValueNotZero, "value is not zero", "value %q is not zero", value.Format(time.RFC3339))
}

func ErrorValueTimeNotAfter(value time.Time, limit time.Time) error {
	return errors.Preparedf(ErrorCodeValueNotAfter, "value is not after the specified time", "value %q is not after %q", value.Format(time.RFC3339), limit.Format(time.RFC3339))
}

func ErrorValueTimeNotAfterNow(value time.Time) error {
	return errors.Preparedf(ErrorCodeValueNotAfter, "value is not after the specified time", "value %q is not after now", value.Format(time.RFC3339))
}

func ErrorValueTimeNotBefore(value time.Time, limit time.Time) error {
	return errors.Preparedf(ErrorCodeValueNotBefore, "value is not before the specified time", "value %q is not before %q", value.Format(time.RFC3339), limit.Format(time.RFC3339))
}

func ErrorValueTimeNotBeforeNow(value time.Time) error {
	return errors.Preparedf(ErrorCodeValueNotBefore, "value is not before the specified time", "value %q is not before now", value.Format(time.RFC3339))
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

func QuoteIfString(interfaceValue interface{}) interface{} {
	if stringValue, ok := interfaceValue.(string); ok {
		return strconv.Quote(stringValue)
	}
	return interfaceValue
}

func ExpressionAsString(expression *regexp.Regexp) string {
	if expression == nil {
		return "<MISSING>"
	}
	return expression.String()
}
