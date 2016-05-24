package calculator_test

import (
	"github.com/tidepool-org/platform/pvn/data/context"
	"github.com/tidepool-org/platform/pvn/data/normalizer"
	"github.com/tidepool-org/platform/pvn/data/types/base/bolus/calculator"
	"github.com/tidepool-org/platform/pvn/data/types/base/testing"
	"github.com/tidepool-org/platform/pvn/data/types/common/bloodglucose"
	"github.com/tidepool-org/platform/pvn/data/validator"
	"github.com/tidepool-org/platform/service"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Calculator Bolus", func() {

	var rawObject = testing.RawBaseObject()

	var embeddedBolus = func(subType string, normal, extended float64, duration int) map[string]interface{} {
		var rawBolus = testing.RawBaseObject()

		rawBolus["subType"] = subType
		rawBolus["type"] = "bolus"

		if normal > 0.0 {
			rawBolus["normal"] = normal
		}
		if extended > 0.0 {
			rawBolus["extended"] = extended
		}
		if duration > 0 {
			rawBolus["duration"] = duration
		}
		return rawBolus
	}

	BeforeEach(func() {

		rawObject["type"] = "wizard"
		rawObject["units"] = "mg/dl"
		rawObject["bgInput"] = 100
		rawObject["carbInput"] = 120
		rawObject["insulinSensitivity"] = 90
		rawObject["insulinCarbRatio"] = 50
		rawObject["insulinOnBoard"] = 70

		rawObject["recommended"] = map[string]interface{}{"net": 50, "correction": -50, "carb": 50}
		rawObject["bgTarget"] = map[string]interface{}{"target": 100, "range": 10}

		rawObject["bolus"] = embeddedBolus("normal", 52.1, 0.0, 0)

	})

	Context("insulinOnBoard", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("negative", rawObject, "insulinOnBoard", -1,
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorIntegerNotInRange(-1, 0, 250), "/insulinOnBoard")},
			),
			Entry("greater than 250", rawObject, "insulinOnBoard", 251,
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorIntegerNotInRange(251, 0, 250), "/insulinOnBoard")},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("within bounds", rawObject, "insulinOnBoard", 99),
		)

	})

	Context("insulinCarbRatio", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("negative", rawObject, "insulinCarbRatio", -1,
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorIntegerNotInRange(-1, 0, 250), "/insulinCarbRatio")},
			),
			Entry("greater than 250", rawObject, "insulinCarbRatio", 251,
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorIntegerNotInRange(251, 0, 250), "/insulinCarbRatio")},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("within bounds", rawObject, "insulinCarbRatio", 99),
		)

	})

	Context("units", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("empty", rawObject, "units", "",
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorStringNotOneOf("", []string{bloodglucose.Mmoll, bloodglucose.MmolL, bloodglucose.Mgdl, bloodglucose.MgdL}), "/units")},
			),
			Entry("not one of the predefined values", rawObject, "units", "wrong",
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorStringNotOneOf("wrong", []string{bloodglucose.Mmoll, bloodglucose.MmolL, bloodglucose.Mgdl, bloodglucose.MgdL}), "/units")},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("mmol/l", rawObject, "units", "mmol/l"),
			Entry("mmol/L", rawObject, "units", "mmol/L"),
			Entry("mg/dl", rawObject, "units", "mg/dl"),
			Entry("mg/dL", rawObject, "units", "mg/dL"),
		)

	})

	Context("bgInput", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("less than 0", rawObject, "bgInput", -0.1,
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorFloatNotInRange(-0.1, bloodglucose.MgdLFromValue, bloodglucose.MgdLToValue), "/bgInput")},
			),
			Entry("greater than 1000", rawObject, "bgInput", 1000.1,
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorFloatNotInRange(1000.1, bloodglucose.MgdLFromValue, bloodglucose.MgdLToValue), "/bgInput")},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("0", rawObject, "bgInput", 0.0),
			Entry("above 0", rawObject, "bgInput", 0.1),
			Entry("below max", rawObject, "bgInput", 990.85745),
			Entry("as integer", rawObject, "bgInput", 4),
		)

	})

	Context("insulinSensitivity", func() {

		BeforeEach(func() {
			rawObject["units"] = bloodglucose.MgdL
		})

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("less than 0", rawObject, "insulinSensitivity", -0.1,
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorFloatNotInRange(-0.1, bloodglucose.MgdLFromValue, bloodglucose.MgdLToValue), "/insulinSensitivity")},
			),
			Entry("greater than 1000", rawObject, "insulinSensitivity", 1000.1,
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorFloatNotInRange(1000.1, bloodglucose.MgdLFromValue, bloodglucose.MgdLToValue), "/insulinSensitivity")},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("0", rawObject, "insulinSensitivity", 0.0),
			Entry("above 0", rawObject, "insulinSensitivity", 0.1),
			Entry("below max mg/dl", rawObject, "insulinSensitivity", 990.85745),
			Entry("as integer", rawObject, "insulinSensitivity", 4),
		)

	})

	Context("carbInput", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("less than 0", rawObject, "carbInput", -1,
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorIntegerNotInRange(-1, 0, 1000), "/carbInput")},
			),
			Entry("greater than 1000", rawObject, "carbInput", 1001,
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorIntegerNotInRange(1001, 0, 1000), "/carbInput")},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("0", rawObject, "carbInput", 0),
			Entry("in range", rawObject, "carbInput", 250),
			Entry("below 1000", rawObject, "carbInput", 999),
		)

	})

	Context("bgTarget", func() {

		BeforeEach(func() {
			rawObject["units"] = bloodglucose.MgdL
		})

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("range less than 0", rawObject, "bgTarget", map[string]interface{}{"target": 100, "range": -1},
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorIntegerNotInRange(-1, 0, 50), "/bgTarget/range")},
			),
			Entry("range greater than 50", rawObject, "bgTarget", map[string]interface{}{"target": 100, "range": 51},
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorIntegerNotInRange(51, 0, 50), "/bgTarget/range")},
			),
			Entry("target less than 0", rawObject, "bgTarget", map[string]interface{}{"target": -0.1, "range": 10},
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorFloatNotInRange(-0.1, bloodglucose.MgdLFromValue, bloodglucose.MgdLToValue), "/bgTarget/target")},
			),
			Entry("target greater than 1000", rawObject, "bgTarget", map[string]interface{}{"target": 1000.1, "range": 10},
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorFloatNotInRange(1000.1, bloodglucose.MgdLFromValue, bloodglucose.MgdLToValue), "/bgTarget/target")},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("range 0", rawObject, "bgTarget", map[string]interface{}{"target": 100, "range": 0}),
			Entry("target 0", rawObject, "bgTarget", map[string]interface{}{"target": 0.0, "range": 10}),
			Entry("range less or equal to 50", rawObject, "bgTarget", map[string]interface{}{"target": 100, "range": 50}),
			Entry("target less or equal to max mgdl", rawObject, "bgTarget", map[string]interface{}{"target": 990.85745, "range": 10}),
		)

	})

	Context("recommended", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("net less than -100", rawObject, "recommended", map[string]interface{}{"net": -101, "correction": -50, "carb": 50},
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorIntegerNotInRange(-101, -100, 100), "/recommended/net")},
			),
			Entry("net greater than 100", rawObject, "recommended", map[string]interface{}{"net": 101, "correction": -50, "carb": 50},
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorIntegerNotInRange(101, -100, 100), "/recommended/net")},
			),
			Entry("correction less than -100", rawObject, "recommended", map[string]interface{}{"net": 50, "correction": -101, "carb": 50},
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorIntegerNotInRange(-101, -100, 100), "/recommended/correction")},
			),
			Entry("correction greater than 100", rawObject, "recommended", map[string]interface{}{"net": 50, "correction": 101, "carb": 50},
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorIntegerNotInRange(101, -100, 100), "/recommended/correction")},
			),
			Entry("carb less than 0", rawObject, "recommended", map[string]interface{}{"net": 50, "correction": -50, "carb": -1},
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorIntegerNotInRange(-1, 0, 100), "/recommended/carb")},
			),
			Entry("carb greater than 100", rawObject, "recommended", map[string]interface{}{"net": 50, "correction": -50, "carb": 101},
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorIntegerNotInRange(101, 0, 100), "/recommended/carb")},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("net more or equal -100", rawObject, "recommended", map[string]interface{}{"net": -100, "correction": -50, "carb": 50}),
			Entry("net less or equal 100", rawObject, "recommended", map[string]interface{}{"net": 100, "correction": -50, "carb": 50}),
			Entry("correction more or equal -100", rawObject, "recommended", map[string]interface{}{"net": 10, "correction": -100, "carb": 50}),
			Entry("correction less or equal 100", rawObject, "recommended", map[string]interface{}{"net": 10, "correction": 100, "carb": 50}),
			Entry("carb more or equal 0", rawObject, "recommended", map[string]interface{}{"net": -100, "correction": -50, "carb": 0}),
			Entry("carb less or equal 50", rawObject, "recommended", map[string]interface{}{"net": 100, "correction": -50, "carb": 50}),
		)

	})

	Context("Normalize", func() {

		Context("blood glucose", func() {

			DescribeTable("when mmol/L", func(val, expected float64) {
				bolusCalculator, err := calculator.New()
				Expect(err).To(BeNil())
				bolusCalculator.Units = &bloodglucose.Mmoll
				bolusCalculator.BloodGlucoseInput = &val
				bolusCalculator.InsulinSensitivity = &val
				bolusCalculator.BloodGlucoseTarget = &calculator.BloodGlucoseTarget{Target: &val}

				testContext := context.NewStandard()
				standardNormalizer, err := normalizer.NewStandard(testContext)
				Expect(err).To(BeNil())
				bolusCalculator.Normalize(standardNormalizer)
				Expect(bolusCalculator.Units).To(Equal(&bloodglucose.MmolL))
				Expect(bolusCalculator.BloodGlucoseInput).To(Equal(&expected))
				Expect(bolusCalculator.InsulinSensitivity).To(Equal(&expected))
				Expect(bolusCalculator.BloodGlucoseTarget.Target).To(Equal(&expected))
			},
				Entry("expected lower bg value", 3.7, 3.7),
				Entry("below max", 54.99, 54.99),
				Entry("expected upper bg value", 23.0, 23.0),
			)

			DescribeTable("when mg/dL", func(val, expected float64) {
				bolusCalculator, err := calculator.New()
				Expect(err).To(BeNil())
				bolusCalculator.Units = &bloodglucose.Mgdl
				bolusCalculator.BloodGlucoseInput = &val
				bolusCalculator.InsulinSensitivity = &val
				bolusCalculator.BloodGlucoseTarget = &calculator.BloodGlucoseTarget{Target: &val}

				testContext := context.NewStandard()
				standardNormalizer, err := normalizer.NewStandard(testContext)
				Expect(err).To(BeNil())
				bolusCalculator.Normalize(standardNormalizer)
				Expect(bolusCalculator.Units).To(Equal(&bloodglucose.MmolL))
				Expect(bolusCalculator.BloodGlucoseInput).To(Equal(&expected))
				Expect(bolusCalculator.InsulinSensitivity).To(Equal(&expected))
				Expect(bolusCalculator.BloodGlucoseTarget.Target).To(Equal(&expected))
			},
				Entry("expected lower bg value", 60.0, 3.33044879462732),
				Entry("below max", 990.85745, 55.0),
				Entry("expected upper bg value", 400.0, 22.202991964182132),
			)
		})

		Context("bolus", func() {
			DescribeTable("valid when embedded", func(rawObject map[string]interface{}, field string, val interface{}) {
				calculatorDatum := testing.ParseAndNormalize(rawObject, field, val)
				calculatorBolus := calculatorDatum.(*calculator.Calculator)
				Expect(calculatorBolus.BolusID).To(Not(BeNil()))
			},
				Entry("normal", rawObject, "bolus", embeddedBolus("normal", 52.1, 0.0, 0)),
				Entry("square", rawObject, "bolus", embeddedBolus("square", 0.0, 52.1, 1000)),
				Entry("dual/square", rawObject, "bolus", embeddedBolus("dual/square", 52.1, 52.1, 1000)),
			)

			DescribeTable("invalid when embedded", func(rawObject map[string]interface{}, field string, val interface{}) {
				calculatorDatum := testing.ParseAndNormalize(rawObject, field, val)
				calculatorBolus := calculatorDatum.(*calculator.Calculator)
				Expect(calculatorBolus.BolusID).To(BeNil())
			},
				Entry("wrong subType", rawObject, "bolus", embeddedBolus("wrong", 0.0, 52.1, 0)),
			)
		})
	})

})
