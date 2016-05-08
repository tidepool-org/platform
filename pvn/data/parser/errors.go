package parser

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"fmt"

	"github.com/tidepool-org/platform/service"
)

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
