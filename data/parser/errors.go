package parser

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"fmt"

	"github.com/tidepool-org/platform/service"
)

// TODO: Review all errors for consistency and language
// Once shipped, Code and Title cannot change

func ErrorTypeNotBoolean(value interface{}) *service.Error {
	return &service.Error{
		Code:   "type-not-boolean",
		Title:  "type is not boolean",
		Detail: fmt.Sprintf("Type is not boolean, but %T", value),
	}
}

func ErrorTypeNotInteger(value interface{}) *service.Error {
	return &service.Error{
		Code:   "type-not-integer",
		Title:  "type is not integer",
		Detail: fmt.Sprintf("Type is not integer, but %T", value),
	}
}

func ErrorTypeNotFloat(value interface{}) *service.Error {
	return &service.Error{
		Code:   "type-not-float",
		Title:  "type is not float",
		Detail: fmt.Sprintf("Type is not float, but %T", value),
	}
}

func ErrorTypeNotString(value interface{}) *service.Error {
	return &service.Error{
		Code:   "type-not-string",
		Title:  "type is not string",
		Detail: fmt.Sprintf("Type is not string, but %T", value),
	}
}

func ErrorTypeNotObject(value interface{}) *service.Error {
	return &service.Error{
		Code:   "type-not-object",
		Title:  "type is not object",
		Detail: fmt.Sprintf("Type is not object, but %T", value),
	}
}

func ErrorTypeNotArray(value interface{}) *service.Error {
	return &service.Error{
		Code:   "type-not-array",
		Title:  "type is not array",
		Detail: fmt.Sprintf("Type is not array, but %T", value),
	}
}

func ErrorNotParsed() *service.Error {
	return &service.Error{
		Code:   "not-parsed",
		Title:  "not parsed",
		Detail: "Not parsed",
	}
}
