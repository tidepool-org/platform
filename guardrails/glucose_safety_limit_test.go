package guardrails_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tidepool-org/devices/api"
	"github.com/tidepool-org/platform/guardrails"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/guardrails/test"
	"github.com/tidepool-org/platform/pointer"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("ValidateGlucoseSafetyLimit", func() {
	var guardRail *api.GlucoseSafetyLimitGuardRail
	var validator *structureValidator.Validator
	var correctionRanges guardrails.CorrectionRanges

	BeforeEach(func() {
		guardRail = test.NewGlucoseSafetyLimitGuardRail()
		validator = structureValidator.New()
		correctionRanges = guardrails.CorrectionRanges{}
	})

	It("doesn't return error with a valid value", func() {
		glucoseSafetyLimit := pointer.FromFloat64(70)
		guardrails.ValidateGlucoseSafetyLimit(glucoseSafetyLimit, correctionRanges, guardRail, validator)
		Expect(validator.Error()).To(BeNil())
	})

	It("returns an error with a value outside of the range", func() {
		glucoseSafetyLimit := pointer.FromFloat64(190)
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
})
