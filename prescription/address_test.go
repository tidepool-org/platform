package prescription_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/structure/validator"

	"github.com/tidepool-org/platform/prescription"
	"github.com/tidepool-org/platform/prescription/test"
)

var _ = Describe("Address", func() {
	Describe("Validate", func() {
		var address *prescription.Address
		var validate structure.Validator

		BeforeEach(func() {
			address = test.RandomAddress()
			validate = validator.New()
			Expect(validate.Validate(address)).ToNot(HaveOccurred())
		})

		It("fails when state is invalid", func() {
			address.State = "invalid"
			Expect(validate.Validate(address)).To(HaveOccurred())
		})

		It("fails when country is not 'US'", func() {
			address.Country = "BG"
			Expect(validate.Validate(address)).To(HaveOccurred())
		})

		It("doesn't fail when postal code is valid", func() {
			address.PostalCode = "12345"
			Expect(validate.Validate(address)).ToNot(HaveOccurred())
			address.PostalCode = "12345-1234"
			Expect(validate.Validate(address)).ToNot(HaveOccurred())
		})

		It("fails when postal code is valid", func() {
			address.PostalCode = "9000"
			Expect(validate.Validate(address)).To(HaveOccurred())
		})
	})

	Describe("ValidateAllRequired", func() {
		var address *prescription.Address
		var validate structure.Validator

		BeforeEach(func() {
			address = test.RandomAddress()
			validate = validator.New()
			address.ValidateAllRequired(validate)
			Expect(validate.Error()).ToNot(HaveOccurred())
		})

		It("fails when line1 is empty", func() {
			address.Line1 = ""
			address.ValidateAllRequired(validate)
			Expect(validate.Error()).To(HaveOccurred())
		})

		It("doesn't fail when line2 is empty", func() {
			address.Line2 = ""
			address.ValidateAllRequired(validate)
			Expect(validate.Error()).ToNot(HaveOccurred())
		})

		It("fails when state is empty", func() {
			address.State = ""
			address.ValidateAllRequired(validate)
			Expect(validate.Error()).To(HaveOccurred())
		})

		It("fails when city is empty", func() {
			address.City = ""
			address.ValidateAllRequired(validate)
			Expect(validate.Error()).To(HaveOccurred())
		})

		It("fails when country is empty", func() {
			address.Country = ""
			address.ValidateAllRequired(validate)
			Expect(validate.Error()).To(HaveOccurred())
		})

		It("fails when postal code is empty", func() {
			address.PostalCode = ""
			address.ValidateAllRequired(validate)
			Expect(validate.Error()).To(HaveOccurred())
		})
	})
})
