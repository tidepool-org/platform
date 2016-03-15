package validate

import (
	"errors"
	"reflect"
	"time"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"
)

//Validator interface
type Validator interface {
	ValidateStruct(s interface{}) error
	RegisterStructValidation(fn validator.StructLevelFunc, types ...interface{})
}

//PlatformValidator type that implements Validator
type PlatformValidator struct {
	validate *validator.Validate
}

//NewPlatformValidator returns initialised PlatformValidator with custom tidepool validation
func NewPlatformValidator() *PlatformValidator {
	validate := validator.New(&validator.Config{TagName: "valid"})
	validate.RegisterValidation("datetime", datetime)
	return &PlatformValidator{validate: validate}
}

//ValidateStruct for the PlatformValidator
func (pv *PlatformValidator) ValidateStruct(s interface{}) error {

	if errs := pv.validate.Struct(s); errs != nil {
		return errors.New(errs.Error())
	}
	return nil
}

// RegisterStructValidation registers a DataStructLevelFunc against a number of data types
// NOTE: this method is not thread-safe it is intended that these all be registered prior to any validation
func (pv *PlatformValidator) RegisterStructValidation(fn validator.StructLevelFunc, types ...interface{}) {
	pv.validate.RegisterStructValidation(fn, types)
}

//datetime is a custom validation method for how we require dates are formated
func datetime(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {

	if timeTime, ok := field.Interface().(time.Time); ok {
		if !timeTime.IsZero() && timeTime.Before(time.Now()) {
			return true
		}
		return false
	}

	if timeStr, ok := field.Interface().(string); ok {
		_, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			//try this format also before we fail
			_, err = time.Parse("2006-01-02T15:04:05", timeStr)
			if err != nil {
				return false
			}
		}
		return true
	}
	return false
}
