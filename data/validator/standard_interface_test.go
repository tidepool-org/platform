package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/log/test"
)

var _ = Describe("StandardInterface", func() {
	It("NewStandardInterface returns nil if context is nil", func() {
		var value interface{} = "one"
		Expect(validator.NewStandardInterface(nil, "ghoul", &value)).To(BeNil())
	})

	Context("with context", func() {
		var standardContext *context.Standard

		BeforeEach(func() {
			var err error
			standardContext, err = context.NewStandard(test.NewLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(standardContext).ToNot(BeNil())
		})

		Context("new validator with nil reference and nil value", func() {
			var standardInterface *validator.StandardInterface
			var result data.Interface

			BeforeEach(func() {
				standardInterface = validator.NewStandardInterface(standardContext, nil, nil)
			})

			It("exists", func() {
				Expect(standardInterface).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardInterface.Exists()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-not-exists"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value does not exist"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value does not exist"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/<nil>"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInterface))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardInterface.NotExists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInterface))
				})
			})
		})

		Context("new validator with valid reference and a value", func() {
			var standardInterface *validator.StandardInterface
			var result data.Interface

			BeforeEach(func() {
				var value interface{} = "one"
				standardInterface = validator.NewStandardInterface(standardContext, "ghoul", &value)
			})

			It("exists", func() {
				Expect(standardInterface).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardInterface.Exists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInterface))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardInterface.NotExists()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-exists"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value exists"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value exists"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/ghoul"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInterface))
				})
			})
		})
	})
})
