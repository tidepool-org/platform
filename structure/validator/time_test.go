package validator_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Time", func() {
	var base *structureBase.Base

	BeforeEach(func() {
		base = structureBase.New().WithSource(structure.NewPointerSource())
	})

	Context("NewTime", func() {
		It("returns successfully", func() {
			value := test.RandomTime()
			Expect(structureValidator.NewTime(base, &value)).ToNot(BeNil())
		})
	})

	Context("with new validator with nil value", func() {
		var validator *structureValidator.Time
		var result structure.Time

		BeforeEach(func() {
			validator = structureValidator.NewTime(base, nil)
			Expect(validator).ToNot(BeNil())
		})

		Context("Exists", func() {
			BeforeEach(func() {
				result = validator.Exists()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueNotExists())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotExists", func() {
			BeforeEach(func() {
				result = validator.NotExists()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Zero", func() {
			BeforeEach(func() {
				result = validator.Zero()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotZero", func() {
			BeforeEach(func() {
				result = validator.NotZero()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("After", func() {
			BeforeEach(func() {
				result = validator.After(test.PastNearTime())
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("AfterNow", func() {
			BeforeEach(func() {
				result = validator.AfterNow(0)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Before", func() {
			BeforeEach(func() {
				result = validator.Before(test.PastNearTime())
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("BeforeNow", func() {
			BeforeEach(func() {
				result = validator.BeforeNow(0)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Using", func() {
			BeforeEach(func() {
				result = validator.Using(func(value time.Time, errorReporter structure.ErrorReporter) {
					errorReporter.ReportError(structureValidator.ErrorValueExists())
				})
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})

	Context("with new validator with value well into the past", func() {
		var validator *structureValidator.Time
		var result structure.Time
		var value time.Time

		BeforeEach(func() {
			value = test.PastFarTime()
			validator = structureValidator.NewTime(base, &value)
			Expect(validator).ToNot(BeNil())
		})

		Context("Exists", func() {
			BeforeEach(func() {
				result = validator.Exists()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotExists", func() {
			BeforeEach(func() {
				result = validator.NotExists()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueExists())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Zero", func() {
			BeforeEach(func() {
				result = validator.Zero()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueNotEmpty())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotZero", func() {
			BeforeEach(func() {
				result = validator.NotZero()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("After", func() {
			BeforeEach(func() {
				result = validator.After(test.PastNearTime())
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueTimeNotAfter(value, test.PastNearTime()))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("AfterNow", func() {
			BeforeEach(func() {
				result = validator.AfterNow(0)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueTimeNotAfterNow(value))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Before", func() {
			BeforeEach(func() {
				result = validator.Before(test.PastNearTime())
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("BeforeNow", func() {
			BeforeEach(func() {
				result = validator.BeforeNow(0)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Using", func() {
			BeforeEach(func() {
				result = validator.Using(func(value time.Time, errorReporter structure.ErrorReporter) {
					Expect(value).To(Equal(value))
					errorReporter.ReportError(structureValidator.ErrorValueExists())
				})
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueExists())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Using (without func)", func() {
			BeforeEach(func() {
				result = validator.Using(nil)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})

	Context("with new validator with value well into the future", func() {
		var validator *structureValidator.Time
		var result structure.Time
		var value time.Time

		BeforeEach(func() {
			value = test.FutureFarTime()
			validator = structureValidator.NewTime(base, &value)
			Expect(validator).ToNot(BeNil())
		})

		Context("Exists", func() {
			BeforeEach(func() {
				result = validator.Exists()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotExists", func() {
			BeforeEach(func() {
				result = validator.NotExists()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueExists())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Zero", func() {
			BeforeEach(func() {
				result = validator.Zero()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueNotEmpty())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotZero", func() {
			BeforeEach(func() {
				result = validator.NotZero()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("After", func() {
			BeforeEach(func() {
				result = validator.After(test.PastNearTime())
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("AfterNow", func() {
			BeforeEach(func() {
				result = validator.AfterNow(0)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Before", func() {
			BeforeEach(func() {
				result = validator.Before(test.PastNearTime())
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueTimeNotBefore(value, test.PastNearTime()))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("BeforeNow", func() {
			BeforeEach(func() {
				result = validator.BeforeNow(0)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueTimeNotBeforeNow(value))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Using", func() {
			BeforeEach(func() {
				result = validator.Using(func(value time.Time, errorReporter structure.ErrorReporter) {
					Expect(value).To(Equal(value))
					errorReporter.ReportError(structureValidator.ErrorValueExists())
				})
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueExists())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Using (without func)", func() {
			BeforeEach(func() {
				result = validator.Using(nil)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})

	Context("with new validator with value zero", func() {
		var validator *structureValidator.Time
		var result structure.Time
		var value time.Time

		BeforeEach(func() {
			validator = structureValidator.NewTime(base, &value)
			Expect(validator).ToNot(BeNil())
		})

		Context("Zero", func() {
			BeforeEach(func() {
				result = validator.Zero()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotZero", func() {
			BeforeEach(func() {
				result = validator.NotZero()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueEmpty())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Using", func() {
			BeforeEach(func() {
				result = validator.Using(func(value time.Time, errorReporter structure.ErrorReporter) {
					Expect(value).To(Equal(value))
					errorReporter.ReportError(structureValidator.ErrorValueExists())
				})
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueExists())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Using (without func)", func() {
			BeforeEach(func() {
				result = validator.Using(nil)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})

	Context("with new validator with value now", func() {
		var validator *structureValidator.Time
		var result structure.Time
		var value time.Time

		BeforeEach(func() {
			value = time.Now().Add(-2 * time.Second)
			validator = structureValidator.NewTime(base, &value)
			Expect(validator).ToNot(BeNil())
		})

		Context("AfterNow with positive threshold", func() {
			BeforeEach(func() {
				result = validator.AfterNow(time.Minute)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("AfterNow with negative threshold", func() {
			BeforeEach(func() {
				result = validator.AfterNow(-time.Minute)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueTimeNotAfterNow(value))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("BeforeNow with positive threshold", func() {
			BeforeEach(func() {
				result = validator.BeforeNow(time.Minute)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("BeforeNow with negative threshold", func() {
			BeforeEach(func() {
				result = validator.BeforeNow(-time.Minute)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueTimeNotBeforeNow(value))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Using", func() {
			BeforeEach(func() {
				result = validator.Using(func(value time.Time, errorReporter structure.ErrorReporter) {
					Expect(value).To(Equal(value))
					errorReporter.ReportError(structureValidator.ErrorValueExists())
				})
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueExists())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Using (without func)", func() {
			BeforeEach(func() {
				result = validator.Using(nil)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})
})
