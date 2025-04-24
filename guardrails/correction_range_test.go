package guardrails_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/devices/api"

	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/guardrails"
	"github.com/tidepool-org/platform/guardrails/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Correction Range", func() {

	Context("Validate", func() {
		var guardRail *api.CorrectionRangeGuardRail
		var glucoseSafetyLimit *float64

		BeforeEach(func() {
			guardRail = test.NewCorrectionRangeGuardRail()
			glucoseSafetyLimit = nil
		})

		Context("Schedule", func() {
			Context("With Low/High values", func() {
				var validator *structureValidator.Validator

				BeforeEach(func() {
					validator = structureValidator.New(logTest.NewLogger())
				})

				It("doesn't return error with a single valid value", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Low: pointer.FromFloat64(90), High: pointer.FromFloat64(91)}},
					}
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					Expect(validator.Error()).To(BeNil())
				})

				It("doesn't return error with multiple valid values", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Low: pointer.FromFloat64(90), High: pointer.FromFloat64(91)}},
						{Target: glucose.Target{Low: pointer.FromFloat64(100), High: pointer.FromFloat64(105)}},
					}
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					Expect(validator.Error()).To(BeNil())
				})

				It("returns an error with an invalid increment", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Low: pointer.FromFloat64(90), High: pointer.FromFloat64(91)}},
						{Target: glucose.Target{Low: pointer.FromFloat64(92.25), High: pointer.FromFloat64(93)}},
						{Target: glucose.Target{Low: pointer.FromFloat64(100), High: pointer.FromFloat64(105)}},
					}
					expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/low")
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					errorsTest.ExpectEqual(validator.Error(), expected)
				})

				It("returns an error with a value that's below the minimum absolute bounds", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Low: pointer.FromFloat64(86), High: pointer.FromFloat64(91)}},
						{Target: glucose.Target{Low: pointer.FromFloat64(100), High: pointer.FromFloat64(105)}},
					}
					expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/0/low")
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					errorsTest.ExpectEqual(validator.Error(), expected)
				})

				It("returns an error with a value that's above the maximum absolute bounds", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Low: pointer.FromFloat64(87), High: pointer.FromFloat64(91)}},
						{Target: glucose.Target{Low: pointer.FromFloat64(100), High: pointer.FromFloat64(181)}},
					}
					expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/high")
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					errorsTest.ExpectEqual(validator.Error(), expected)
				})

				It("returns an error for each value below the glucose safety limit", func() {
					glucoseSafetyLimit = pointer.FromFloat64(101)
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Low: pointer.FromFloat64(90), High: pointer.FromFloat64(91)}},
						{Target: glucose.Target{Low: pointer.FromFloat64(100), High: pointer.FromFloat64(105)}},
					}
					expected := []error{
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/0/low"),
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/0/high"),
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/low"),
					}
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					errorsTest.ExpectEqual(validator.Error(), expected...)
				})

				It("returns an error when the number of segments is higher than the guardrail", func() {
					maxSegments := int32(2)
					guardRail.MaxSegments = &maxSegments
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Low: pointer.FromFloat64(90), High: pointer.FromFloat64(91)}},
						{Target: glucose.Target{Low: pointer.FromFloat64(100), High: pointer.FromFloat64(105)}},
						{Target: glucose.Target{Low: pointer.FromFloat64(101), High: pointer.FromFloat64(104)}},
					}
					expected := errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(3, 2), "")
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					errorsTest.ExpectEqual(validator.Error(), expected)
				})
			})

			Context("With Target values", func() {
				var validator *structureValidator.Validator

				BeforeEach(func() {
					validator = structureValidator.New(logTest.NewLogger())
				})

				It("doesn't return error with a single valid value", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Target: pointer.FromFloat64(90)}},
					}
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					Expect(validator.Error()).To(BeNil())
				})

				It("doesn't return error with multiple valid values", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Target: pointer.FromFloat64(90)}},
						{Target: glucose.Target{Target: pointer.FromFloat64(105)}},
					}
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					Expect(validator.Error()).To(BeNil())
				})

				It("returns an error with an invalid increment", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Target: pointer.FromFloat64(90)}},
						{Target: glucose.Target{Target: pointer.FromFloat64(92.25)}},
						{Target: glucose.Target{Target: pointer.FromFloat64(100)}},
					}
					expected := []error{
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/low"),
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/high"),
					}
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					errorsTest.ExpectEqual(validator.Error(), expected...)
				})

				It("returns an error with a value that's below the minimum absolute bounds", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Target: pointer.FromFloat64(86)}},
						{Target: glucose.Target{Target: pointer.FromFloat64(90)}},
						{Target: glucose.Target{Target: pointer.FromFloat64(100)}},
					}
					expected := []error{
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/0/low"),
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/0/high"),
					}
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					errorsTest.ExpectEqual(validator.Error(), expected...)
				})

				It("returns an error with a value that's above the maximum absolute bounds", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Target: pointer.FromFloat64(90)}},
						{Target: glucose.Target{Target: pointer.FromFloat64(181)}},
					}
					expected := []error{
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/low"),
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/high"),
					}
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					errorsTest.ExpectEqual(validator.Error(), expected...)
				})
			})

			Context("With Target/High values", func() {
				var validator *structureValidator.Validator

				BeforeEach(func() {
					validator = structureValidator.New(logTest.NewLogger())
				})

				It("doesn't return error with a single valid value", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Target: pointer.FromFloat64(90), High: pointer.FromFloat64(91)}},
					}
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					Expect(validator.Error()).To(BeNil())
				})

				It("doesn't return error with multiple valid values", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Target: pointer.FromFloat64(90), High: pointer.FromFloat64(91)}},
						{Target: glucose.Target{Target: pointer.FromFloat64(105), High: pointer.FromFloat64(106)}},
					}
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					Expect(validator.Error()).To(BeNil())
				})

				It("returns an error with an invalid target increment", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Target: pointer.FromFloat64(90), High: pointer.FromFloat64(91)}},
						{Target: glucose.Target{Target: pointer.FromFloat64(92.25), High: pointer.FromFloat64(91)}},
						{Target: glucose.Target{Target: pointer.FromFloat64(100), High: pointer.FromFloat64(101)}},
					}
					expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/low")
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					errorsTest.ExpectEqual(validator.Error(), expected)
				})

				It("returns an error with an invalid high increment", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Target: pointer.FromFloat64(90), High: pointer.FromFloat64(91)}},
						{Target: glucose.Target{Target: pointer.FromFloat64(92), High: pointer.FromFloat64(92.25)}},
						{Target: glucose.Target{Target: pointer.FromFloat64(100), High: pointer.FromFloat64(101)}},
					}
					expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/high")
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					errorsTest.ExpectEqual(validator.Error(), expected)
				})

				It("returns an error with a target value that's below the minimum absolute bounds", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Target: pointer.FromFloat64(86), High: pointer.FromFloat64(91)}},
						{Target: glucose.Target{Target: pointer.FromFloat64(100), High: pointer.FromFloat64(101)}},
					}
					expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/0/low")
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					errorsTest.ExpectEqual(validator.Error(), expected)
				})

				It("returns an error with a high value that's above the maximum absolute bounds", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Target: pointer.FromFloat64(87), High: pointer.FromFloat64(91)}},
						{Target: glucose.Target{Target: pointer.FromFloat64(100), High: pointer.FromFloat64(181)}},
					}
					expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/high")
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					errorsTest.ExpectEqual(validator.Error(), expected)
				})
			})

			Context("With Target/Range values", func() {
				var validator *structureValidator.Validator

				BeforeEach(func() {
					validator = structureValidator.New(logTest.NewLogger())
				})

				It("doesn't return error with a single valid value", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Target: pointer.FromFloat64(90), Range: pointer.FromFloat64(1)}},
					}
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					Expect(validator.Error()).To(BeNil())
				})

				It("doesn't return error with multiple valid values", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Target: pointer.FromFloat64(90), Range: pointer.FromFloat64(1)}},
						{Target: glucose.Target{Target: pointer.FromFloat64(105), Range: pointer.FromFloat64(1)}},
					}
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					Expect(validator.Error()).To(BeNil())
				})

				It("returns an error with an invalid target increment", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Target: pointer.FromFloat64(90), Range: pointer.FromFloat64(1)}},
						{Target: glucose.Target{Target: pointer.FromFloat64(92.25), Range: pointer.FromFloat64(1)}},
						{Target: glucose.Target{Target: pointer.FromFloat64(100), Range: pointer.FromFloat64(1)}},
					}
					expected := []error{
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/low"),
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/high"),
					}
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					errorsTest.ExpectEqual(validator.Error(), expected...)
				})

				It("returns an error with an invalid range increment", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Target: pointer.FromFloat64(90), Range: pointer.FromFloat64(1)}},
						{Target: glucose.Target{Target: pointer.FromFloat64(92), Range: pointer.FromFloat64(1.25)}},
						{Target: glucose.Target{Target: pointer.FromFloat64(100), Range: pointer.FromFloat64(1)}},
					}
					expected := []error{
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/low"),
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/high"),
					}
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					errorsTest.ExpectEqual(validator.Error(), expected...)
				})

				It("returns an error when the lower bound is below the minimum absolute bounds", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Target: pointer.FromFloat64(87), Range: pointer.FromFloat64(1)}},
						{Target: glucose.Target{Target: pointer.FromFloat64(100), Range: pointer.FromFloat64(1)}},
					}
					expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/0/low")
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					errorsTest.ExpectEqual(validator.Error(), expected)
				})

				It("returns an error when the upper bound is above the maximum absolute bounds", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Target: pointer.FromFloat64(90), Range: pointer.FromFloat64(1)}},
						{Target: glucose.Target{Target: pointer.FromFloat64(180), Range: pointer.FromFloat64(1)}},
					}
					expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/high")
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					errorsTest.ExpectEqual(validator.Error(), expected)
				})
			})
		})

		Context("Preprandial", func() {
			Context("With Low/High values", func() {
				var validator *structureValidator.Validator

				BeforeEach(func() {
					validator = structureValidator.New(logTest.NewLogger())
					guardRail = test.NewPremealCorrectionRangeGuardRail()
				})

				It("doesn't return error with a single valid value", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Low: pointer.FromFloat64(90), High: pointer.FromFloat64(91)}},
					}
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					Expect(validator.Error()).To(BeNil())
				})

				It("doesn't return error with multiple valid values", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Low: pointer.FromFloat64(90), High: pointer.FromFloat64(91)}},
						{Target: glucose.Target{Low: pointer.FromFloat64(100), High: pointer.FromFloat64(105)}},
					}
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					Expect(validator.Error()).To(BeNil())
				})

				It("returns an error with an invalid increment", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Low: pointer.FromFloat64(90), High: pointer.FromFloat64(91)}},
						{Target: glucose.Target{Low: pointer.FromFloat64(92.25), High: pointer.FromFloat64(93)}},
						{Target: glucose.Target{Low: pointer.FromFloat64(100), High: pointer.FromFloat64(105)}},
					}
					expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/low")
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					errorsTest.ExpectEqual(validator.Error(), expected)
				})

				It("returns an error with a value that's below the minimum absolute bounds", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Low: pointer.FromFloat64(66), High: pointer.FromFloat64(91)}},
						{Target: glucose.Target{Low: pointer.FromFloat64(100), High: pointer.FromFloat64(105)}},
					}
					expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/0/low")
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					errorsTest.ExpectEqual(validator.Error(), expected)
				})

				It("returns an error with a value that's above the maximum absolute bounds", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Low: pointer.FromFloat64(87), High: pointer.FromFloat64(91)}},
						{Target: glucose.Target{Low: pointer.FromFloat64(100), High: pointer.FromFloat64(131)}},
					}
					expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/high")
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					errorsTest.ExpectEqual(validator.Error(), expected)
				})

				It("returns errors for each value below the glucose safety limit", func() {
					glucoseSafetyLimit = pointer.FromFloat64(102)
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Low: pointer.FromFloat64(90), High: pointer.FromFloat64(91)}},
						{Target: glucose.Target{Low: pointer.FromFloat64(100), High: pointer.FromFloat64(105)}},
					}
					expected := []error{
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/0/low"),
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/0/high"),
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/low"),
					}
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					errorsTest.ExpectEqual(validator.Error(), expected...)
				})
			})
		})

		Context("PhysicalActivity", func() {
			Context("With Low/High values", func() {
				var validator *structureValidator.Validator

				BeforeEach(func() {
					validator = structureValidator.New(logTest.NewLogger())
					guardRail = test.NewWorkoutCorrectionRangeGuardRail()
				})

				It("doesn't return error with a single valid value", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Low: pointer.FromFloat64(87), High: pointer.FromFloat64(250)}},
					}
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					Expect(validator.Error()).To(BeNil())
				})

				It("doesn't return error with multiple valid values", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Low: pointer.FromFloat64(90), High: pointer.FromFloat64(91)}},
						{Target: glucose.Target{Low: pointer.FromFloat64(100), High: pointer.FromFloat64(105)}},
					}
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					Expect(validator.Error()).To(BeNil())
				})

				It("returns an error with an invalid increment", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Low: pointer.FromFloat64(90), High: pointer.FromFloat64(91)}},
						{Target: glucose.Target{Low: pointer.FromFloat64(92.25), High: pointer.FromFloat64(93)}},
						{Target: glucose.Target{Low: pointer.FromFloat64(100), High: pointer.FromFloat64(105)}},
					}
					expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/low")
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					errorsTest.ExpectEqual(validator.Error(), expected)
				})

				It("returns an error with a value that's below the minimum absolute bounds", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Low: pointer.FromFloat64(86), High: pointer.FromFloat64(91)}},
						{Target: glucose.Target{Low: pointer.FromFloat64(100), High: pointer.FromFloat64(105)}},
					}
					expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/0/low")
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					errorsTest.ExpectEqual(validator.Error(), expected)
				})

				It("returns an error with a value that's above the maximum absolute bounds", func() {
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Low: pointer.FromFloat64(87), High: pointer.FromFloat64(91)}},
						{Target: glucose.Target{Low: pointer.FromFloat64(100), High: pointer.FromFloat64(251)}},
					}
					expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/high")
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					errorsTest.ExpectEqual(validator.Error(), expected)
				})

				It("returns errors for each value below the glucose safety limit", func() {
					glucoseSafetyLimit = pointer.FromFloat64(102)
					var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
						{Target: glucose.Target{Low: pointer.FromFloat64(90), High: pointer.FromFloat64(91)}},
						{Target: glucose.Target{Low: pointer.FromFloat64(100), High: pointer.FromFloat64(105)}},
					}
					expected := []error{
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/0/low"),
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/0/high"),
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/low"),
					}
					guardrails.ValidateBloodGlucoseTargetSchedule(schedule, glucoseSafetyLimit, guardRail, validator)
					errorsTest.ExpectEqual(validator.Error(), expected...)
				})
			})
		})
	})

	Context("GetBounds", func() {
		It("Generates bounds correctly", func() {
			var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
				{Target: glucose.Target{Low: pointer.FromFloat64(110), High: pointer.FromFloat64(120)}},
				{Target: glucose.Target{Low: pointer.FromFloat64(100), High: pointer.FromFloat64(105)}},
			}
			var premeal = glucose.Target{Low: pointer.FromFloat64(160), High: pointer.FromFloat64(250)}
			var workout = glucose.Target{Low: pointer.FromFloat64(170), High: pointer.FromFloat64(175)}
			correctionRanges := guardrails.CorrectionRanges{
				Schedule:         &schedule,
				Preprandial:      &premeal,
				PhysicalActivity: &workout,
			}

			bounds := correctionRanges.GetBounds()
			Expect(bounds).ToNot(BeNil())
			Expect(bounds.Lower).To(Equal(float64(100)))
			Expect(bounds.Upper).To(Equal(float64(250)))
		})
	})
})
