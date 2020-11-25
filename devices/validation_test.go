package devices_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tidepool-org/devices/api"

	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/devices"
	"github.com/tidepool-org/platform/devices/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Validation", func() {
	Context("ValidateBasalRateSchedule", func() {
		var guardRail *api.BasalRatesGuardRail
		var validator *structureValidator.Validator

		BeforeEach(func() {
			guardRail = test.NewBasalRateGuardRail()
			validator = structureValidator.New()
		})

		It("doesn't return error with a single valid value", func() {
			var schedule pump.BasalRateStartArray = []*pump.BasalRateStart{
				{Rate: pointer.FromFloat64(0.55)},
			}
			devices.ValidateBasalRateSchedule(schedule, guardRail, validator)
			Expect(validator.Error()).To(BeNil())
		})

		It("doesn't return error with multiple valid values", func() {
			var schedule pump.BasalRateStartArray = []*pump.BasalRateStart{
				{Rate: pointer.FromFloat64(0.55)},
				{Rate: pointer.FromFloat64(15.55)},
			}
			devices.ValidateBasalRateSchedule(schedule, guardRail, validator)
			Expect(validator.Error()).To(BeNil())
		})

		It("returns an error with an invalid value", func() {
			var schedule pump.BasalRateStartArray = []*pump.BasalRateStart{
				{Rate: pointer.FromFloat64(0.55)},
				{Rate: pointer.FromFloat64(0.56)},
				{Rate: pointer.FromFloat64(15.55)},
			}
			expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/rate")
			devices.ValidateBasalRateSchedule(schedule, guardRail, validator)
			errorsTest.ExpectEqual(validator.Error(), expected)
		})
	})

	Context("ValidateGlucoseSafetyLimit", func() {
		var guardRail *api.GlucoseSafetyLimitGuardRail
		var validator *structureValidator.Validator

		BeforeEach(func() {
			guardRail = test.NewGlucoseSafetyLimitGuardRail()
			validator = structureValidator.New()
		})

		It("doesn't return error with a valid value", func() {
			suspendThreshold := pointer.FromFloat64(70)
			devices.ValidateGlucoseSafetyLimit(suspendThreshold, guardRail, validator)
			Expect(validator.Error()).To(BeNil())
		})

		It("returns an error with a value outside of the range", func() {
			suspendThreshold := pointer.FromFloat64(190)
			expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "")
			devices.ValidateGlucoseSafetyLimit(suspendThreshold, guardRail, validator)
			errorsTest.ExpectEqual(validator.Error(), expected)
		})

		It("returns an error with a fractional value", func() {
			suspendThreshold := pointer.FromFloat64(70.5)
			expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "")
			devices.ValidateGlucoseSafetyLimit(suspendThreshold, guardRail, validator)
			errorsTest.ExpectEqual(validator.Error(), expected)
		})
	})

	Context("ValidateBloodGlucoseTargetSchedule", func() {
		var guardRail *api.CorrectionRangeGuardRail

		BeforeEach(func() {
			guardRail = test.NewCorrectionRangeGuardRail()
		})

		Context("High", func() {
			var validator *structureValidator.Validator

			BeforeEach(func() {
				validator = structureValidator.New()
			})

			It("doesn't return error with a single valid value", func() {
				var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
					{Target: glucose.Target{High: pointer.FromFloat64(90)}},
				}
				devices.ValidateBloodGlucoseTargetSchedule(schedule, guardRail, validator)
				Expect(validator.Error()).To(BeNil())
			})

			It("doesn't return error with multiple valid values", func() {
				var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
					{Target: glucose.Target{High: pointer.FromFloat64(90)}},
					{Target: glucose.Target{High: pointer.FromFloat64(100)}},
				}
				devices.ValidateBloodGlucoseTargetSchedule(schedule, guardRail, validator)
				Expect(validator.Error()).To(BeNil())
			})

			It("returns an error with an invalid value", func() {
				var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
					{Target: glucose.Target{High: pointer.FromFloat64(90)}},
					{Target: glucose.Target{High: pointer.FromFloat64(90.25)}},
					{Target: glucose.Target{High: pointer.FromFloat64(100)}},
				}
				expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/high")
				devices.ValidateBloodGlucoseTargetSchedule(schedule, guardRail, validator)
				errorsTest.ExpectEqual(validator.Error(), expected)
			})
		})

		Context("Low", func() {
			var validator *structureValidator.Validator

			BeforeEach(func() {
				validator = structureValidator.New()
			})

			It("doesn't return error with a single valid value", func() {
				var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
					{Target: glucose.Target{Low: pointer.FromFloat64(90)}},
				}
				devices.ValidateBloodGlucoseTargetSchedule(schedule, guardRail, validator)
				Expect(validator.Error()).To(BeNil())
			})

			It("doesn't return error with multiple valid values", func() {
				var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
					{Target: glucose.Target{Low: pointer.FromFloat64(90)}},
					{Target: glucose.Target{Low: pointer.FromFloat64(100)}},
				}
				devices.ValidateBloodGlucoseTargetSchedule(schedule, guardRail, validator)
				Expect(validator.Error()).To(BeNil())
			})

			It("returns an error with an invalid value", func() {
				var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
					{Target: glucose.Target{Low: pointer.FromFloat64(90)}},
					{Target: glucose.Target{Low: pointer.FromFloat64(90.25)}},
					{Target: glucose.Target{Low: pointer.FromFloat64(100)}},
				}
				expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/low")
				devices.ValidateBloodGlucoseTargetSchedule(schedule, guardRail, validator)
				errorsTest.ExpectEqual(validator.Error(), expected)
			})
		})

		Context("Range", func() {
			var validator *structureValidator.Validator

			BeforeEach(func() {
				validator = structureValidator.New()
			})

			It("doesn't return error with a single valid value", func() {
				var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
					{Target: glucose.Target{Range: pointer.FromFloat64(90)}},
				}
				devices.ValidateBloodGlucoseTargetSchedule(schedule, guardRail, validator)
				Expect(validator.Error()).To(BeNil())
			})

			It("doesn't return error with multiple valid values", func() {
				var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
					{Target: glucose.Target{Range: pointer.FromFloat64(90)}},
					{Target: glucose.Target{Range: pointer.FromFloat64(100)}},
				}
				devices.ValidateBloodGlucoseTargetSchedule(schedule, guardRail, validator)
				Expect(validator.Error()).To(BeNil())
			})

			It("returns an error with an invalid value", func() {
				var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
					{Target: glucose.Target{Range: pointer.FromFloat64(90)}},
					{Target: glucose.Target{Range: pointer.FromFloat64(90.25)}},
					{Target: glucose.Target{Range: pointer.FromFloat64(100)}},
				}
				expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/range")
				devices.ValidateBloodGlucoseTargetSchedule(schedule, guardRail, validator)
				errorsTest.ExpectEqual(validator.Error(), expected)
			})
		})

		Context("Target", func() {
			var validator *structureValidator.Validator

			BeforeEach(func() {
				validator = structureValidator.New()
			})

			It("doesn't return error with a single valid value", func() {
				var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
					{Target: glucose.Target{Target: pointer.FromFloat64(90)}},
				}
				devices.ValidateBloodGlucoseTargetSchedule(schedule, guardRail, validator)
				Expect(validator.Error()).To(BeNil())
			})

			It("doesn't return error with multiple valid values", func() {
				var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
					{Target: glucose.Target{Target: pointer.FromFloat64(90)}},
					{Target: glucose.Target{Target: pointer.FromFloat64(100)}},
				}
				devices.ValidateBloodGlucoseTargetSchedule(schedule, guardRail, validator)
				Expect(validator.Error()).To(BeNil())
			})

			It("returns an error with an invalid value", func() {
				var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
					{Target: glucose.Target{Target: pointer.FromFloat64(90)}},
					{Target: glucose.Target{Target: pointer.FromFloat64(90.25)}},
					{Target: glucose.Target{Target: pointer.FromFloat64(100)}},
				}
				expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/target")
				devices.ValidateBloodGlucoseTargetSchedule(schedule, guardRail, validator)
				errorsTest.ExpectEqual(validator.Error(), expected)
			})
		})
	})

	Context("ValidateCarbohydrateRatioSchedule", func() {
		var guardRail *api.CarbohydrateRatioGuardRail
		var validator *structureValidator.Validator

		BeforeEach(func() {
			guardRail = test.NewCarbohydrateRatioGuardRail()
			validator = structureValidator.New()
		})

		It("doesn't return error with a single valid value", func() {
			var schedule pump.CarbohydrateRatioStartArray = []*pump.CarbohydrateRatioStart{
				{Amount: pointer.FromFloat64(120.01)},
			}
			devices.ValidateCarbohydrateRatioSchedule(schedule, guardRail, validator)
			Expect(validator.Error()).To(BeNil())
		})

		It("doesn't return error with multiple valid values", func() {
			var schedule pump.CarbohydrateRatioStartArray = []*pump.CarbohydrateRatioStart{
				{Amount: pointer.FromFloat64(120.01)},
				{Amount: pointer.FromFloat64(10.00)},
			}
			devices.ValidateCarbohydrateRatioSchedule(schedule, guardRail, validator)
			Expect(validator.Error()).To(BeNil())
		})

		It("returns an error with an invalid value", func() {
			var schedule pump.CarbohydrateRatioStartArray = []*pump.CarbohydrateRatioStart{
				{Amount: pointer.FromFloat64(120.01)},
				{Amount: pointer.FromFloat64(0.99)},
				{Amount: pointer.FromFloat64(10.00)},
			}
			expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/amount")
			devices.ValidateCarbohydrateRatioSchedule(schedule, guardRail, validator)
			errorsTest.ExpectEqual(validator.Error(), expected)
		})
	})

	Context("ValidateInsulinSensitivitySchedule", func() {
		var guardRail *api.InsulinSensitivityGuardRail
		var validator *structureValidator.Validator

		BeforeEach(func() {
			guardRail = test.NewInsulinSensitivityGuardRail()
			validator = structureValidator.New()
		})

		It("doesn't return error with a single valid value", func() {
			var schedule pump.InsulinSensitivityStartArray = []*pump.InsulinSensitivityStart{
				{Amount: pointer.FromFloat64(120.00)},
			}
			devices.ValidateInsulinSensitivitySchedule(schedule, guardRail, validator)
			Expect(validator.Error()).To(BeNil())
		})

		It("doesn't return error with multiple valid values", func() {
			var schedule pump.InsulinSensitivityStartArray = []*pump.InsulinSensitivityStart{
				{Amount: pointer.FromFloat64(120.00)},
				{Amount: pointer.FromFloat64(10.00)},
			}
			devices.ValidateInsulinSensitivitySchedule(schedule, guardRail, validator)
			Expect(validator.Error()).To(BeNil())
		})

		It("returns an error with an invalid value", func() {
			var schedule pump.InsulinSensitivityStartArray = []*pump.InsulinSensitivityStart{
				{Amount: pointer.FromFloat64(120.00)},
				{Amount: pointer.FromFloat64(120.5)},
				{Amount: pointer.FromFloat64(10.00)},
			}
			expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/amount")
			devices.ValidateInsulinSensitivitySchedule(schedule, guardRail, validator)
			errorsTest.ExpectEqual(validator.Error(), expected)
		})
	})

	Context("ValidateBasalRateMaximum", func() {
		var guardRail *api.BasalRateMaximumGuardRail
		var validator *structureValidator.Validator

		BeforeEach(func() {
			guardRail = test.NewBasalRateMaximumGuardRail()
			validator = structureValidator.New()
		})

		It("doesn't return error with a valid value", func() {
			var value = pump.BasalRateMaximum{
				Value: pointer.FromFloat64(1.00),
			}
			devices.ValidateBasalRateMaximum(value, guardRail, validator)
			Expect(validator.Error()).To(BeNil())
		})

		It("returns an error with an invalid value", func() {
			var value = pump.BasalRateMaximum{
				Value: pointer.FromFloat64(31.00),
			}
			expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/value")
			devices.ValidateBasalRateMaximum(value, guardRail, validator)
			errorsTest.ExpectEqual(validator.Error(), expected)
		})
	})

	Context("ValidateBolusAmountMaximum", func() {
		var guardRail *api.BolusAmountMaximumGuardRail
		var validator *structureValidator.Validator

		BeforeEach(func() {
			guardRail = test.NewBolusAmountMaximumGuardRail()
			validator = structureValidator.New()
		})

		It("doesn't return error with a valid value", func() {
			var value = pump.BolusAmountMaximum{
				Value: pointer.FromFloat64(1.00),
			}
			devices.ValidateBolusAmountMaximum(value, guardRail, validator)
			Expect(validator.Error()).To(BeNil())
		})

		It("returns an error with an invalid value", func() {
			var value = pump.BolusAmountMaximum{
				Value: pointer.FromFloat64(31.00),
			}
			expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/value")
			devices.ValidateBolusAmountMaximum(value, guardRail, validator)
			errorsTest.ExpectEqual(validator.Error(), expected)
		})
	})
})
