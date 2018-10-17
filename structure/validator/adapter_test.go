package validator_test

import (
	"math/rand"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/pointer"
	testStructure "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Adapter", func() {
	Context("ValidatableWithInt", func() {
		var validatableWithInt *testStructure.ValidatableWithInt
		var i *int

		BeforeEach(func() {
			validatableWithInt = testStructure.NewValidatableWithInt()
			i = pointer.FromInt(rand.Int())
		})

		Context("NewValidatableWithIntAdapter", func() {
			It("return successfully", func() {
				Expect(structureValidator.NewValidatableWithIntAdapter(validatableWithInt, i)).ToNot(BeNil())
			})
		})

		Context("with new validatable with int adapter", func() {
			var validatableWithIntAdapter *structureValidator.ValidatableWithIntAdapter

			BeforeEach(func() {
				validatableWithIntAdapter = structureValidator.NewValidatableWithIntAdapter(validatableWithInt, i)
				Expect(validatableWithIntAdapter).ToNot(BeNil())
			})

			Context("Validate", func() {
				It("returns successfully", func() {
					validatableWithIntAdapter.Validate(nil)
					Expect(validatableWithInt.ValidateInputs).To(Equal([]testStructure.ValidatableWithIntInput{{Validator: nil, Int: i}}))
				})
			})
		})
	})

	Context("ValidatableWithString", func() {
		var validatableWithString *testStructure.ValidatableWithString
		var str *string

		BeforeEach(func() {
			validatableWithString = testStructure.NewValidatableWithString()
			str = pointer.FromString(test.NewText(1, 32))
		})

		Context("NewValidatableWithStringAdapter", func() {
			It("return successfully", func() {
				Expect(structureValidator.NewValidatableWithStringAdapter(validatableWithString, str)).ToNot(BeNil())
			})
		})

		Context("with new validatable with string adapter", func() {
			var validatableWithStringAdapter *structureValidator.ValidatableWithStringAdapter

			BeforeEach(func() {
				validatableWithStringAdapter = structureValidator.NewValidatableWithStringAdapter(validatableWithString, str)
				Expect(validatableWithStringAdapter).ToNot(BeNil())
			})

			Context("Validate", func() {
				It("returns successfully", func() {
					validatableWithStringAdapter.Validate(nil)
					Expect(validatableWithString.ValidateInputs).To(Equal([]testStructure.ValidatableWithStringInput{{Validator: nil, String: str}}))
				})
			})
		})
	})

	Context("ValidatableWithStringArray", func() {
		var validatableWithStringArray *testStructure.ValidatableWithStringArray
		var strArray *[]string

		BeforeEach(func() {
			validatableWithStringArray = testStructure.NewValidatableWithStringArray()
			strArray = &[]string{test.NewText(1, 32), test.NewText(1, 32), test.NewText(1, 32)}
		})

		Context("NewValidatableWithStringArrayAdapter", func() {
			It("return successfully", func() {
				Expect(structureValidator.NewValidatableWithStringArrayAdapter(validatableWithStringArray, strArray)).ToNot(BeNil())
			})
		})

		Context("with new validatable with string array adapter", func() {
			var validatableWithStringArrayAdapter *structureValidator.ValidatableWithStringArrayAdapter

			BeforeEach(func() {
				validatableWithStringArrayAdapter = structureValidator.NewValidatableWithStringArrayAdapter(validatableWithStringArray, strArray)
				Expect(validatableWithStringArrayAdapter).ToNot(BeNil())
			})

			Context("Validate", func() {
				It("returns successfully", func() {
					validatableWithStringArrayAdapter.Validate(nil)
					Expect(validatableWithStringArray.ValidateInputs).To(Equal([]testStructure.ValidatableWithStringArrayInput{{Validator: nil, StringArray: strArray}}))
				})
			})
		})
	})
})
