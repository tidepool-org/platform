package validate

import "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/asaskevich/govalidator"

//Validator interface
type Validator interface {
	ValidateStruct(s interface{}) (bool, error)
}

//PlatformValidator type that implements Validator
type PlatformValidator struct{}

//ValidateStruct for the PlatformValidator
func (platformValidator PlatformValidator) ValidateStruct(s interface{}) (bool, error) {
	return govalidator.ValidateStruct(s)
}
