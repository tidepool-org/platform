package types

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

func ErrorValueMissing() *service.Error {
	return &service.Error{
		Code:   "value-missing",
		Title:  "value is missing",
		Detail: "Value is missing",
	}
}

func ErrorTypeInvalid(invalidType string) *service.Error {
	return &service.Error{
		Code:   "type-invalid",
		Title:  "type is invalid",
		Detail: fmt.Sprintf("Type '%s' is invalid", invalidType),
	}
}

func ErrorSubTypeInvalid(invalidSubType string) *service.Error {
	return &service.Error{
		Code:   "sub-type-invalid",
		Title:  "sub type is invalid",
		Detail: fmt.Sprintf("Sub type '%s' is invalid", invalidSubType),
	}
}
