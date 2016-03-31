package validate_test

import (
	"reflect"

	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
	valid "github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"

	. "github.com/tidepool-org/platform/validate"
)

var _ = Describe("Validate", func() {

	Context("using custom validator", func() {
		type ValidationTest struct {
			Offset int `json:"offset" valid:"custom"`
		}
		var (
			testValidator = func(v *valid.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
				//a place holder for more through validation
				if val, ok := field.Interface().(int); ok {
					if val == 5 {
						return true
					}
				}
				return false
			}
			validationFailureReasons = ErrorReasons{
				"custom": "Bad offset sorry",
			}
			validator = NewPlatformValidator()

			processing = ErrorProcessing{BasePath: "0/validationtest", ErrorsArray: NewErrorsArray()}
		)

		BeforeEach(func() {
			validator.RegisterValidation("custom", testValidator)
			validator.SetErrorReasons(validationFailureReasons)
		})

		Context("succeeds", func() {

			It("when offset match's expected value", func() {
				none := ValidationTest{Offset: 5}
				validator.Struct(none, processing)
				Expect(processing.ErrorsArray.HasErrors()).To(BeFalse())
			})
		})

		Context("fails", func() {

			It("when offset doesn't match expected value", func() {
				none := ValidationTest{Offset: 0}
				validator.Struct(none, processing)
				Expect(processing.ErrorsArray.HasErrors()).To(BeTrue())
			})

			It("gives meaningfull failure message", func() {
				none := ValidationTest{Offset: 0}
				validator.Struct(none, processing)
				Expect(processing.ErrorsArray.HasErrors()).To(BeTrue())
				Expect(processing.ErrorsArray.Errors[0].Detail).To(ContainSubstring("'Offset' failed with 'Bad offset sorry' when given '0'"))
			})
		})

	})

})
