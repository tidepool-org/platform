package validator

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
	"html"
	"strings"
	"time"

	"github.com/tidepool-org/platform/service"
)

// TODO: Review all errors for consistency and language
// Once shipped, Code and Title cannot change

func ErrorValueDoesNotExist() *service.Error {
	return &service.Error{
		Code:   "value-does-not-exist",
		Title:  "value does not exist",
		Detail: "Value does not exist",
	}
}

func ErrorValueNotTrue() *service.Error {
	return &service.Error{
		Code:   "value-not-true",
		Title:  "value is not true",
		Detail: "Value is not true",
	}
}

func ErrorValueNotFalse() *service.Error {
	return &service.Error{
		Code:   "value-not-false",
		Title:  "value is not false",
		Detail: "Value is not false",
	}
}

func ErrorValueNotEqualTo(value interface{}, limit interface{}) *service.Error {
	return &service.Error{
		Code:   "value-out-of-range",
		Title:  "value is out of range",
		Detail: fmt.Sprintf("Value %v is not equal to %v", value, limit),
	}
}

func ErrorValueEqualTo(value interface{}, limit interface{}) *service.Error {
	return &service.Error{
		Code:   "value-out-of-range",
		Title:  "value is out of range",
		Detail: fmt.Sprintf("Value %v is equal to %v", value, limit),
	}
}

func ErrorValueNotLessThan(value interface{}, limit interface{}) *service.Error {
	return &service.Error{
		Code:   "value-out-of-range",
		Title:  "value is out of range",
		Detail: fmt.Sprintf("Value %v is not less than %v", value, limit),
	}
}

func ErrorValueNotLessThanOrEqual(value interface{}, limit interface{}) *service.Error {
	return &service.Error{
		Code:   "value-out-of-range",
		Title:  "value is out of range",
		Detail: fmt.Sprintf("Value %v is not less than or equal to %v", value, limit),
	}
}

func ErrorValueNotGreaterThan(value interface{}, limit interface{}) *service.Error {
	return &service.Error{
		Code:   "value-out-of-range",
		Title:  "value is out of range",
		Detail: fmt.Sprintf("Value %v is not greater than %v", value, limit),
	}
}

func ErrorValueNotGreaterThanOrEqual(value interface{}, limit interface{}) *service.Error {
	return &service.Error{
		Code:   "value-out-of-range",
		Title:  "value is out of range",
		Detail: fmt.Sprintf("Value %v is not greater than or equal to %v", value, limit),
	}
}

func ErrorIntegerNotInRange(value int, lowerlimit int, upperLimit int) *service.Error {
	return &service.Error{
		Code:   "value-out-of-range",
		Title:  "value is out of range",
		Detail: fmt.Sprintf("Value %d is not between %d and %d", value, lowerlimit, upperLimit),
	}
}

func ErrorFloatNotInRange(value float64, lowerlimit float64, upperLimit float64) *service.Error {
	return &service.Error{
		Code:   "value-out-of-range",
		Title:  "value is out of range",
		Detail: fmt.Sprintf("Value %v is not between %v and %v", value, lowerlimit, upperLimit),
	}
}

func ErrorIntegerOneOf(value int, disallowedValues []int) *service.Error {
	disallowedValuesStrings := []string{}
	for _, disallowedValue := range disallowedValues {
		disallowedValuesStrings = append(disallowedValuesStrings, fmt.Sprintf("%d", disallowedValue))
	}
	disallowedValuesString := strings.Join(disallowedValuesStrings, ", ")
	return &service.Error{
		Code:   "value-disallowed",
		Title:  "value is one of the disallowed values",
		Detail: fmt.Sprintf("Value %d is one of [%s]", value, disallowedValuesString),
	}
}

func ErrorIntegerNotOneOf(value int, allowedValues []int) *service.Error {
	allowedValuesStrings := []string{}
	for _, allowedValue := range allowedValues {
		allowedValuesStrings = append(allowedValuesStrings, fmt.Sprintf("%d", allowedValue))
	}
	allowedValuesString := strings.Join(allowedValuesStrings, ", ")
	return &service.Error{
		Code:   "value-not-allowed",
		Title:  "value is not one of the allowed values",
		Detail: fmt.Sprintf("Value %d is not one of [%s]", value, allowedValuesString),
	}
}

func ErrorFloatOneOf(value float64, disallowedValues []float64) *service.Error {
	disallowedValuesStrings := []string{}
	for _, disallowedValue := range disallowedValues {
		disallowedValuesStrings = append(disallowedValuesStrings, fmt.Sprintf("%v", disallowedValue))
	}
	disallowedValuesString := strings.Join(disallowedValuesStrings, ", ")
	return &service.Error{
		Code:   "value-disallowed",
		Title:  "value is one of the disallowed values",
		Detail: fmt.Sprintf("Value %v is one of [%s]", value, disallowedValuesString),
	}
}

func ErrorFloatNotOneOf(value float64, allowedValues []float64) *service.Error {
	allowedValuesStrings := []string{}
	for _, allowedValue := range allowedValues {
		allowedValuesStrings = append(allowedValuesStrings, fmt.Sprintf("%v", allowedValue))
	}
	allowedValuesString := strings.Join(allowedValuesStrings, ", ")
	return &service.Error{
		Code:   "value-not-allowed",
		Title:  "value is not one of the allowed values",
		Detail: fmt.Sprintf("Value %v is not one of [%s]", value, allowedValuesString),
	}
}

func ErrorLengthNotEqualTo(length int, limit int) *service.Error {
	return &service.Error{
		Code:   "length-out-of-range",
		Title:  "length is out of range",
		Detail: fmt.Sprintf("Length %d is not equal to %d", length, limit),
	}
}

func ErrorLengthEqualTo(length int, limit int) *service.Error {
	return &service.Error{
		Code:   "length-out-of-range",
		Title:  "length is out of range",
		Detail: fmt.Sprintf("Length %d is equal to %d", length, limit),
	}
}

func ErrorLengthNotLessThan(length int, limit int) *service.Error {
	return &service.Error{
		Code:   "length-out-of-range",
		Title:  "length is out of range",
		Detail: fmt.Sprintf("Length %d is not less than %d", length, limit),
	}
}

func ErrorLengthNotLessThanOrEqual(length int, limit int) *service.Error {
	return &service.Error{
		Code:   "length-out-of-range",
		Title:  "length is out of range",
		Detail: fmt.Sprintf("Length %d is not less than or equal to %d", length, limit),
	}
}

func ErrorLengthNotGreaterThan(length int, limit int) *service.Error {
	return &service.Error{
		Code:   "length-out-of-range",
		Title:  "length is out of range",
		Detail: fmt.Sprintf("Length %d is not greater than %d", length, limit),
	}
}

func ErrorLengthNotGreaterThanOrEqual(length int, limit int) *service.Error {
	return &service.Error{
		Code:   "length-out-of-range",
		Title:  "length is out of range",
		Detail: fmt.Sprintf("Length %d is not greater than or equal to %d", length, limit),
	}
}

func ErrorLengthNotInRange(length int, lowerlimit int, upperLimit int) *service.Error {
	return &service.Error{
		Code:   "length-out-of-range",
		Title:  "length is out of range",
		Detail: fmt.Sprintf("Length %d is not between %d and %d", length, lowerlimit, upperLimit),
	}
}

func ErrorStringOneOf(value string, disallowedValues []string) *service.Error {
	disallowedValuesStrings := []string{}
	for _, disallowedValue := range disallowedValues {
		disallowedValuesStrings = append(disallowedValuesStrings, fmt.Sprintf("'%s'", html.EscapeString(disallowedValue)))
	}
	disallowedValuesString := strings.Join(disallowedValuesStrings, ", ")
	return &service.Error{
		Code:   "value-disallowed",
		Title:  "value is one of the disallowed values",
		Detail: fmt.Sprintf("Value '%s' is one of [%v]", value, disallowedValuesString),
	}
}

func ErrorStringNotOneOf(value string, allowedValues []string) *service.Error {
	allowedValuesStrings := []string{}
	for _, allowedValue := range allowedValues {
		allowedValuesStrings = append(allowedValuesStrings, fmt.Sprintf("'%s'", html.EscapeString(allowedValue)))
	}
	allowedValuesString := strings.Join(allowedValuesStrings, ", ")
	return &service.Error{
		Code:   "value-not-allowed",
		Title:  "value is not one of the allowed values",
		Detail: fmt.Sprintf("Value '%s' is not one of [%v]", value, allowedValuesString),
	}
}

func ErrorTimeNotValid(value string, timeLayout string) *service.Error {
	return &service.Error{
		Code:   "time-not-valid",
		Title:  "value is not a valid time",
		Detail: fmt.Sprintf("Value '%s' is not a valid time of format '%s'", value, timeLayout),
	}
}

func ErrorTimeNotAfter(value time.Time, limit time.Time, timeLayout string) *service.Error {
	return &service.Error{
		Code:   "time-not-after",
		Title:  "value is not after the specified time",
		Detail: fmt.Sprintf("Value '%s' is not after '%s'", value.Format(timeLayout), limit.Format(timeLayout)),
	}
}

func ErrorTimeNotAfterNow(value time.Time, timeLayout string) *service.Error {
	return &service.Error{
		Code:   "time-not-after",
		Title:  "value is not after the specified time",
		Detail: fmt.Sprintf("Value '%s' is not after now", value.Format(timeLayout)),
	}
}

func ErrorTimeNotBefore(value time.Time, limit time.Time, timeLayout string) *service.Error {
	return &service.Error{
		Code:   "time-not-before",
		Title:  "value is not before the specified time",
		Detail: fmt.Sprintf("Value '%s' is not before '%s'", value.Format(timeLayout), limit.Format(timeLayout)),
	}
}

func ErrorTimeNotBeforeNow(value time.Time, timeLayout string) *service.Error {
	return &service.Error{
		Code:   "time-not-before",
		Title:  "value is not before the specified time",
		Detail: fmt.Sprintf("Value '%s' is not before now", value.Format(timeLayout)),
	}
}
