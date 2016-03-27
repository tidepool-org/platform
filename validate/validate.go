package validate

import (
	"fmt"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"
)

//Validator interface
type Validator interface {
	Struct(s interface{}, errorProcessing ErrorProcessing)
	//Field(field interface{}, tag ValidationTag) Errors
	RegisterValidation(tag ValidationTag, fn validator.Func)
}

//PlatformValidator type that implements Validator
type PlatformValidator struct {
	validate *validator.Validate
	reasons  ErrorReasons
}

//ValidationTag that all tags will be of this type
type ValidationTag string

// ErrorReasons is a type of map[ValidationTag]string
// it allows us to map a ValidationTag to a reason why the validation failed
type ErrorReasons map[ValidationTag]string

//NewPlatformValidator returns initialised PlatformValidator with custom tidepool validation
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
				fmt.Sprintf("'%s' failed with '%s' when given '%v' for type '%s' ", v.Field, reason, v.Value, v.Type)),
			)
		}
	}
}

//Struct validation for the PlatformValidator
func (pv *PlatformValidator) Struct(s interface{}, errorProcessing ErrorProcessing) {
	validationErrors := pv.validate.Struct(s)
	if validationErrors != nil {
		pv.toErrorsArray(validationErrors.(validator.ValidationErrors), errorProcessing)
	}
}

//RegisterValidation so we can add our own validation functions
func (pv *PlatformValidator) RegisterValidation(tag ValidationTag, fn validator.Func) {
	pv.validate.RegisterValidation(string(tag), fn)
}
