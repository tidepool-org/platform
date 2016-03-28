package validate

import (
	"fmt"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"
)

type Validator interface {
	Struct(s interface{}, errorProcessing ErrorProcessing)
	RegisterValidation(tag ValidationTag, fn validator.Func)
}

type PlatformValidator struct {
	validate *validator.Validate
	reasons  ErrorReasons
}

type ValidationTag string

// ErrorReasons is a type of map[ValidationTag]string
// it allows us to map a ValidationTag to a reason why the validation failed
type ErrorReasons map[ValidationTag]string

func NewPlatformValidator() *PlatformValidator {
	validate := validator.New(&validator.Config{TagName: "valid"})
	return &PlatformValidator{validate: validate}
}

func (pv *PlatformValidator) SetErrorReasons(reasons ErrorReasons) *PlatformValidator {
	pv.reasons = reasons
	return pv
}

func (pv *PlatformValidator) toErrorsArray(ve validator.ValidationErrors, errorProcessing ErrorProcessing) {
	for _, v := range ve {
		if reason, ok := pv.reasons[ValidationTag(v.Tag)]; ok {
			errorProcessing.Append(NewPointerError(
				fmt.Sprintf("%s/%s", errorProcessing.BasePath, v.Type),
				"Validation Error",
				fmt.Sprintf("'%s' failed with '%s' when given '%v'", v.Field, reason, v.Value)),
			)
		}
	}
}

func (pv *PlatformValidator) Struct(s interface{}, errorProcessing ErrorProcessing) {
	validationErrors := pv.validate.Struct(s)
	if validationErrors != nil {
		pv.toErrorsArray(validationErrors.(validator.ValidationErrors), errorProcessing)
	}
}

func (pv *PlatformValidator) RegisterValidation(tag ValidationTag, fn validator.Func) {
	pv.validate.RegisterValidation(string(tag), fn)
}
