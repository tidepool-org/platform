package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/log"
)

var _ = Describe("StandardBoolean", func() {
	It("NewStandardBoolean returns nil if context is nil", func() {
		value := false
		Expect(validator.NewStandardBoolean(nil, "zombie", &value)).To(BeNil())
	})

	Context("with context", func() {
		var standardContext *context.Standard

		BeforeEach(func() {
			var err error
			standardContext, err = context.NewStandard(log.NewNullLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(standardContext).ToNot(BeNil())
		})

		Context("new validator with nil reference and nil value", func() {
			var standardBoolean *validator.StandardBoolean
			var result data.Boolean

			BeforeEach(func() {
				standardBoolean = validator.NewStandardBoolean(standardContext, nil, nil)
			})

			It("exists", func() {
				Expect(standardBoolean).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardBoolean.Exists()
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
					Expect(result).To(BeIdenticalTo(standardBoolean))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardBoolean.NotExists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardBoolean))
				})
			})

			Context("True", func() {
				BeforeEach(func() {
					result = standardBoolean.True()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardBoolean))
				})
			})

			Context("False", func() {
				BeforeEach(func() {
					result = standardBoolean.False()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardBoolean))
				})
			})
		})

		Context("new validator with valid reference and true value", func() {
			var standardBoolean *validator.StandardBoolean
			var result data.Boolean

			BeforeEach(func() {
				value := true
				standardBoolean = validator.NewStandardBoolean(standardContext, "zombie", &value)
			})

			It("exists", func() {
				Expect(standardBoolean).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardBoolean.Exists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardBoolean))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardBoolean.NotExists()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-exists"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value exists"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value exists"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/zombie"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardBoolean))
				})
			})

			Context("True", func() {
				BeforeEach(func() {
					result = standardBoolean.True()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardBoolean))
				})
			})

			Context("False", func() {
				BeforeEach(func() {
					result = standardBoolean.False()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-not-false"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not false"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value is not false"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/zombie"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardBoolean))
				})
			})
		})

		Context("new validator with valid reference and false value", func() {
			var standardBoolean *validator.StandardBoolean
			var result data.Boolean

			BeforeEach(func() {
				value := false
				standardBoolean = validator.NewStandardBoolean(standardContext, "zombie", &value)
			})

			It("exists", func() {
				Expect(standardBoolean).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardBoolean.Exists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardBoolean))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardBoolean.NotExists()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-exists"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value exists"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value exists"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/zombie"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardBoolean))
				})
			})

			Context("True", func() {
				BeforeEach(func() {
					result = standardBoolean.True()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-not-true"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not true"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value is not true"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/zombie"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardBoolean))
				})
			})

			Context("False", func() {
				BeforeEach(func() {
					result = standardBoolean.False()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardBoolean))
				})
			})
		})
	})
})
