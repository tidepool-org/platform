package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/pvn/data"
	"github.com/tidepool-org/platform/pvn/data/context"
	"github.com/tidepool-org/platform/pvn/data/validator"
)

var _ = Describe("StandardInterface", func() {
	It("New returns nil if context is nil", func() {
		var value interface{} = "one"
		Expect(validator.NewStandardInterface(nil, "ghoul", &value)).To(BeNil())
	})

	Context("with context", func() {
		var standardContext *context.Standard

		BeforeEach(func() {
			standardContext = context.NewStandard()
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
					Expect(standardContext.Errors()[0].Code).To(Equal("value-does-not-exist"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value does not exist"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value does not exist"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/<nil>"))
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
		})
	})
})
