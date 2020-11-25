package guardrails_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tidepool-org/devices/api"
	"github.com/tidepool-org/platform/guardrails"

	"github.com/tidepool-org/platform/data/types/settings/pump"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/guardrails/test"
	"github.com/tidepool-org/platform/pointer"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("ValidateBasalRateMaximum", func() {
	var guardRail *api.BasalRateMaximumGuardRail
	var validator *structureValidator.Validator
	var basalRateSchedule pump.BasalRateStartArray
	var carbRatioSchedule pump.CarbohydrateRatioStartArray

	BeforeEach(func() {
		guardRail = test.NewBasalRateMaximumGuardRail()
		validator = structureValidator.New()
		basalRateSchedule = make(pump.BasalRateStartArray, 0)
		carbRatioSchedule = make(pump.CarbohydrateRatioStartArray, 0)
	})

	It("doesn't return error with a valid value", func() {
		var value = pump.BasalRateMaximum{
			Value: pointer.FromFloat64(1.00),
		}
		guardrails.ValidateBasalRateMaximum(value, basalRateSchedule, carbRatioSchedule, guardRail, validator)
		Expect(validator.Error()).To(BeNil())
	})

	It("returns an error with an invalid value", func() {
		var value = pump.BasalRateMaximum{
			Value: pointer.FromFloat64(31.00),
		}
		expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/value")
		guardrails.ValidateBasalRateMaximum(value, basalRateSchedule, carbRatioSchedule, guardRail, validator)
		errorsTest.ExpectEqual(validator.Error(), expected)
	})
})
