package validate_test

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	validator "gopkg.in/bluesuncorp/validator.v8"

	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("Validate", func() {

	Context("using custom platform validator", func() {
		type ValidationTest struct {
			Offset int `json:"offset" valid:"custom"`
		}
		var (
			testValidator = func(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
				//a place holder for more through validation
				if val, ok := field.Interface().(int); ok {
					if val == 5 {
						return true
					}
				}
				return false
			}

			failureReasons = validate.FailureReasons{
				"Offset": validate.ValidationInfo{FieldName: "offset", Message: "Bad offset sorry"},
			}

			platformValidator = validate.NewPlatformValidator()

			processing validate.ErrorProcessing
		)

		BeforeEach(func() {
			platformValidator.RegisterValidation("custom", testValidator)
			platformValidator.SetFailureReasons(failureReasons)
			processing = validate.NewErrorProcessing("0")
		})

		Context("succeeds", func() {

			It("when offset match's expected value", func() {
				none := ValidationTest{Offset: 5}
				platformValidator.Struct(none, processing)
				Expect(processing.HasErrors()).To(BeFalse())
			})
		})

		Context("fails", func() {

			It("when offset doesn't match expected value", func() {
				none := ValidationTest{Offset: 0}
				platformValidator.Struct(none, processing)
				Expect(processing.HasErrors()).To(BeTrue())
			})

			It("gives meaningfull failure message", func() {
				none := ValidationTest{Offset: 0}
				platformValidator.Struct(none, processing)
				Expect(processing.HasErrors()).To(BeTrue())
				Expect(len(processing.GetErrors())).To(Equal(1))
				Expect(processing.GetErrors()[0].Detail).To(Equal("Bad offset sorry given '0'"))
			})
		})
	})
})
