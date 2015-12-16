package validate

import "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/asaskevich/govalidator"

type Validator interface {
	ValidateStruct(s interface{}) (bool, error)
}

type PlatformValidator struct{}

func (this PlatformValidator) Validate(s interface{}) (bool, error) {
	return govalidator.ValidateStruct(s)
}
