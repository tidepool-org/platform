package validate

import (
	"errors"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"
)

//Validator interface
type Validator interface {
	ValidateStruct(s interface{}) error
	RegisterValidation(key string, fn validator.Func)
}

//PlatformValidator type that implements Validator
type PlatformValidator struct {
	validate *validator.Validate
}

//NewPlatformValidator returns initialised PlatformValidator with custom tidepool validation
func NewPlatformValidator() *PlatformValidator {
	validate := validator.New(&validator.Config{TagName: "valid"})
	return &PlatformValidator{validate: validate}
}

//ValidateStruct for the PlatformValidator
func (pv *PlatformValidator) ValidateStruct(s interface{}) error {

	if errs := pv.validate.Struct(s); errs != nil {
		return errors.New(errs.Error())
	}
	return nil
}

func (pv *PlatformValidator) RegisterValidation(key string, fn validator.Func) {
	pv.validate.RegisterValidation(key, fn)
}
