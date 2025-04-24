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

var _ = Describe("ValidateGlucoseSafetyLimit", func() {
	var guardRail *api.GlucoseSafetyLimitGuardRail
	var validator *structureValidator.Validator
	var correctionRanges guardrails.CorrectionRanges

	BeforeEach(func() {
		guardRail = test.NewGlucoseSafetyLimitGuardRail()
		validator = structureValidator.New(logTest.NewLogger())
		correctionRanges = guardrails.CorrectionRanges{}
	})

	It("doesn't return error with a valid value", func() {
		glucoseSafetyLimit := pointer.FromFloat64(70)
		guardrails.ValidateGlucoseSafetyLimit(glucoseSafetyLimit, correctionRanges, guardRail, validator)
		Expect(validator.Error()).To(BeNil())
	})

	It("doesn't return an error when set to the maximum of the device guardrail", func() {
		glucoseSafetyLimit := pointer.FromFloat64(110)
		guardrails.ValidateGlucoseSafetyLimit(glucoseSafetyLimit, correctionRanges, guardRail, validator)
		Expect(validator.Error()).To(BeNil())
	})

	It("returns an error with a value outside of the range", func() {
		glucoseSafetyLimit := pointer.FromFloat64(111)
		expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "")
		guardrails.ValidateGlucoseSafetyLimit(glucoseSafetyLimit, correctionRanges, guardRail, validator)
		errorsTest.ExpectEqual(validator.Error(), expected)
	})

	It("returns an error with a fractional value", func() {
		glucoseSafetyLimit := pointer.FromFloat64(70.5)
		expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "")
		guardrails.ValidateGlucoseSafetyLimit(glucoseSafetyLimit, correctionRanges, guardRail, validator)
		errorsTest.ExpectEqual(validator.Error(), expected)
	})

	It("doesn't return error when set lower than the lowest scheduled correction range", func() {
		glucoseSafetyLimit := pointer.FromFloat64(70)
		var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
			{Target: glucose.Target{Low: pointer.FromFloat64(110), High: pointer.FromFloat64(120)}},
			{Target: glucose.Target{Low: pointer.FromFloat64(100), High: pointer.FromFloat64(105)}},
		}
		var premeal = glucose.Target{Low: pointer.FromFloat64(160), High: pointer.FromFloat64(250)}
		var workout = glucose.Target{Low: pointer.FromFloat64(170), High: pointer.FromFloat64(175)}
		correctionRanges = guardrails.CorrectionRanges{
			Schedule:         &schedule,
			Preprandial:      &premeal,
			PhysicalActivity: &workout,
		}
		guardrails.ValidateGlucoseSafetyLimit(glucoseSafetyLimit, correctionRanges, guardRail, validator)
		Expect(validator.Error()).To(BeNil())
	})

	It("doesn't return error when set to the lower bound of all correction ranges when lowest is in the schedule", func() {
		glucoseSafetyLimit := pointer.FromFloat64(100)
		var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
			{Target: glucose.Target{Low: pointer.FromFloat64(110), High: pointer.FromFloat64(120)}},
			{Target: glucose.Target{Low: pointer.FromFloat64(100), High: pointer.FromFloat64(105)}},
		}
		var premeal = glucose.Target{Low: pointer.FromFloat64(160), High: pointer.FromFloat64(250)}
		var workout = glucose.Target{Low: pointer.FromFloat64(170), High: pointer.FromFloat64(175)}
		correctionRanges = guardrails.CorrectionRanges{
			Schedule:         &schedule,
			Preprandial:      &premeal,
			PhysicalActivity: &workout,
		}
		guardrails.ValidateGlucoseSafetyLimit(glucoseSafetyLimit, correctionRanges, guardRail, validator)
		Expect(validator.Error()).To(BeNil())
	})

	It("doesn't return error when set to the lower bound of all correction ranges when the lowest is premeal", func() {
		glucoseSafetyLimit := pointer.FromFloat64(95)
		var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
			{Target: glucose.Target{Low: pointer.FromFloat64(110), High: pointer.FromFloat64(120)}},
			{Target: glucose.Target{Low: pointer.FromFloat64(100), High: pointer.FromFloat64(105)}},
		}
		var premeal = glucose.Target{Low: pointer.FromFloat64(95), High: pointer.FromFloat64(250)}
		var workout = glucose.Target{Low: pointer.FromFloat64(170), High: pointer.FromFloat64(175)}
		correctionRanges = guardrails.CorrectionRanges{
			Schedule:         &schedule,
			Preprandial:      &premeal,
			PhysicalActivity: &workout,
		}
		guardrails.ValidateGlucoseSafetyLimit(glucoseSafetyLimit, correctionRanges, guardRail, validator)
		Expect(validator.Error()).To(BeNil())
	})

	It("doesn't return error when set to the lower bound of all correction ranges when the lowest is workout", func() {
		glucoseSafetyLimit := pointer.FromFloat64(95)
		var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
			{Target: glucose.Target{Low: pointer.FromFloat64(110), High: pointer.FromFloat64(120)}},
			{Target: glucose.Target{Low: pointer.FromFloat64(100), High: pointer.FromFloat64(105)}},
		}
		var premeal = glucose.Target{Low: pointer.FromFloat64(120), High: pointer.FromFloat64(250)}
		var workout = glucose.Target{Low: pointer.FromFloat64(95), High: pointer.FromFloat64(175)}
		correctionRanges = guardrails.CorrectionRanges{
			Schedule:         &schedule,
			Preprandial:      &premeal,
			PhysicalActivity: &workout,
		}
		guardrails.ValidateGlucoseSafetyLimit(glucoseSafetyLimit, correctionRanges, guardRail, validator)
		Expect(validator.Error()).To(BeNil())
	})

	It("returns an error when set higher the lowest scheduled correction range", func() {
		glucoseSafetyLimit := pointer.FromFloat64(101)
		var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
			{Target: glucose.Target{Low: pointer.FromFloat64(110), High: pointer.FromFloat64(120)}},
			{Target: glucose.Target{Low: pointer.FromFloat64(100), High: pointer.FromFloat64(105)}},
		}
		var premeal = glucose.Target{Low: pointer.FromFloat64(120), High: pointer.FromFloat64(250)}
		var workout = glucose.Target{Low: pointer.FromFloat64(130), High: pointer.FromFloat64(175)}
		correctionRanges = guardrails.CorrectionRanges{
			Schedule:         &schedule,
			Preprandial:      &premeal,
			PhysicalActivity: &workout,
		}
		expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "")
		guardrails.ValidateGlucoseSafetyLimit(glucoseSafetyLimit, correctionRanges, guardRail, validator)
		errorsTest.ExpectEqual(validator.Error(), expected)
	})

	It("returns an error when set higher the premeal correction range", func() {
		glucoseSafetyLimit := pointer.FromFloat64(101)
		var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
			{Target: glucose.Target{Low: pointer.FromFloat64(110), High: pointer.FromFloat64(120)}},
			{Target: glucose.Target{Low: pointer.FromFloat64(104), High: pointer.FromFloat64(105)}},
		}
		var premeal = glucose.Target{Low: pointer.FromFloat64(100), High: pointer.FromFloat64(250)}
		var workout = glucose.Target{Low: pointer.FromFloat64(130), High: pointer.FromFloat64(175)}
		correctionRanges = guardrails.CorrectionRanges{
			Schedule:         &schedule,
			Preprandial:      &premeal,
			PhysicalActivity: &workout,
		}
		expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "")
		guardrails.ValidateGlucoseSafetyLimit(glucoseSafetyLimit, correctionRanges, guardRail, validator)
		errorsTest.ExpectEqual(validator.Error(), expected)
	})

	It("returns an error when set higher the workout correction range", func() {
		glucoseSafetyLimit := pointer.FromFloat64(101)
		var schedule pump.BloodGlucoseTargetStartArray = []*pump.BloodGlucoseTargetStart{
			{Target: glucose.Target{Low: pointer.FromFloat64(110), High: pointer.FromFloat64(120)}},
			{Target: glucose.Target{Low: pointer.FromFloat64(104), High: pointer.FromFloat64(105)}},
		}
		var premeal = glucose.Target{Low: pointer.FromFloat64(120), High: pointer.FromFloat64(250)}
		var workout = glucose.Target{Low: pointer.FromFloat64(100), High: pointer.FromFloat64(175)}
		correctionRanges = guardrails.CorrectionRanges{
			Schedule:         &schedule,
			Preprandial:      &premeal,
			PhysicalActivity: &workout,
		}
		expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "")
		guardrails.ValidateGlucoseSafetyLimit(glucoseSafetyLimit, correctionRanges, guardRail, validator)
		errorsTest.ExpectEqual(validator.Error(), expected)
	})
})
