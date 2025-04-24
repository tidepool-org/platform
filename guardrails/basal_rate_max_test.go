package guardrails_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/devices/api"

	"github.com/tidepool-org/platform/data/types/settings/pump"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/guardrails"
	"github.com/tidepool-org/platform/guardrails/test"
	logTest "github.com/tidepool-org/platform/log/test"
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
		validator = structureValidator.New(logTest.NewLogger())
		basalRateSchedule = make(pump.BasalRateStartArray, 0)
		carbRatioSchedule = make(pump.CarbohydrateRatioStartArray, 0)
	})

	It("doesn't return error with a valid value", func() {
		value := pump.BasalRateMaximum{
			Value: pointer.FromFloat64(12.35),
		}
		guardrails.ValidateBasalRateMaximum(value, &basalRateSchedule, &carbRatioSchedule, guardRail, validator)
		Expect(validator.Error()).To(BeNil())
	})

	It("returns an error with an invalid value", func() {
		var value = pump.BasalRateMaximum{
			Value: pointer.FromFloat64(31.00),
		}
		expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/value")
		guardrails.ValidateBasalRateMaximum(value, &basalRateSchedule, &carbRatioSchedule, guardRail, validator)
		errorsTest.ExpectEqual(validator.Error(), expected)
	})

	It("doesn't return an error when set equal to the highest scheduled basal rate", func() {
		basalRateSchedule = pump.BasalRateStartArray{
			&pump.BasalRateStart{Rate: pointer.FromFloat64(12.55)},
			&pump.BasalRateStart{Rate: pointer.FromFloat64(10.05)},
			&pump.BasalRateStart{Rate: pointer.FromFloat64(13.65)},
		}
		value := pump.BasalRateMaximum{
			Value: pointer.FromFloat64(13.65),
		}
		guardrails.ValidateBasalRateMaximum(value, &basalRateSchedule, &carbRatioSchedule, guardRail, validator)
		Expect(validator.Error()).To(BeNil())
	})

	It("doesn't return an error when set higher than the highest scheduled basal rate", func() {
		basalRateSchedule = pump.BasalRateStartArray{
			&pump.BasalRateStart{Rate: pointer.FromFloat64(12.55)},
			&pump.BasalRateStart{Rate: pointer.FromFloat64(10.05)},
			&pump.BasalRateStart{Rate: pointer.FromFloat64(13.65)},
		}
		value := pump.BasalRateMaximum{
			Value: pointer.FromFloat64(14.70),
		}
		guardrails.ValidateBasalRateMaximum(value, &basalRateSchedule, &carbRatioSchedule, guardRail, validator)
		Expect(validator.Error()).To(BeNil())
	})

	It("returns an error when set to lower than the highest scheduled basal rate", func() {
		basalRateSchedule = pump.BasalRateStartArray{
			&pump.BasalRateStart{Rate: pointer.FromFloat64(12.55)},
			&pump.BasalRateStart{Rate: pointer.FromFloat64(10.55)},
			&pump.BasalRateStart{Rate: pointer.FromFloat64(13.55)},
		}
		value := pump.BasalRateMaximum{
			Value: pointer.FromFloat64(13.50),
		}
		expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/value")
		guardrails.ValidateBasalRateMaximum(value, &basalRateSchedule, &carbRatioSchedule, guardRail, validator)
		errorsTest.ExpectEqual(validator.Error(), expected)
	})

	It("doesn't return an error when set lower than the highest scheduled basal rate", func() {
		basalRateSchedule = pump.BasalRateStartArray{
			&pump.BasalRateStart{Rate: pointer.FromFloat64(12.55)},
			&pump.BasalRateStart{Rate: pointer.FromFloat64(10.05)},
			&pump.BasalRateStart{Rate: pointer.FromFloat64(13.65)},
		}
		value := pump.BasalRateMaximum{
			Value: pointer.FromFloat64(13.65),
		}
		guardrails.ValidateBasalRateMaximum(value, &basalRateSchedule, &carbRatioSchedule, guardRail, validator)
		Expect(validator.Error()).To(BeNil())
	})

	It("doesn't return an error when set equal to min((70 / (lowest scheduled carb ratio), pump max basal rate))", func() {
		carbRatioSchedule = pump.CarbohydrateRatioStartArray{
			&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(2)},
			&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(50)},
			&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(150)},
		}
		value := pump.BasalRateMaximum{
			Value: pointer.FromFloat64(30.00),
		}
		guardrails.ValidateBasalRateMaximum(value, &basalRateSchedule, &carbRatioSchedule, guardRail, validator)
		Expect(validator.Error()).To(BeNil())
	})

	It("doesn't return an error when set equal to min((70 / (lowest scheduled carb ratio), pump max basal rate))", func() {
		carbRatioSchedule = pump.CarbohydrateRatioStartArray{
			&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(50)},
			&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(7)},
			&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(150)},
		}
		value := pump.BasalRateMaximum{
			Value: pointer.FromFloat64(10.0),
		}
		guardrails.ValidateBasalRateMaximum(value, &basalRateSchedule, &carbRatioSchedule, guardRail, validator)
		Expect(validator.Error()).To(BeNil())
	})

	It("doesn't return an error when set lower than min((70 / (lowest scheduled carb ratio), pump max basal rate))", func() {
		carbRatioSchedule = pump.CarbohydrateRatioStartArray{
			&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(50)},
			&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(3)},
			&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(150)},
		}
		value := pump.BasalRateMaximum{
			Value: pointer.FromFloat64(23.30),
		}
		guardrails.ValidateBasalRateMaximum(value, &basalRateSchedule, &carbRatioSchedule, guardRail, validator)
		Expect(validator.Error()).To(BeNil())
	})

	It("returns an error when set higher than min((70 / (lowest scheduled carb ratio), pump max basal rate))", func() {
		carbRatioSchedule = pump.CarbohydrateRatioStartArray{
			&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(149)},
			&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(150)},
		}
		value := pump.BasalRateMaximum{
			Value: pointer.FromFloat64(0.5),
		}
		expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/value")
		guardrails.ValidateBasalRateMaximum(value, &basalRateSchedule, &carbRatioSchedule, guardRail, validator)
		errorsTest.ExpectEqual(validator.Error(), expected)
	})

	It("returns an error when set higher than pump max basal rate)", func() {
		carbRatioSchedule = pump.CarbohydrateRatioStartArray{
			&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(50)},
			&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(2)},
			&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(150)},
		}
		basalRateSchedule = pump.BasalRateStartArray{
			&pump.BasalRateStart{Rate: pointer.FromFloat64(0.55)},
			&pump.BasalRateStart{Rate: pointer.FromFloat64(20.00)},
			&pump.BasalRateStart{Rate: pointer.FromFloat64(0.05)},
		}
		value := pump.BasalRateMaximum{
			Value: pointer.FromFloat64(30.05),
		}
		expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/value")
		guardrails.ValidateBasalRateMaximum(value, &basalRateSchedule, &carbRatioSchedule, guardRail, validator)
		errorsTest.ExpectEqual(validator.Error(), expected)
	})

	When("(70 / (lowest scheduled carb ratio) > (hSBR*6.4) and value is below pump max basal rate", func() {
		It("returns an error when set higher than max((70 / (lowest scheduled carb ratio), (hSBR*6.4))", func() {
			carbRatioSchedule = pump.CarbohydrateRatioStartArray{
				&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(50)},
				&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(7)},
				&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(150)},
			}
			basalRateSchedule = pump.BasalRateStartArray{
				&pump.BasalRateStart{Rate: pointer.FromFloat64(0.55)},
				&pump.BasalRateStart{Rate: pointer.FromFloat64(1.05)},
				&pump.BasalRateStart{Rate: pointer.FromFloat64(0.05)},
			}

			value := pump.BasalRateMaximum{
				Value: pointer.FromFloat64(10.05),
			}
			expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/value")
			guardrails.ValidateBasalRateMaximum(value, &basalRateSchedule, &carbRatioSchedule, guardRail, validator)
			errorsTest.ExpectEqual(validator.Error(), expected)
		})

		It("doesn't return an error when set equal to max((70 / (lowest scheduled carb ratio), (hSBR*6.4))", func() {
			carbRatioSchedule = pump.CarbohydrateRatioStartArray{
				&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(50)},
				&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(7)},
				&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(150)},
			}
			basalRateSchedule = pump.BasalRateStartArray{
				&pump.BasalRateStart{Rate: pointer.FromFloat64(0.55)},
				&pump.BasalRateStart{Rate: pointer.FromFloat64(1.05)},
				&pump.BasalRateStart{Rate: pointer.FromFloat64(0.05)},
			}

			value := pump.BasalRateMaximum{
				Value: pointer.FromFloat64(10.0),
			}
			guardrails.ValidateBasalRateMaximum(value, &basalRateSchedule, &carbRatioSchedule, guardRail, validator)
			Expect(validator.Error()).To(BeNil())
		})

		It("doesn't return an error when set lower than max((70 / (lowest scheduled carb ratio), (hSBR*6.4))", func() {
			carbRatioSchedule = pump.CarbohydrateRatioStartArray{
				&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(50)},
				&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(7)},
				&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(150)},
			}
			basalRateSchedule = pump.BasalRateStartArray{
				&pump.BasalRateStart{Rate: pointer.FromFloat64(0.55)},
				&pump.BasalRateStart{Rate: pointer.FromFloat64(1.05)},
				&pump.BasalRateStart{Rate: pointer.FromFloat64(0.05)},
			}

			value := pump.BasalRateMaximum{
				Value: pointer.FromFloat64(9.95),
			}
			guardrails.ValidateBasalRateMaximum(value, &basalRateSchedule, &carbRatioSchedule, guardRail, validator)
			Expect(validator.Error()).To(BeNil())
		})
	})

	When("(70 / (lowest scheduled carb ratio) < (hSBR*6.4) and value is below pump max basal rate", func() {
		It("returns an error when set higher than max((70 / (lowest scheduled carb ratio), (hSBR*6.4))", func() {
			carbRatioSchedule = pump.CarbohydrateRatioStartArray{
				&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(50)},
				&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(14)},
				&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(150)},
			}
			basalRateSchedule = pump.BasalRateStartArray{
				&pump.BasalRateStart{Rate: pointer.FromFloat64(0.55)},
				&pump.BasalRateStart{Rate: pointer.FromFloat64(1.50)},
				&pump.BasalRateStart{Rate: pointer.FromFloat64(0.05)},
			}

			value := pump.BasalRateMaximum{
				Value: pointer.FromFloat64(9.65),
			}
			expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/value")
			guardrails.ValidateBasalRateMaximum(value, &basalRateSchedule, &carbRatioSchedule, guardRail, validator)
			errorsTest.ExpectEqual(validator.Error(), expected)
		})

		It("doesn't return an error when set equal to max((70 / (lowest scheduled carb ratio), (hSBR*6.4))", func() {
			carbRatioSchedule = pump.CarbohydrateRatioStartArray{
				&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(50)},
				&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(14)},
				&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(150)},
			}
			basalRateSchedule = pump.BasalRateStartArray{
				&pump.BasalRateStart{Rate: pointer.FromFloat64(0.55)},
				&pump.BasalRateStart{Rate: pointer.FromFloat64(1.50)},
				&pump.BasalRateStart{Rate: pointer.FromFloat64(0.05)},
			}

			value := pump.BasalRateMaximum{
				Value: pointer.FromFloat64(9.60),
			}
			guardrails.ValidateBasalRateMaximum(value, &basalRateSchedule, &carbRatioSchedule, guardRail, validator)
			Expect(validator.Error()).To(BeNil())
		})

		It("doesn't return an error when set lower than max((70 / (lowest scheduled carb ratio), (hSBR*6.4))", func() {
			carbRatioSchedule = pump.CarbohydrateRatioStartArray{
				&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(50)},
				&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(14)},
				&pump.CarbohydrateRatioStart{Amount: pointer.FromFloat64(150)},
			}
			basalRateSchedule = pump.BasalRateStartArray{
				&pump.BasalRateStart{Rate: pointer.FromFloat64(0.55)},
				&pump.BasalRateStart{Rate: pointer.FromFloat64(1.50)},
				&pump.BasalRateStart{Rate: pointer.FromFloat64(0.05)},
			}

			value := pump.BasalRateMaximum{
				Value: pointer.FromFloat64(9.55),
			}
			guardrails.ValidateBasalRateMaximum(value, &basalRateSchedule, &carbRatioSchedule, guardRail, validator)
			Expect(validator.Error()).To(BeNil())
		})
	})

})
