package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/structure"
	testStructure "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Bool", func() {
	var base *testStructure.Base

	BeforeEach(func() {
		base = testStructure.NewBase()
	})

	AfterEach(func() {
		base.Expectations()
	})

	Context("NewBool", func() {
		It("returns successfully", func() {
			value := true
			Expect(structureValidator.NewBool(base, &value)).ToNot(BeNil())
		})
	})

	Context("with new validator with nil value", func() {
		var validator *structureValidator.Bool
		var result structure.Bool

		BeforeEach(func() {
			validator = structureValidator.NewBool(base, nil)
			Expect(validator).ToNot(BeNil())
		})

		Context("Exists", func() {
			BeforeEach(func() {
				result = validator.Exists()
			})

			It("reports the expected error", func() {
				Expect(base.ReportErrorInputs).To(HaveLen(1))
				Expect(base.ReportErrorInputs[0]).To(MatchError("value does not exist"))
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
				Expect(base.ReportErrorInputs).To(BeEmpty())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("True", func() {
			BeforeEach(func() {
				result = validator.True()
			})

			It("does not report an error", func() {
				Expect(base.ReportErrorInputs).To(BeEmpty())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("False", func() {
			BeforeEach(func() {
				result = validator.False()
			})

			It("does not report an error", func() {
				Expect(base.ReportErrorInputs).To(BeEmpty())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})

	Context("with new validator with true value", func() {
		var validator *structureValidator.Bool
		var result structure.Bool
		var value bool

		BeforeEach(func() {
			value = true
			validator = structureValidator.NewBool(base, &value)
			Expect(validator).ToNot(BeNil())
		})

		Context("Exists", func() {
			BeforeEach(func() {
				result = validator.Exists()
			})

			It("does not report an error", func() {
				Expect(base.ReportErrorInputs).To(BeEmpty())
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
				Expect(base.ReportErrorInputs).To(HaveLen(1))
				Expect(base.ReportErrorInputs[0]).To(MatchError("value exists"))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("True", func() {
			BeforeEach(func() {
				result = validator.True()
			})

			It("does not report an error", func() {
				Expect(base.ReportErrorInputs).To(BeEmpty())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("False", func() {
			BeforeEach(func() {
				result = validator.False()
			})

			It("reports the expected error", func() {
				Expect(base.ReportErrorInputs).To(HaveLen(1))
				Expect(base.ReportErrorInputs[0]).To(MatchError("value is not false"))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})

	Context("with new validator with false value", func() {
		var validator *structureValidator.Bool
		var result structure.Bool
		var value bool

		BeforeEach(func() {
			value = false
			validator = structureValidator.NewBool(base, &value)
			Expect(validator).ToNot(BeNil())
		})

		Context("Exists", func() {
			BeforeEach(func() {
				result = validator.Exists()
			})

			It("does not report an error", func() {
				Expect(base.ReportErrorInputs).To(BeEmpty())
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
				Expect(base.ReportErrorInputs).To(HaveLen(1))
				Expect(base.ReportErrorInputs[0]).To(MatchError("value exists"))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("True", func() {
			BeforeEach(func() {
				result = validator.True()
			})

			It("reports the expected error", func() {
				Expect(base.ReportErrorInputs).To(HaveLen(1))
				Expect(base.ReportErrorInputs[0]).To(MatchError("value is not true"))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("False", func() {
			BeforeEach(func() {
				result = validator.False()
			})

			It("does not report an error", func() {
				Expect(base.ReportErrorInputs).To(BeEmpty())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})
})
