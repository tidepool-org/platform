package types

import "github.com/tidepool-org/platform/validate"

var _platformValidator = validate.NewPlatformValidator()

func GetPlatformValidator() *validate.PlatformValidator {
	return _platformValidator
}
