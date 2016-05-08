package types

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
		Detail: fmt.Sprintf("Type %s is invalid", invalidType),
	}
}

func ErrorSubTypeInvalid(invalidSubType string) *service.Error {
	return &service.Error{
		Code:   "sub-type-invalid",
		Title:  "sub type is invalid",
		Detail: fmt.Sprintf("Sub type %s is invalid", invalidSubType),
	}
}
