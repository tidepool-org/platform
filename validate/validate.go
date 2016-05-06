package validate

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"fmt"
	"reflect"

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
	Allowed   map[string]bool
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

func buildErrors(err *validator.FieldError, info ValidationInfo, errorProcessing ErrorProcessing) {

	//TODO: debug needs removing
	//fmt.Println("field namespace:", err.FieldNamespace, "name namespace:", err.NameNamespace)

	switch err.Kind {
	case reflect.Slice:
		if len(info.Allowed) > 0 {
			if actual, ok := err.Value.([]string); ok {
				for i := range actual {
					if _, ok := info.Allowed[actual[i]]; !ok {
						errorProcessing.AppendPointerError(
							fmt.Sprintf("%s/%d", info.FieldName, i),
							"Validation Error",
							fmt.Sprintf("%s given '%v'", info.Message, err.Value),
						)
					}
				}
			}
		} else {
			errorProcessing.AppendPointerError(info.FieldName, "Validation Error", info.Message)
		}
	default:
		errorProcessing.AppendPointerError(
			info.FieldName,
			"Validation Error",
			fmt.Sprintf("%s given '%v'", info.Message, err.Value),
		)
	}
}

func (pv *PlatformValidator) toErrorsArray(ve validator.ValidationErrors, errorProcessing ErrorProcessing) {
	for _, v := range ve {

		if reason, ok := pv.FailureReasons[v.NameNamespace]; ok {
			buildErrors(v, reason, errorProcessing)
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
