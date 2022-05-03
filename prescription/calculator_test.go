package prescription_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/structure/validator"

	"github.com/tidepool-org/platform/prescription"
	"github.com/tidepool-org/platform/prescription/test"
)

var _ = Describe("Calculator", func() {
	var calculator *prescription.Calculator
	var validate structure.Validator

	BeforeEach(func() {
		calculator = test.RandomCalculator()
		validate = validator.New()
		Expect(validate.Validate(calculator)).ToNot(HaveOccurred())
	})

	Describe("Validate", func() {
		BeforeEach(func() {
			validate = validator.New()
		})

		It("doesn't fail with nil method", func() {
			calculator.Method = nil
			Expect(validate.Validate(calculator)).ToNot(HaveOccurred())
		})

		It("fails with invalid method", func() {
			calculator.Method = pointer.FromString("invalidMethod")
			Expect(validate.Validate(calculator)).To(HaveOccurred())
		})

		Describe("method is 'totalDailyDose'", func() {
			BeforeEach(func() {
				calculator.Method = pointer.FromString(prescription.CalculatorMethodTotalDailyDose)
			})

			It("fails with empty 'totalDailyDose' input", func() {
				calculator.TotalDailyDose = nil
				Expect(validate.Validate(calculator)).To(HaveOccurred())
			})

			It("fails with invalid 'totalDailyDose' input", func() {
				calculator.TotalDailyDose = pointer.FromFloat64(-1.0)
				Expect(validate.Validate(calculator)).To(HaveOccurred())
			})

			It("fails with empty 'TotalDailyDoseScaleFactor' input", func() {
				calculator.TotalDailyDoseScaleFactor = nil
				Expect(validate.Validate(calculator)).To(HaveOccurred())
			})

			It("fails with invalid 'TotalDailyDoseScaleFactor' input", func() {
				calculator.TotalDailyDoseScaleFactor = pointer.FromFloat64(2)
				Expect(validate.Validate(calculator)).To(HaveOccurred())
			})
		})

		Describe("method is 'weight'", func() {
			BeforeEach(func() {
				calculator.Method = pointer.FromString(prescription.CalculatorMethodWeight)
			})

			It("fails with empty 'weight' input", func() {
				calculator.Weight = nil
				Expect(validate.Validate(calculator)).To(HaveOccurred())
			})

			It("fails with invalid 'weight' input", func() {
				calculator.Weight = pointer.FromFloat64(-1.0)
				Expect(validate.Validate(calculator)).To(HaveOccurred())
			})

			It("fails with empty 'weightUnits' input", func() {
				calculator.WeightUnits = nil
				Expect(validate.Validate(calculator)).To(HaveOccurred())
			})

			It("fails with invalid 'weightUnits' input", func() {
				calculator.WeightUnits = pointer.FromString("invalidUnits")
				Expect(validate.Validate(calculator)).To(HaveOccurred())
			})
		})

		Describe("method is 'totalDailyDoseAndWeight'", func() {
			BeforeEach(func() {
				calculator.Method = pointer.FromString(prescription.CalculatorMethodTotalDailyDoseAndWeight)
			})

			It("fails with empty 'totalDailyDose' input", func() {
				calculator.TotalDailyDose = nil
				Expect(validate.Validate(calculator)).To(HaveOccurred())
			})

			It("fails with invalid 'totalDailyDose' input", func() {
				calculator.TotalDailyDose = pointer.FromFloat64(-1.0)
				Expect(validate.Validate(calculator)).To(HaveOccurred())
			})

			It("fails with empty 'TotalDailyDoseScaleFactor' input", func() {
				calculator.TotalDailyDoseScaleFactor = nil
				Expect(validate.Validate(calculator)).To(HaveOccurred())
			})

			It("fails with invalid 'TotalDailyDoseScaleFactor' input", func() {
				calculator.TotalDailyDoseScaleFactor = pointer.FromFloat64(2)
				Expect(validate.Validate(calculator)).To(HaveOccurred())
			})
			It("fails with empty 'weight' input", func() {
				calculator.Weight = nil
				Expect(validate.Validate(calculator)).To(HaveOccurred())
			})

			It("fails with invalid 'weight' input", func() {
				calculator.Weight = pointer.FromFloat64(-1.0)
				Expect(validate.Validate(calculator)).To(HaveOccurred())
			})

			It("fails with empty 'weightUnits' input", func() {
				calculator.WeightUnits = nil
				Expect(validate.Validate(calculator)).To(HaveOccurred())
			})

			It("fails with invalid 'weightUnits' input", func() {
				calculator.WeightUnits = pointer.FromString("invalidUnits")
				Expect(validate.Validate(calculator)).To(HaveOccurred())
			})
		})
	})
})
