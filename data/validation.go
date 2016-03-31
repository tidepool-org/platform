package data

import "github.com/tidepool-org/platform/validate"

var _platformValidator = validate.NewPlatformValidator()

func getPlatformValidator() *validate.PlatformValidator {
	return _platformValidator
}
