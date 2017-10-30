package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/pointer"
	testStructure "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Adapter", func() {
	Context("ValidatableWithString", func() {
		var validatableWithString *testStructure.ValidatableWithString
		var str *string

		BeforeEach(func() {
			validatableWithString = testStructure.NewValidatableWithString()
			str = pointer.String(test.NewText(1, 32))
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
