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

	"github.com/tidepool-org/platform/service"
)

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

func ErrorIntegerNotOneOf(value int, possibleValues []int) *service.Error {
	possibleValuesStrings := []string{}
	for _, possibleValue := range possibleValues {
		possibleValuesStrings = append(possibleValuesStrings, fmt.Sprintf("%d", possibleValue))
	}
	possibleValuesString := strings.Join(possibleValuesStrings, ", ")
	return &service.Error{
		Code:   "value-not-allowed",
		Title:  "value is not one of the allowed values",
		Detail: fmt.Sprintf("Value %d is not one of [%s]", value, possibleValuesString),
	}
}

func ErrorFloatNotOneOf(value float64, possibleValues []float64) *service.Error {
	possibleValuesStrings := []string{}
	for _, possibleValue := range possibleValues {
		possibleValuesStrings = append(possibleValuesStrings, fmt.Sprintf("%v", possibleValue))
	}
	possibleValuesString := strings.Join(possibleValuesStrings, ", ")
	return &service.Error{
		Code:   "value-not-allowed",
		Title:  "value is not one of the allowed values",
		Detail: fmt.Sprintf("Value %v is not one of [%s]", value, possibleValuesString),
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

func ErrorStringNotOneOf(value string, possibleValues []string) *service.Error {
	possibleValuesStrings := []string{}
	for _, possibleValue := range possibleValues {
		possibleValuesStrings = append(possibleValuesStrings, fmt.Sprintf("'%s'", html.EscapeString(possibleValue)))
	}
	possibleValuesString := strings.Join(possibleValuesStrings, ", ")
	return &service.Error{
		Code:   "value-not-allowed",
		Title:  "value is not one of the allowed values",
		Detail: fmt.Sprintf("Value '%s' is not one of [%v]", value, possibleValuesString),
	}
}
