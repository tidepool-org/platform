package validate

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"
)

//Validator interface
type Validator interface {
	Struct(s interface{}) Errors
	Field(field interface{}, tag ValidationTag) Errors
	RegisterValidation(tag ValidationTag, fn validator.Func)
}

//PlatformValidator type that implements Validator
type PlatformValidator struct {
	validate *validator.Validate
}

// FieldError contains a single field's validation error along
// with other properties that may be needed for error message creation
type FieldError struct {
	Field string
	Tag   string
	Type  reflect.Type
	Value interface{}
}

//ValidationTag that all tags will be of this type
type ValidationTag string

// Errors is a type of map[string]*FieldError
// it exists to allow for multiple errors to be passed from this library
// and yet still subscribe to the error interface
type Errors map[string]*FieldError

// ErrorReasons is a type of map[ValidationTag]string
// it allows us to map a ValidationTag to a reason why the validation failed
type ErrorReasons map[ValidationTag]string

// Error is intended for use in development + debugging and not intended to be a production error message.
// It allows ValidationErrors to subscribe to the Error interface.
// All information to create an error message specific to your application is contained within
// the FieldError found within the ValidationErrors map
func (e Errors) Error() string {
	const fieldErrMsg = "Key: '%s' Error:Field validation for '%s' failed on the '%s' tag"
	if len(e) > 0 {

		buff := bytes.NewBufferString("")

		for key, err := range e {
			buff.WriteString(fmt.Sprintf(fieldErrMsg, key, err.Field, err.Tag))
			buff.WriteString("\n")
		}

		return strings.TrimSpace(buff.String())
	}
	return ""
}

//GetError returns a formatted error message for the user
func (e Errors) GetError(reasons ErrorReasons) error {

	const (
		fieldErrorMsg = "Error:Field validation for '%s' failed with '%s' when given '%v' for type '%s' "
		tagErrorMsg   = "Error:Field validation failed with '%s' when given '%v' for type '%s'"
	)

	if len(e) > 0 {
		buff := bytes.NewBufferString("")

		for _, err := range e {

			if reason, ok := reasons[ValidationTag(err.Tag)]; ok {

				if err.Field == "" {
					buff.WriteString(fmt.Sprintf(tagErrorMsg, reason, err.Value, err.Type))
				} else {
					buff.WriteString(fmt.Sprintf(fieldErrorMsg, err.Field, reason, err.Value, err.Type))
				}

			} else {
				buff.WriteString(fmt.Sprintf(fieldErrorMsg, err.Field, err.Tag, err.Value, err.Type))
			}
			buff.WriteString("\n")
		}
		return errors.New(strings.TrimSpace(buff.String()))
	}

	return nil
}

//NewPlatformValidator returns initialised PlatformValidator with custom tidepool validation
func NewPlatformValidator() *PlatformValidator {
	validate := validator.New(&validator.Config{TagName: "valid"})
	return &PlatformValidator{validate: validate}
}

func toErrors(ve validator.ValidationErrors) Errors {
	errs := Errors{}
	for k, v := range ve {
		errs[k] = &FieldError{Field: v.Field, Tag: v.Tag, Type: v.Type, Value: v.Value}
	}
	return errs
}

//Struct validation for the PlatformValidator
func (pv *PlatformValidator) Struct(s interface{}) Errors {
	errs := pv.validate.Struct(s)
	if errs != nil {
		return toErrors(errs.(validator.ValidationErrors))
	}
	return nil
}

//Field for the PlatformValidator
//NOTE: this is really used in tests at this stage
func (pv *PlatformValidator) Field(field interface{}, tag ValidationTag) Errors {
	errs := pv.validate.Field(field, string(tag))
	if errs != nil {
		return toErrors(errs.(validator.ValidationErrors))
	}
	return nil
}

//RegisterValidation so we can add our own validation functions
func (pv *PlatformValidator) RegisterValidation(tag ValidationTag, fn validator.Func) {
	pv.validate.RegisterValidation(string(tag), fn)
}
