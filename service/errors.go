package service

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func ErrorInternalServerFailure() *Error {
	return &Error{
		Code:   "internal-server-failure",
		Status: http.StatusInternalServerError,
		Title:  "internal server failure",
		Detail: "Internal server failure",
	}
}

func ErrorUnauthenticated() *Error {
	return &Error{
		Code:   "unauthenticated",
		Status: http.StatusUnauthorized,
		Title:  "auth token is invalid",
		Detail: "Auth token is invalid",
	}
}

func ErrorUnauthorized() *Error {
	return &Error{
		Code:   "unauthorized",
		Status: http.StatusForbidden,
		Title:  "auth token is not authorized for requested action",
		Detail: "Auth token is not authorized for requested action",
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

func ErrorValueStringNotOneOf(value string, allowedValues []string) *Error {
	allowedValuesStrings := []string{}
	for _, allowedValue := range allowedValues {
		allowedValuesStrings = append(allowedValuesStrings, strconv.Quote(allowedValue))
	}
	allowedValuesString := strings.Join(allowedValuesStrings, ", ")
	return &Error{
		Code:   "value-not-allowed",
		Title:  "value is not one of the allowed values",
		Detail: fmt.Sprintf("Value %q is not one of [%v]", value, allowedValuesString),
	}
}

func QuoteIfString(interfaceValue interface{}) interface{} {
	if stringValue, ok := interfaceValue.(string); ok {
		return strconv.Quote(stringValue)
	}
	return interfaceValue
}
