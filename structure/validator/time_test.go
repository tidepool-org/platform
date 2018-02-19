package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"

	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Time", func() {
	var base *structureBase.Base

	BeforeEach(func() {
		base = structureBase.New()
	})

	Context("NewTime", func() {
		It("returns successfully", func() {
			value := time.Now()
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
				result = validator.After(time.Unix(1451567655, 0).UTC())
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
				result = validator.Before(time.Unix(1451567655, 0).UTC())
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
	})

	Context("with new validator with value well into the past", func() {
		var validator *structureValidator.Time
		var result structure.Time
		var value time.Time

		BeforeEach(func() {
			var err error
			value, err = time.Parse("2006-01-02T15:04:05Z07:00", "1990-01-01T14:15:16Z")
			Expect(err).ToNot(HaveOccurred())
			Expect(value.IsZero()).ToNot(BeTrue())
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
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueTimeNotZero(value))
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
				result = validator.After(time.Unix(1451567655, 0).UTC())
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueTimeNotAfter(value, time.Unix(1451567655, 0).UTC()))
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
				result = validator.Before(time.Unix(1451567655, 0).UTC())
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
	})

	Context("with new validator with value well into the future", func() {
		var validator *structureValidator.Time
		var result structure.Time
		var value time.Time

		BeforeEach(func() {
			var err error
			value, err = time.Parse("2006-01-02T15:04:05Z07:00", "2090-01-01T14:15:16Z")
			Expect(err).ToNot(HaveOccurred())
			Expect(value.IsZero()).ToNot(BeTrue())
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
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueTimeNotZero(value))
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
				result = validator.After(time.Unix(1451567655, 0).UTC())
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
				result = validator.Before(time.Unix(1451567655, 0).UTC())
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueTimeNotBefore(value, time.Unix(1451567655, 0).UTC()))
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
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueTimeZero(value))
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
	})
})
