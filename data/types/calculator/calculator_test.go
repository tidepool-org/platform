package calculator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/calculator"
	"github.com/tidepool-org/platform/data/types/testing"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
)

func NewRawObjectWithMmolL() map[string]interface{} {
	rawObject := testing.RawBaseObject()
	rawObject["type"] = "wizard"
	rawObject["bgInput"] = 12.0
	rawObject["carbInput"] = 120.0
	rawObject["insulinSensitivity"] = 7.0
	rawObject["insulinCarbRatio"] = 50.0
	rawObject["insulinOnBoard"] = 70

	rawObject["recommended"] = map[string]interface{}{"net": 50, "correction": -50, "carb": 50}
	rawObject["bgTarget"] = map[string]interface{}{"target": 8.0, "range": 1.0}

	rawObject["bolus"] = NewEmbeddedBolus("bolus", "normal", 52.1, 0.0, 0)

	rawObject["units"] = glucose.MmolL
	return rawObject
}

func NewRawObjectWithMmoll() map[string]interface{} {
	rawObject := NewRawObjectWithMmolL()
	rawObject["units"] = glucose.Mmoll
	return rawObject
}

func NewRawObjectWithMgdL() map[string]interface{} {
	rawObject := testing.RawBaseObject()
	rawObject["type"] = "wizard"
	rawObject["bgInput"] = 100
	rawObject["carbInput"] = 120.0
	rawObject["insulinSensitivity"] = 90
	rawObject["insulinCarbRatio"] = 50.0
	rawObject["insulinOnBoard"] = 70

	rawObject["recommended"] = map[string]interface{}{"net": 50, "correction": -50, "carb": 50}
	rawObject["bgTarget"] = map[string]interface{}{"target": 100, "range": 10.0}

	rawObject["bolus"] = NewEmbeddedBolus("bolus", "normal", 52.1, 0.0, 0)

	rawObject["units"] = glucose.MgdL
	return rawObject
}

func NewRawObjectWithMgdl() map[string]interface{} {
	rawObject := NewRawObjectWithMgdL()
	rawObject["units"] = glucose.Mgdl
	return rawObject
}

func NewMeta() interface{} {
	return &types.Meta{
		Type: "wizard",
	}
}

func NewEmbeddedBolus(datumType interface{}, subType interface{}, normal, extended float64, duration int) map[string]interface{} {
	var rawBolus = testing.RawBaseObject()

	if datumType != nil {
		rawBolus["type"] = datumType
	}
	if subType != nil {
		rawBolus["subType"] = subType
	}

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

var _ = Describe("Calculator", func() {
	Context("insulinOnBoard", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is negative", NewRawObjectWithMgdl(), "insulinOnBoard", -1,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(-1, 0, 250), "/insulinOnBoard", NewMeta())},
			),
			Entry("is greater than 250", NewRawObjectWithMgdl(), "insulinOnBoard", 251,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(251, 0, 250), "/insulinOnBoard", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is within bounds", NewRawObjectWithMgdl(), "insulinOnBoard", 99),
		)
	})

	Context("insulinCarbRatio", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is negative", NewRawObjectWithMgdl(), "insulinCarbRatio", -1.0,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(-1.0, 0.0, 250.0), "/insulinCarbRatio", NewMeta())},
			),
			Entry("is greater than 250.0", NewRawObjectWithMgdl(), "insulinCarbRatio", 251.0,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(251.0, 0.0, 250.0), "/insulinCarbRatio", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is within bounds", NewRawObjectWithMgdl(), "insulinCarbRatio", 99.0),
		)
	})

	Context("units", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is empty", NewRawObjectWithMgdl(), "units", "",
				[]*service.Error{testing.ComposeError(service.ErrorValueStringNotOneOf("", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta())},
			),
			Entry("is not one of the predefined values", NewRawObjectWithMgdl(), "units", "wrong",
				[]*service.Error{testing.ComposeError(service.ErrorValueStringNotOneOf("wrong", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is mmol/L", NewRawObjectWithMmolL(), "units", "mmol/L"),
			Entry("is mmol/l", NewRawObjectWithMmoll(), "units", "mmol/l"),
			Entry("is mg/dL", NewRawObjectWithMgdL(), "units", "mg/dL"),
			Entry("is mg/dl", NewRawObjectWithMgdl(), "units", "mg/dl"),
		)
	})

	Context("bgInput", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is less than 0", NewRawObjectWithMgdl(), "bgInput", -0.1,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(-0.1, glucose.MgdLLowerLimit, glucose.MgdLUpperLimit), "/bgInput", NewMeta())},
			),
			Entry("is greater than 1000", NewRawObjectWithMgdl(), "bgInput", 1000.1,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(1000.1, glucose.MgdLLowerLimit, glucose.MgdLUpperLimit), "/bgInput", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is 0", NewRawObjectWithMgdl(), "bgInput", 0.0),
			Entry("is above 0", NewRawObjectWithMgdl(), "bgInput", 0.1),
			Entry("is below max", NewRawObjectWithMgdl(), "bgInput", glucose.MgdLUpperLimit),
			Entry("is an integer", NewRawObjectWithMgdl(), "bgInput", 4),
		)
	})

	Context("insulinSensitivity", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is less than 0", NewRawObjectWithMgdL(), "insulinSensitivity", -0.1,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(-0.1, glucose.MgdLLowerLimit, glucose.MgdLUpperLimit), "/insulinSensitivity", NewMeta())},
			),
			Entry("is greater than 1000", NewRawObjectWithMgdL(), "insulinSensitivity", 1000.1,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(1000.1, glucose.MgdLLowerLimit, glucose.MgdLUpperLimit), "/insulinSensitivity", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is 0", NewRawObjectWithMgdL(), "insulinSensitivity", 0.0),
			Entry("is above 0", NewRawObjectWithMgdL(), "insulinSensitivity", 0.1),
			Entry("is below max mg/dl", NewRawObjectWithMgdL(), "insulinSensitivity", glucose.MgdLUpperLimit),
			Entry("is an integer", NewRawObjectWithMgdL(), "insulinSensitivity", 4),
		)
	})

	Context("carbInput", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is less than 0", NewRawObjectWithMgdl(), "carbInput", -1.0,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(-1.0, 0.0, 1000.0), "/carbInput", NewMeta())},
			),
			Entry("is greater than 1000", NewRawObjectWithMgdl(), "carbInput", 1001.0,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(1001.0, 0.0, 1000.0), "/carbInput", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is 0", NewRawObjectWithMgdl(), "carbInput", 0.0),
			Entry("is in range", NewRawObjectWithMgdl(), "carbInput", 250.0),
			Entry("is below 1000", NewRawObjectWithMgdl(), "carbInput", 999.0),
		)
	})

	Context("bgTarget", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("has range less than 0", NewRawObjectWithMgdL(), "bgTarget", map[string]interface{}{"target": 100, "range": -1},
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(-1, 0, 100), "/bgTarget/range", NewMeta())},
			),
			Entry("has range greater than target", NewRawObjectWithMgdL(), "bgTarget", map[string]interface{}{"target": 100, "range": 101},
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(101, 0, 100), "/bgTarget/range", NewMeta())},
			),
			Entry("has target less than 0", NewRawObjectWithMgdL(), "bgTarget", map[string]interface{}{"target": -0.1, "range": 10},
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(-0.1, glucose.MgdLLowerLimit, glucose.MgdLUpperLimit), "/bgTarget/target", NewMeta())},
			),
			Entry("has target greater than 1000", NewRawObjectWithMgdL(), "bgTarget", map[string]interface{}{"target": 1000.1, "range": 10},
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(1000.1, glucose.MgdLLowerLimit, glucose.MgdLUpperLimit), "/bgTarget/target", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("has range 0", NewRawObjectWithMgdL(), "bgTarget", map[string]interface{}{"target": 100, "range": 0}),
			Entry("has target 0", NewRawObjectWithMgdL(), "bgTarget", map[string]interface{}{"target": 0.0, "range": 0}),
			Entry("has range less or equal to target", NewRawObjectWithMgdL(), "bgTarget", map[string]interface{}{"target": 100, "range": 100}),
			Entry("has target less or equal to max mgdl", NewRawObjectWithMgdL(), "bgTarget", map[string]interface{}{"target": glucose.MgdLUpperLimit, "range": 0}),
		)
	})

	Context("recommended", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("has net less than -100", NewRawObjectWithMgdl(), "recommended", map[string]interface{}{"net": -101, "correction": -50, "carb": 50},
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(-101, -100, 100), "/recommended/net", NewMeta())},
			),
			Entry("has net greater than 100", NewRawObjectWithMgdl(), "recommended", map[string]interface{}{"net": 101, "correction": -50, "carb": 50},
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(101, -100, 100), "/recommended/net", NewMeta())},
			),
			Entry("has correction less than -100", NewRawObjectWithMgdl(), "recommended", map[string]interface{}{"net": 50, "correction": -101, "carb": 50},
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(-101, -100, 100), "/recommended/correction", NewMeta())},
			),
			Entry("has correction greater than 100", NewRawObjectWithMgdl(), "recommended", map[string]interface{}{"net": 50, "correction": 101, "carb": 50},
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(101, -100, 100), "/recommended/correction", NewMeta())},
			),
			Entry("has carb less than 0", NewRawObjectWithMgdl(), "recommended", map[string]interface{}{"net": 50, "correction": -50, "carb": -1},
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(-1, 0, 100), "/recommended/carb", NewMeta())},
			),
			Entry("has carb greater than 100", NewRawObjectWithMgdl(), "recommended", map[string]interface{}{"net": 50, "correction": -50, "carb": 101},
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(101, 0, 100), "/recommended/carb", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("has net more or equal -100", NewRawObjectWithMgdl(), "recommended", map[string]interface{}{"net": -100, "correction": -50, "carb": 50}),
			Entry("has net less or equal 100", NewRawObjectWithMgdl(), "recommended", map[string]interface{}{"net": 100, "correction": -50, "carb": 50}),
			Entry("has correction more or equal -100", NewRawObjectWithMgdl(), "recommended", map[string]interface{}{"net": 10, "correction": -100, "carb": 50}),
			Entry("has correction less or equal 100", NewRawObjectWithMgdl(), "recommended", map[string]interface{}{"net": 10, "correction": 100, "carb": 50}),
			Entry("has carb more or equal 0", NewRawObjectWithMgdl(), "recommended", map[string]interface{}{"net": -100, "correction": -50, "carb": 0}),
			Entry("has carb less or equal 50", NewRawObjectWithMgdl(), "recommended", map[string]interface{}{"net": 100, "correction": -50, "carb": 50}),
		)
	})

	Context("bolus", func() {
		DescribeTable("invalid when type", testing.ExpectFieldNotValid,
			Entry("is missing", NewRawObjectWithMgdl(), "bolus", NewEmbeddedBolus(nil, "normal", 52.1, 0.0, 0),
				[]*service.Error{testing.ComposeError(service.ErrorValueNotExists(), "/bolus/type", NewMeta())},
			),
			Entry("is not valid", NewRawObjectWithMgdl(), "bolus", NewEmbeddedBolus("invalid", "normal", 52.1, 0.0, 0),
				[]*service.Error{testing.ComposeError(service.ErrorValueStringNotOneOf("invalid", []string{"bolus"}), "/bolus/type", NewMeta())},
			),
		)

		DescribeTable("invalid when subType", testing.ExpectFieldNotValid,
			Entry("is missing", NewRawObjectWithMgdl(), "bolus", NewEmbeddedBolus("bolus", nil, 0.0, 52.1, 0),
				[]*service.Error{testing.ComposeError(service.ErrorValueNotExists(), "/bolus/subType", NewMeta())},
			),
			Entry("is not valid", NewRawObjectWithMgdl(), "bolus", NewEmbeddedBolus("bolus", "invalid", 0.0, 52.1, 0),
				[]*service.Error{testing.ComposeError(service.ErrorValueStringNotOneOf("invalid", []string{"dual/square", "normal", "square"}), "/bolus/subType", NewMeta())},
			),
		)
	})

	Context("Normalize", func() {
		Context("blood glucose", func() {
			DescribeTable("when mmol/L", func(val, expected float64) {
				bolusCalculator := calculator.Init()
				units := glucose.MmolL
				bolusCalculator.Units = &units
				bolusCalculator.BloodGlucoseInput = &val
				bolusCalculator.InsulinSensitivity = &val
				bolusCalculator.BloodGlucoseTarget = &glucose.Target{Target: &val}

				testContext, err := context.NewStandard(log.NewNull())
				Expect(err).ToNot(HaveOccurred())
				Expect(testContext).ToNot(BeNil())
				standardNormalizer, err := normalizer.NewStandard(testContext)
				Expect(err).ToNot(HaveOccurred())
				Expect(standardNormalizer).ToNot(BeNil())
				bolusCalculator.Normalize(standardNormalizer)
				Expect(*bolusCalculator.Units).To(Equal(glucose.MmolL))
				Expect(*bolusCalculator.BloodGlucoseInput).To(Equal(expected))
				Expect(*bolusCalculator.InsulinSensitivity).To(Equal(expected))
				Expect(*bolusCalculator.BloodGlucoseTarget.Target).To(Equal(expected))
			},
				Entry("is expected lower bg value", 3.7, 3.7),
				Entry("is below max", 54.99, 54.99),
				Entry("is expected upper bg value", 23.0, 23.0),
			)

			DescribeTable("when mg/dL", func(val, expected float64) {
				bolusCalculator := calculator.Init()
				units := glucose.MgdL
				bolusCalculator.Units = &units
				bolusCalculator.BloodGlucoseInput = &val
				bolusCalculator.InsulinSensitivity = &val
				bolusCalculator.BloodGlucoseTarget = &glucose.Target{Target: &val}

				testContext, err := context.NewStandard(log.NewNull())
				Expect(err).ToNot(HaveOccurred())
				Expect(testContext).ToNot(BeNil())
				standardNormalizer, err := normalizer.NewStandard(testContext)
				Expect(err).ToNot(HaveOccurred())
				Expect(standardNormalizer).ToNot(BeNil())
				bolusCalculator.Normalize(standardNormalizer)
				Expect(*bolusCalculator.Units).To(Equal(glucose.MmolL))
				Expect(*bolusCalculator.BloodGlucoseInput).To(Equal(expected))
				Expect(*bolusCalculator.InsulinSensitivity).To(Equal(expected))
				Expect(*bolusCalculator.BloodGlucoseTarget.Target).To(Equal(expected))
			},
				Entry("is expected lower bg value", 60.0, 3.33045),
				Entry("is below max", glucose.MgdLUpperLimit, 55.50748),
				Entry("is expected upper bg value", 400.0, 22.20299),
			)
		})

		Context("bolus", func() {
			DescribeTable("valid when embedded", func(rawObject map[string]interface{}, field string, val interface{}) {
				calculatorDatum := testing.ParseAndNormalize(rawObject, field, val)
				calculatorBolus := calculatorDatum.(*calculator.Calculator)
				Expect(calculatorBolus.BolusID).To(Not(BeNil()))
			},
				Entry("is normal", NewRawObjectWithMgdl(), "bolus", NewEmbeddedBolus("bolus", "normal", 52.1, 0.0, 0)),
				Entry("is square", NewRawObjectWithMgdl(), "bolus", NewEmbeddedBolus("bolus", "square", 0.0, 52.1, 1000)),
				Entry("is dual/square", NewRawObjectWithMgdl(), "bolus", NewEmbeddedBolus("bolus", "dual/square", 52.1, 52.1, 1000)),
			)
		})
	})
})
