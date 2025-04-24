package validator_test

import (
	"math/rand"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/pointer"
	structureTest "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Adapter", func() {
	Context("ValidatableWithInt", func() {
		var validatableWithInt *structureTest.ValidatableWithInt
		var i *int

		BeforeEach(func() {
			validatableWithInt = structureTest.NewValidatableWithInt()
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
					Expect(validatableWithInt.ValidateInputs).To(Equal([]structureTest.ValidatableWithIntInput{{Validator: nil, Int: i}}))
				})
			})
		})
	})

	Context("ValidatableWithString", func() {
		var validatableWithString *structureTest.ValidatableWithString
		var str *string

		BeforeEach(func() {
			validatableWithString = structureTest.NewValidatableWithString()
			str = pointer.FromString(test.RandomStringFromRange(1, 32))
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
					Expect(validatableWithString.ValidateInputs).To(Equal([]structureTest.ValidatableWithStringInput{{Validator: nil, String: str}}))
				})
			})
		})
	})

	Context("ValidatableWithStringArray", func() {
		var validatableWithStringArray *structureTest.ValidatableWithStringArray
		var strArray *[]string

		BeforeEach(func() {
			validatableWithStringArray = structureTest.NewValidatableWithStringArray()
			strArray = &[]string{test.RandomStringFromRange(1, 32), test.RandomStringFromRange(1, 32), test.RandomStringFromRange(1, 32)}
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
					Expect(validatableWithStringArray.ValidateInputs).To(Equal([]structureTest.ValidatableWithStringArrayInput{{Validator: nil, StringArray: strArray}}))
				})
			})
		})
	})
})
