package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/log"
)

var _ = Describe("StandardObject", func() {
	It("NewStandardObject returns nil if context is nil", func() {
		value := map[string]interface{}{}
		Expect(validator.NewStandardObject(nil, "lich", &value)).To(BeNil())
	})

	Context("with context", func() {
		var standardContext *context.Standard

		BeforeEach(func() {
			var err error
			standardContext, err = context.NewStandard(log.NewNull())
			Expect(err).ToNot(HaveOccurred())
			Expect(standardContext).ToNot(BeNil())
		})

		Context("new validator with nil reference and nil value", func() {
			var standardObject *validator.StandardObject
			var result data.Object

			BeforeEach(func() {
				standardObject = validator.NewStandardObject(standardContext, nil, nil)
			})

			It("exists", func() {
				Expect(standardObject).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardObject.Exists()
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
					Expect(result).To(BeIdenticalTo(standardObject))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardObject.NotExists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObject))
				})
			})

			Context("Empty", func() {
				BeforeEach(func() {
					result = standardObject.Empty()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObject))
				})
			})

			Context("NotEmpty", func() {
				BeforeEach(func() {
					result = standardObject.NotEmpty()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObject))
				})
			})

		})

		Context("new validator with valid reference and an empty value", func() {
			var standardObject *validator.StandardObject
			var result data.Object

			BeforeEach(func() {
				value := map[string]interface{}{}
				standardObject = validator.NewStandardObject(standardContext, "lich", &value)
			})

			It("exists", func() {
				Expect(standardObject).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardObject.Exists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObject))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardObject.NotExists()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-exists"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value exists"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value exists"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/lich"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObject))
				})
			})

			Context("Empty", func() {
				BeforeEach(func() {
					result = standardObject.Empty()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObject))
				})
			})

			Context("NotEmpty", func() {
				BeforeEach(func() {
					result = standardObject.NotEmpty()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-empty"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is empty"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value is empty"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/lich"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObject))
				})
			})
		})

		Context("new validator with valid reference and an non-empty value", func() {
			var standardObject *validator.StandardObject
			var result data.Object

			BeforeEach(func() {
				value := map[string]interface{}{"a": "one", "b": "two"}
				standardObject = validator.NewStandardObject(standardContext, "lich", &value)
			})

			It("exists", func() {
				Expect(standardObject).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardObject.Exists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObject))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardObject.NotExists()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-exists"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value exists"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value exists"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/lich"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObject))
				})
			})

			Context("Empty", func() {
				BeforeEach(func() {
					result = standardObject.Empty()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-not-empty"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not empty"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value is not empty"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/lich"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObject))
				})
			})

			Context("NotEmpty", func() {
				BeforeEach(func() {
					result = standardObject.NotEmpty()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObject))
				})
			})

		})
	})
})
