package validate

import (
	"errors"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"
)

//Validator interface
type Validator interface {
	ValidateStruct(s interface{}) error
}

//PlatformValidator type that implements Validator
type PlatformValidator struct{}

//ValidateStruct for the PlatformValidator
func (platformValidator PlatformValidator) ValidateStruct(s interface{}) error {
	validate := validator.New(&validator.Config{TagName: "valid"})

	if errs := validate.Struct(s); errs != nil {
		return errors.New(errs.Error())
	}
	return nil
}
