package guardrails_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidepool-org/devices/api"

	"github.com/tidepool-org/platform/guardrails"

	"github.com/tidepool-org/platform/data/types/settings/pump"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/guardrails/test"
	"github.com/tidepool-org/platform/pointer"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("ValidateBolusAmountMaximum", func() {
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
		guardrails.ValidateBolusAmountMaximum(value, guardRail, validator)
		Expect(validator.Error()).To(BeNil())
	})

	It("doesn't return error when set to the pump max supported value", func() {
		var value = pump.BolusAmountMaximum{
			Value: pointer.FromFloat64(30.00),
		}
		guardrails.ValidateBolusAmountMaximum(value, guardRail, validator)
		Expect(validator.Error()).To(BeNil())
	})

	It("doesn't return error when set to the pump min supported value", func() {
		var value = pump.BolusAmountMaximum{
			Value: pointer.FromFloat64(0.05),
		}
		guardrails.ValidateBolusAmountMaximum(value, guardRail, validator)
		Expect(validator.Error()).To(BeNil())
	})

	It("returns an error with a value higher than the pump max supported value", func() {
		var value = pump.BolusAmountMaximum{
			Value: pointer.FromFloat64(30.05),
		}
		expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/value")
		guardrails.ValidateBolusAmountMaximum(value, guardRail, validator)
		errorsTest.ExpectEqual(validator.Error(), expected)
	})

	It("returns an error with a value lower than the pump min supported value", func() {
		var value = pump.BolusAmountMaximum{
			Value: pointer.FromFloat64(0.0),
		}
		expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/value")
		guardrails.ValidateBolusAmountMaximum(value, guardRail, validator)
		errorsTest.ExpectEqual(validator.Error(), expected)
	})

	It("returns an error with an invalid increment", func() {
		var value = pump.BolusAmountMaximum{
			Value: pointer.FromFloat64(0.06),
		}
		expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/value")
		guardrails.ValidateBolusAmountMaximum(value, guardRail, validator)
		errorsTest.ExpectEqual(validator.Error(), expected)
	})
})
