package validate_test

import (
	"reflect"

	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
	valid "github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"

	. "github.com/tidepool-org/platform/validate"
)

var _ = Describe("Validate", func() {

	Context("basic struct", func() {
		type testStruct struct {
			ID      string      `json:"id"  valid:"required"`
			Type    string      `json:"type"  valid:"required"`
			Offset  int         `json:"offset,omitempty"  valid:"omitempty,min=-10,max=10"`
			Payload interface{} `json:"payload,omitempty"  valid:"-"`
		}
		var (
			validator = NewPlatformValidator()
		)
		Context("is valid", func() {
			It("all feilds set correctly", func() {
				all := testStruct{ID: "someid", Type: "test", Offset: -9, Payload: map[string]string{"some": "stuff"}}
				Expect(validator.Struct(all)).To(BeNil())
			})
			It("only core feilds set correctly", func() {
				core := testStruct{ID: "someid", Type: "test"}
				Expect(validator.Struct(core)).To(BeNil())
			})
		})
		Context("is invalid when", func() {
			It("expected feilds aren't set", func() {
				none := testStruct{}
				Expect(validator.Struct(none)).ToNot(BeNil())
			})
			It("expected feilds are incorrectly set", func() {
				invalid := testStruct{ID: "", Type: ""}
				Expect(validator.Struct(invalid)).ToNot(BeNil())
			})
			It("optional feild set are incorrectly", func() {
				invalid := testStruct{ID: "someid", Type: "a_type", Offset: -100}
				Expect(validator.Struct(invalid)).ToNot(BeNil())
			})
		})
	})

	Context("nested struct", func() {
		type nestedUser struct {
			Firstname string `json:"firstname"  valid:"required"`
			Lastname  string `json:"lastname,omitempty"  valid:"omitempty,required"`
		}
		type testStruct struct {
			ID     string `json:"id"  valid:"required"`
			Offset int    `json:"offset,omitempty"  valid:"omitempty,required"`
			User   nestedUser
		}
		var (
			validator = NewPlatformValidator()
		)
		Context("is valid when", func() {
			It("all feilds set correctly", func() {
				all := testStruct{ID: "someid", Offset: 9, User: nestedUser{Firstname: "first", Lastname: "last"}}
				Expect(validator.Struct(all)).To(BeNil())
			})
			It("only core feilds set correctly", func() {
				core := testStruct{ID: "someid", User: nestedUser{Firstname: "first"}}
				Expect(validator.Struct(core)).To(BeNil())
			})
		})
		Context("is invalid when", func() {
			It("expected feilds aren't set", func() {
				none := testStruct{}
				Expect(validator.Struct(none)).ToNot(BeNil())
			})
			It("nested core feilds aren't set", func() {
				core := testStruct{ID: "someid", User: nestedUser{Firstname: ""}}
				Expect(validator.Struct(core)).ToNot(BeNil())
			})
			It("expected feilds are incorrectly set", func() {
				invalid := testStruct{ID: "", User: nestedUser{Firstname: ""}}
				Expect(validator.Struct(invalid)).ToNot(BeNil())
			})
		})
	})

	Context("custom validator", func() {
		var (
			validator = NewPlatformValidator()

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
		)
		type ValidationTest struct {
			Offset int `json:"offset" valid:"custom"`
		}

		BeforeEach(func() {
			validator.RegisterValidation("custom", testValidator)
		})

		It("fails struct validation ", func() {
			none := ValidationTest{Offset: 0}
			Expect(validator.Struct(none)).ToNot(BeNil())
		})
		It("fails struct validation gives meaningfull message", func() {
			none := ValidationTest{Offset: 0}
			errs := validator.Struct(none)
			if errs != nil {
				Expect(errs.GetError(validationFailureReasons).Error()).To(Equal("Error:Field validation for 'Offset' failed with 'Bad offset sorry' when given '0' for type 'int'"))
			}
		})
		It("fails field validation", func() {
			none := ValidationTest{Offset: 0}
			Expect(validator.Field(none.Offset, "custom")).ToNot(BeNil())
		})
		It("fails field validation gives meaningfull message", func() {
			none := ValidationTest{Offset: 99}
			errs := validator.Field(none.Offset, "custom")
			if errs != nil {
				Expect(errs.GetError(validationFailureReasons).Error()).To(Equal("Error:Field validation failed with 'Bad offset sorry' when given '99' for type 'int'"))
			}
		})
		It("passes field validation", func() {
			good := ValidationTest{Offset: 5}
			Expect(validator.Field(good.Offset, "custom")).To(BeNil())
		})
	})

})
