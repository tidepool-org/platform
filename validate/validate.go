package validate

import (
	"fmt"

	validator "gopkg.in/bluesuncorp/validator.v8"
)

type Validator interface {
	Struct(s interface{}, errorProcessing ErrorProcessing)
	RegisterValidation(tag ValidationTag, fn validator.Func)
}

type PlatformValidator struct {
	validate *validator.Validate
	FailureReasons
}

type ValidationInfo struct {
	FieldName string
	Message   string
}

type ValidationTag string

type FailureReasons map[string]ValidationInfo

func NewPlatformValidator() *PlatformValidator {
	validate := validator.New(&validator.Config{TagName: "valid"})
	return &PlatformValidator{validate: validate}
}

func (pv *PlatformValidator) SetFailureReasons(reasons FailureReasons) *PlatformValidator {
	pv.FailureReasons = reasons
	return pv
}

func (pv *PlatformValidator) toErrorsArray(ve validator.ValidationErrors, errorProcessing ErrorProcessing) {
	for _, v := range ve {

		if reason, ok := pv.FailureReasons[v.Field]; ok {
			errorProcessing.AppendPointerError(
				reason.FieldName,
				"Validation Error",
				fmt.Sprintf("%s given '%v'", reason.Message, v.Value),
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
