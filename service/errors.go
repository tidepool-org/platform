package service

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func ErrorInternalServerFailure() *Error {
	return &Error{
		Code:   "internal-server-failure",
		Status: http.StatusInternalServerError,
		Title:  "internal server failure",
		Detail: "Internal server failure",
	}
}

func ErrorAuthenticationTokenMissing() *Error {
	return &Error{
		Code:   "authentication-token-missing",
		Status: http.StatusUnauthorized,
		Title:  "authentication token missing",
		Detail: "Authentication token missing",
	}
}

func ErrorUnauthenticated() *Error {
	return &Error{
		Code:   "unauthenticated",
		Status: http.StatusUnauthorized,
		Title:  "authentication token is invalid",
		Detail: "Authentication token is invalid",
	}
}

func ErrorUnauthorized() *Error {
	return &Error{
		Code:   "unauthorized",
		Status: http.StatusForbidden,
		Title:  "authentication token is not authorized for requested action",
		Detail: "Authentication token is not authorized for requested action",
	}
}

func ErrorJSONMalformed() *Error {
	return &Error{
		Code:   "json-malformed",
		Status: http.StatusBadRequest,
		Title:  "json is malformed",
		Detail: "JSON is malformed",
	}
}

func ErrorTypeNotBoolean(value interface{}) *Error {
	return &Error{
		Code:   "type-not-boolean",
		Title:  "type is not boolean",
		Detail: fmt.Sprintf("Type is not boolean, but %T", value),
	}
}

func ErrorTypeNotUnsignedInteger(value interface{}) *Error {
	return &Error{
		Code:   "type-not-unsigned-integer",
		Title:  "type is not unsigned integer",
		Detail: fmt.Sprintf("Type is not unsigned integer, but %T", value),
	}
}

func ErrorTypeNotInteger(value interface{}) *Error {
	return &Error{
		Code:   "type-not-integer",
		Title:  "type is not integer",
		Detail: fmt.Sprintf("Type is not integer, but %T", value),
	}
}

func ErrorTypeNotFloat(value interface{}) *Error {
	return &Error{
		Code:   "type-not-float",
		Title:  "type is not float",
		Detail: fmt.Sprintf("Type is not float, but %T", value),
	}
}

func ErrorTypeNotString(value interface{}) *Error {
	return &Error{
		Code:   "type-not-string",
		Title:  "type is not string",
		Detail: fmt.Sprintf("Type is not string, but %T", value),
	}
}

func ErrorTypeNotObject(value interface{}) *Error {
	return &Error{
		Code:   "type-not-object",
		Title:  "type is not object",
		Detail: fmt.Sprintf("Type is not object, but %T", value),
	}
}

func ErrorTypeNotArray(value interface{}) *Error {
	return &Error{
		Code:   "type-not-array",
		Title:  "type is not array",
		Detail: fmt.Sprintf("Type is not array, but %T", value),
	}
}

func ErrorValueNotExists() *Error {
	return &Error{
		Code:   "value-not-exists",
		Title:  "value does not exist",
		Detail: "Value does not exist",
	}
}

func ErrorValueExists() *Error {
	return &Error{
		Code:   "value-exists",
		Title:  "value exists",
		Detail: "Value exists",
	}
}

func ErrorValueNotEmpty() *Error {
	return &Error{
		Code:   "value-not-empty",
		Title:  "value is not empty",
		Detail: "Value is not empty",
	}
}

func ErrorValueEmpty() *Error {
	return &Error{
		Code:   "value-empty",
		Title:  "value is empty",
		Detail: "Value is empty",
	}
}

func ErrorValueNotTrue() *Error {
	return &Error{
		Code:   "value-not-true",
		Title:  "value is not true",
		Detail: "Value is not true",
	}
}

func ErrorValueNotFalse() *Error {
	return &Error{
		Code:   "value-not-false",
		Title:  "value is not false",
		Detail: "Value is not false",
	}
}

func ErrorValueNotEqualTo(value interface{}, limit interface{}) *Error {
	return &Error{
		Code:   "value-out-of-range",
		Title:  "value is out of range",
		Detail: fmt.Sprintf("Value %v is not equal to %v", QuoteIfString(value), QuoteIfString(limit)),
	}
}

func ErrorValueEqualTo(value interface{}, limit interface{}) *Error {
	return &Error{
		Code:   "value-out-of-range",
		Title:  "value is out of range",
		Detail: fmt.Sprintf("Value %v is equal to %v", QuoteIfString(value), QuoteIfString(limit)),
	}
}

func ErrorValueNotLessThan(value interface{}, limit interface{}) *Error {
	return &Error{
		Code:   "value-out-of-range",
		Title:  "value is out of range",
		Detail: fmt.Sprintf("Value %v is not less than %v", QuoteIfString(value), QuoteIfString(limit)),
	}
}

func ErrorValueNotLessThanOrEqualTo(value interface{}, limit interface{}) *Error {
	return &Error{
		Code:   "value-out-of-range",
		Title:  "value is out of range",
		Detail: fmt.Sprintf("Value %v is not less than or equal to %v", QuoteIfString(value), QuoteIfString(limit)),
	}
}

func ErrorValueNotGreaterThan(value interface{}, limit interface{}) *Error {
	return &Error{
		Code:   "value-out-of-range",
		Title:  "value is out of range",
		Detail: fmt.Sprintf("Value %v is not greater than %v", QuoteIfString(value), QuoteIfString(limit)),
	}
}

func ErrorValueNotGreaterThanOrEqualTo(value interface{}, limit interface{}) *Error {
	return &Error{
		Code:   "value-out-of-range",
		Title:  "value is out of range",
		Detail: fmt.Sprintf("Value %v is not greater than or equal to %v", QuoteIfString(value), QuoteIfString(limit)),
	}
}

func ErrorValueNotInRange(value interface{}, lowerLimit interface{}, upperLimit interface{}) *Error {
	return &Error{
		Code:   "value-out-of-range",
		Title:  "value is out of range",
		Detail: fmt.Sprintf("Value %v is not between %v and %v", QuoteIfString(value), QuoteIfString(lowerLimit), QuoteIfString(upperLimit)),
	}
}

func ErrorValueIntegerOneOf(value int, disallowedValues []int) *Error {
	disallowedValuesStrings := []string{}
	for _, disallowedValue := range disallowedValues {
		disallowedValuesStrings = append(disallowedValuesStrings, fmt.Sprintf("%d", disallowedValue))
	}
	disallowedValuesString := strings.Join(disallowedValuesStrings, ", ")
	return &Error{
		Code:   "value-disallowed",
		Title:  "value is one of the disallowed values",
		Detail: fmt.Sprintf("Value %d is one of [%s]", value, disallowedValuesString),
	}
}

func ErrorValueIntegerNotOneOf(value int, allowedValues []int) *Error {
	allowedValuesStrings := []string{}
	for _, allowedValue := range allowedValues {
		allowedValuesStrings = append(allowedValuesStrings, fmt.Sprintf("%d", allowedValue))
	}
	allowedValuesString := strings.Join(allowedValuesStrings, ", ")
	return &Error{
		Code:   "value-not-allowed",
		Title:  "value is not one of the allowed values",
		Detail: fmt.Sprintf("Value %d is not one of [%s]", value, allowedValuesString),
	}
}

func ErrorValueFloatOneOf(value float64, disallowedValues []float64) *Error {
	disallowedValuesStrings := []string{}
	for _, disallowedValue := range disallowedValues {
		disallowedValuesStrings = append(disallowedValuesStrings, fmt.Sprintf("%v", disallowedValue))
	}
	disallowedValuesString := strings.Join(disallowedValuesStrings, ", ")
	return &Error{
		Code:   "value-disallowed",
		Title:  "value is one of the disallowed values",
		Detail: fmt.Sprintf("Value %v is one of [%s]", value, disallowedValuesString),
	}
}

func ErrorValueFloatNotOneOf(value float64, allowedValues []float64) *Error {
	allowedValuesStrings := []string{}
	for _, allowedValue := range allowedValues {
		allowedValuesStrings = append(allowedValuesStrings, fmt.Sprintf("%v", allowedValue))
	}
	allowedValuesString := strings.Join(allowedValuesStrings, ", ")
	return &Error{
		Code:   "value-not-allowed",
		Title:  "value is not one of the allowed values",
		Detail: fmt.Sprintf("Value %v is not one of [%s]", value, allowedValuesString),
	}
}

func ErrorValueStringOneOf(value string, disallowedValues []string) *Error {
	disallowedValuesStrings := []string{}
	for _, disallowedValue := range disallowedValues {
		disallowedValuesStrings = append(disallowedValuesStrings, strconv.Quote(disallowedValue))
	}
	disallowedValuesString := strings.Join(disallowedValuesStrings, ", ")
	return &Error{
		Code:   "value-disallowed",
		Title:  "value is one of the disallowed values",
		Detail: fmt.Sprintf("Value %s is one of [%v]", strconv.Quote(value), disallowedValuesString),
	}
}

func ErrorValueStringNotOneOf(value string, allowedValues []string) *Error {
	allowedValuesStrings := []string{}
	for _, allowedValue := range allowedValues {
		allowedValuesStrings = append(allowedValuesStrings, strconv.Quote(allowedValue))
	}
	allowedValuesString := strings.Join(allowedValuesStrings, ", ")
	return &Error{
		Code:   "value-not-allowed",
		Title:  "value is not one of the allowed values",
		Detail: fmt.Sprintf("Value %s is not one of [%v]", strconv.Quote(value), allowedValuesString),
	}
}

func ErrorValueTimeNotValid(value string, timeLayout string) *Error {
	return &Error{
		Code:   "value-not-valid",
		Title:  "value is not a valid time",
		Detail: fmt.Sprintf("Value %s is not a valid time of format %s", strconv.Quote(value), strconv.Quote(timeLayout)),
	}
}

func ErrorValueTimeNotAfter(value time.Time, limit time.Time, timeLayout string) *Error {
	return &Error{
		Code:   "value-not-after",
		Title:  "value is not after the specified time",
		Detail: fmt.Sprintf("Value %s is not after %s", strconv.Quote(value.Format(timeLayout)), strconv.Quote(limit.Format(timeLayout))),
	}
}

func ErrorValueTimeNotAfterNow(value time.Time, timeLayout string) *Error {
	return &Error{
		Code:   "value-not-after",
		Title:  "value is not after the specified time",
		Detail: fmt.Sprintf("Value %s is not after now", strconv.Quote(value.Format(timeLayout))),
	}
}

func ErrorValueTimeNotBefore(value time.Time, limit time.Time, timeLayout string) *Error {
	return &Error{
		Code:   "value-not-before",
		Title:  "value is not before the specified time",
		Detail: fmt.Sprintf("Value %s is not before %s", strconv.Quote(value.Format(timeLayout)), strconv.Quote(limit.Format(timeLayout))),
	}
}

func ErrorValueTimeNotBeforeNow(value time.Time, timeLayout string) *Error {
	return &Error{
		Code:   "value-not-before",
		Title:  "value is not before the specified time",
		Detail: fmt.Sprintf("Value %s is not before now", strconv.Quote(value.Format(timeLayout))),
	}
}

func ErrorLengthNotEqualTo(length int, limit int) *Error {
	return &Error{
		Code:   "length-out-of-range",
		Title:  "length is out of range",
		Detail: fmt.Sprintf("Length %d is not equal to %d", length, limit),
	}
}

func ErrorLengthEqualTo(length int, limit int) *Error {
	return &Error{
		Code:   "length-out-of-range",
		Title:  "length is out of range",
		Detail: fmt.Sprintf("Length %d is equal to %d", length, limit),
	}
}

func ErrorLengthNotLessThan(length int, limit int) *Error {
	return &Error{
		Code:   "length-out-of-range",
		Title:  "length is out of range",
		Detail: fmt.Sprintf("Length %d is not less than %d", length, limit),
	}
}

func ErrorLengthNotLessThanOrEqualTo(length int, limit int) *Error {
	return &Error{
		Code:   "length-out-of-range",
		Title:  "length is out of range",
		Detail: fmt.Sprintf("Length %d is not less than or equal to %d", length, limit),
	}
}

func ErrorLengthNotGreaterThan(length int, limit int) *Error {
	return &Error{
		Code:   "length-out-of-range",
		Title:  "length is out of range",
		Detail: fmt.Sprintf("Length %d is not greater than %d", length, limit),
	}
}

func ErrorLengthNotGreaterThanOrEqualTo(length int, limit int) *Error {
	return &Error{
		Code:   "length-out-of-range",
		Title:  "length is out of range",
		Detail: fmt.Sprintf("Length %d is not greater than or equal to %d", length, limit),
	}
}

func ErrorLengthNotInRange(length int, lowerLimit int, upperLimit int) *Error {
	return &Error{
		Code:   "length-out-of-range",
		Title:  "length is out of range",
		Detail: fmt.Sprintf("Length %d is not between %d and %d", length, lowerLimit, upperLimit),
	}
}

func QuoteIfString(interfaceValue interface{}) interface{} {
	if stringValue, ok := interfaceValue.(string); ok {
		return strconv.Quote(stringValue)
	}
	return interfaceValue
}
