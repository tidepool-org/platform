package calculator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	testDataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/bolus/combination"
	testDataTypesBolusCombination "github.com/tidepool-org/platform/data/types/bolus/combination/test"
	"github.com/tidepool-org/platform/data/types/bolus/extended"
	testDataTypesBolusExtended "github.com/tidepool-org/platform/data/types/bolus/extended/test"
	"github.com/tidepool-org/platform/data/types/bolus/normal"
	testDataTypesBolusNormal "github.com/tidepool-org/platform/data/types/bolus/normal/test"
	"github.com/tidepool-org/platform/data/types/calculator"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &types.Meta{
		Type: "wizard",
	}
}

func NewCalculator(units *string) *calculator.Calculator {
	datum := calculator.New()
	datum.Base = *testDataTypes.NewBase()
	datum.Type = "wizard"
	datum.BloodGlucoseInput = pointer.FromFloat64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(units)))
	datum.BloodGlucoseTarget = testDataBloodGlucose.NewTarget(units)
	datum.CarbohydrateInput = pointer.FromFloat64(test.RandomFloat64FromRange(calculator.CarbohydrateInputMinimum, calculator.CarbohydrateInputMaximum))
	datum.InsulinCarbohydrateRatio = pointer.FromFloat64(test.RandomFloat64FromRange(calculator.InsulinCarbohydrateRatioMinimum, calculator.InsulinCarbohydrateRatioMaximum))
	datum.InsulinOnBoard = pointer.FromFloat64(test.RandomFloat64FromRange(calculator.InsulinOnBoardMinimum, calculator.InsulinOnBoardMaximum))
	datum.InsulinSensitivity = pointer.FromFloat64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(units)))
	datum.Recommended = NewRecommended()
	datum.Units = units
	return datum
}

func NewCalculatorWithBolusCombination(units *string) *calculator.Calculator {
	var bolus data.Datum
	bolus = testDataTypesBolusCombination.NewCombination()
	datum := NewCalculator(units)
	datum.Bolus = &bolus
	return datum
}

func NewCalculatorWithBolusExtended(units *string) *calculator.Calculator {
	var bolus data.Datum
	bolus = testDataTypesBolusExtended.NewExtended()
	datum := NewCalculator(units)
	datum.Bolus = &bolus
	return datum
}

func NewCalculatorWithBolusNormal(units *string) *calculator.Calculator {
	var bolus data.Datum
	bolus = testDataTypesBolusNormal.NewNormal()
	datum := NewCalculator(units)
	datum.Bolus = &bolus
	return datum
}

func NewCalculatorWithBolusID(units *string) *calculator.Calculator {
	datum := NewCalculator(units)
	datum.BolusID = pointer.FromString(dataTest.RandomID())
	return datum
}

func CloneCalculator(datum *calculator.Calculator) *calculator.Calculator {
	if datum == nil {
		return nil
	}
	clone := calculator.New()
	clone.Base = *testDataTypes.CloneBase(&datum.Base)
	clone.BloodGlucoseInput = test.CloneFloat64(datum.BloodGlucoseInput)
	clone.BloodGlucoseTarget = testDataBloodGlucose.CloneTarget(datum.BloodGlucoseTarget)
	if datum.Bolus != nil {
		switch bolus := (*datum.Bolus).(type) {
		case *combination.Combination:
			clone.Bolus = data.DatumAsPointer(testDataTypesBolusCombination.CloneCombination(bolus))
		case *extended.Extended:
			clone.Bolus = data.DatumAsPointer(testDataTypesBolusExtended.CloneExtended(bolus))
		case *normal.Normal:
			clone.Bolus = data.DatumAsPointer(testDataTypesBolusNormal.CloneNormal(bolus))
		}
	}
	clone.BolusID = test.CloneString(datum.BolusID)
	clone.CarbohydrateInput = test.CloneFloat64(datum.CarbohydrateInput)
	clone.InsulinCarbohydrateRatio = test.CloneFloat64(datum.InsulinCarbohydrateRatio)
	clone.InsulinOnBoard = test.CloneFloat64(datum.InsulinOnBoard)
	clone.InsulinSensitivity = test.CloneFloat64(datum.InsulinSensitivity)
	clone.Recommended = CloneRecommended(datum.Recommended)
	clone.Units = test.CloneString(datum.Units)
	return clone
}

var _ = Describe("Calculator", func() {
	It("Type is expected", func() {
		Expect(calculator.Type).To(Equal("wizard"))
	})

	It("CarbohydrateInputMaximum is expected", func() {
		Expect(calculator.CarbohydrateInputMaximum).To(Equal(1000.0))
	})

	It("CarbohydrateInputMinimum is expected", func() {
		Expect(calculator.CarbohydrateInputMinimum).To(Equal(0.0))
	})

	It("InsulinCarbohydrateRatioMaximum is expected", func() {
		Expect(calculator.InsulinCarbohydrateRatioMaximum).To(Equal(250.0))
	})

	It("InsulinCarbohydrateRatioMinimum is expected", func() {
		Expect(calculator.InsulinCarbohydrateRatioMinimum).To(Equal(0.0))
	})

	It("InsulinOnBoardMaximum is expected", func() {
		Expect(calculator.InsulinOnBoardMaximum).To(Equal(250.0))
	})

	It("InsulinOnBoardMinimum is expected", func() {
		Expect(calculator.InsulinOnBoardMinimum).To(Equal(0.0))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := calculator.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("wizard"))
			Expect(datum.BloodGlucoseInput).To(BeNil())
			Expect(datum.BloodGlucoseTarget).To(BeNil())
			Expect(datum.Bolus).To(BeNil())
			Expect(datum.BolusID).To(BeNil())
			Expect(datum.CarbohydrateInput).To(BeNil())
			Expect(datum.InsulinCarbohydrateRatio).To(BeNil())
			Expect(datum.InsulinOnBoard).To(BeNil())
			Expect(datum.InsulinSensitivity).To(BeNil())
			Expect(datum.Recommended).To(BeNil())
			Expect(datum.Units).To(BeNil())
		})
	})

	Context("Calculator", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *calculator.Calculator, units *string), expectedErrors ...error) {
					datum := NewCalculator(units)
					mutator(datum, units)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) {},
				),
				Entry("type missing",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.Type = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &types.Meta{}),
				),
				Entry("type invalid",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.Type = "invalidType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "wizard"), "/type", &types.Meta{Type: "invalidType"}),
				),
				Entry("type wizard",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.Type = "wizard" },
				),
				Entry("units missing; blood glucose input missing",
					nil,
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseInput = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; blood glucose input out of range (lower)",
					nil,
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseInput = pointer.FromFloat64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; blood glucose input in range (lower)",
					nil,
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseInput = pointer.FromFloat64(0.0) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; blood glucose input in range (upper)",
					nil,
					func(datum *calculator.Calculator, units *string) {
						datum.BloodGlucoseInput = pointer.FromFloat64(1000.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; blood glucose input out of range (upper)",
					nil,
					func(datum *calculator.Calculator, units *string) {
						datum.BloodGlucoseInput = pointer.FromFloat64(1000.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; blood glucose target missing",
					nil,
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseTarget = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; blood glucose target invalid",
					nil,
					func(datum *calculator.Calculator, units *string) {
						datum.BloodGlucoseTarget = dataBloodGlucose.NewTarget()
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgTarget/target", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; blood glucose target valid",
					nil,
					func(datum *calculator.Calculator, units *string) {
						datum.BloodGlucoseTarget = testDataBloodGlucose.NewTarget(units)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; bolus missing",
					nil,
					func(datum *calculator.Calculator, units *string) { datum.Bolus = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; bolus id missing",
					nil,
					func(datum *calculator.Calculator, units *string) { datum.BolusID = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; carbohydrate input missing",
					nil,
					func(datum *calculator.Calculator, units *string) { datum.CarbohydrateInput = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; carbohydrate input out of range (lower)",
					nil,
					func(datum *calculator.Calculator, units *string) { datum.CarbohydrateInput = pointer.FromFloat64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/carbInput", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; carbohydrate input in range (lower)",
					nil,
					func(datum *calculator.Calculator, units *string) { datum.CarbohydrateInput = pointer.FromFloat64(0.0) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; carbohydrate input in range (upper)",
					nil,
					func(datum *calculator.Calculator, units *string) {
						datum.CarbohydrateInput = pointer.FromFloat64(1000.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; carbohydrate input out of range (upper)",
					nil,
					func(datum *calculator.Calculator, units *string) {
						datum.CarbohydrateInput = pointer.FromFloat64(1000.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/carbInput", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; insulin carbohydrate ratio missing",
					nil,
					func(datum *calculator.Calculator, units *string) { datum.InsulinCarbohydrateRatio = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; insulin carbohydrate ratio out of range (lower)",
					nil,
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinCarbohydrateRatio = pointer.FromFloat64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 250.0), "/insulinCarbRatio", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; insulin carbohydrate ratio in range (lower)",
					nil,
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinCarbohydrateRatio = pointer.FromFloat64(0.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; insulin carbohydrate ratio in range (upper)",
					nil,
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinCarbohydrateRatio = pointer.FromFloat64(250.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; insulin carbohydrate ratio out of range (upper)",
					nil,
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinCarbohydrateRatio = pointer.FromFloat64(250.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, 0.0, 250.0), "/insulinCarbRatio", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; insulin on board missing",
					nil,
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; insulin on board out of range (lower)",
					nil,
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = pointer.FromFloat64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 250.0), "/insulinOnBoard", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; insulin on board in range (lower)",
					nil,
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = pointer.FromFloat64(0.0) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; insulin on board in range (upper)",
					nil,
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = pointer.FromFloat64(250.0) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; insulin on board out of range (upper)",
					nil,
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = pointer.FromFloat64(250.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, 0.0, 250.0), "/insulinOnBoard", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; insulin sensitivity missing",
					nil,
					func(datum *calculator.Calculator, units *string) { datum.InsulinSensitivity = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; insulin sensitivity out of range (lower)",
					nil,
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinSensitivity = pointer.FromFloat64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; insulin sensitivity in range (lower)",
					nil,
					func(datum *calculator.Calculator, units *string) { datum.InsulinSensitivity = pointer.FromFloat64(0.0) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; insulin sensitivity in range (upper)",
					nil,
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinSensitivity = pointer.FromFloat64(55.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; insulin sensitivity out of range (upper)",
					nil,
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinSensitivity = pointer.FromFloat64(1000.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; recommended missing",
					nil,
					func(datum *calculator.Calculator, units *string) { datum.Recommended = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; recommended invalid",
					nil,
					func(datum *calculator.Calculator, units *string) {
						datum.Recommended.Carbohydrate = pointer.FromFloat64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/recommended/carb", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; recommended valid",
					nil,
					func(datum *calculator.Calculator, units *string) { datum.Recommended = NewRecommended() },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units invalid; blood glucose input missing",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseInput = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; blood glucose input out of range (lower)",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseInput = pointer.FromFloat64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; blood glucose input in range (lower)",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseInput = pointer.FromFloat64(0.0) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; blood glucose input in range (upper)",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) {
						datum.BloodGlucoseInput = pointer.FromFloat64(1000.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; blood glucose input out of range (upper)",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) {
						datum.BloodGlucoseInput = pointer.FromFloat64(1000.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; blood glucose target missing",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseTarget = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; blood glucose target invalid",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) {
						datum.BloodGlucoseTarget = dataBloodGlucose.NewTarget()
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgTarget/target", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; blood glucose target valid",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) {
						datum.BloodGlucoseTarget = testDataBloodGlucose.NewTarget(units)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; bolus missing",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) { datum.Bolus = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; bolus id missing",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) { datum.BolusID = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; carbohydrate input missing",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) { datum.CarbohydrateInput = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; carbohydrate input out of range (lower)",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) { datum.CarbohydrateInput = pointer.FromFloat64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/carbInput", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; carbohydrate input in range (lower)",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) { datum.CarbohydrateInput = pointer.FromFloat64(0.0) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; carbohydrate input in range (upper)",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) {
						datum.CarbohydrateInput = pointer.FromFloat64(1000.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; carbohydrate input out of range (upper)",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) {
						datum.CarbohydrateInput = pointer.FromFloat64(1000.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/carbInput", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; insulin carbohydrate ratio missing",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinCarbohydrateRatio = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; insulin carbohydrate ratio out of range (lower)",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinCarbohydrateRatio = pointer.FromFloat64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 250.0), "/insulinCarbRatio", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; insulin carbohydrate ratio in range (lower)",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinCarbohydrateRatio = pointer.FromFloat64(0.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; insulin carbohydrate ratio in range (upper)",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinCarbohydrateRatio = pointer.FromFloat64(250.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; insulin carbohydrate ratio out of range (upper)",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinCarbohydrateRatio = pointer.FromFloat64(250.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, 0.0, 250.0), "/insulinCarbRatio", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; insulin on board missing",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; insulin on board out of range (lower)",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = pointer.FromFloat64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 250.0), "/insulinOnBoard", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; insulin on board in range (lower)",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = pointer.FromFloat64(0.0) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; insulin on board in range (upper)",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = pointer.FromFloat64(250.0) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; insulin on board out of range (upper)",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = pointer.FromFloat64(250.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, 0.0, 250.0), "/insulinOnBoard", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; insulin sensitivity missing",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinSensitivity = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; insulin sensitivity out of range (lower)",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinSensitivity = pointer.FromFloat64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; insulin sensitivity in range (lower)",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinSensitivity = pointer.FromFloat64(0.0) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; insulin sensitivity in range (upper)",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinSensitivity = pointer.FromFloat64(55.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; insulin sensitivity out of range (upper)",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinSensitivity = pointer.FromFloat64(1000.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; recommended missing",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) { datum.Recommended = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; recommended invalid",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) {
						datum.Recommended.Carbohydrate = pointer.FromFloat64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/recommended/carb", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; recommended valid",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) { datum.Recommended = NewRecommended() },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units mmol/L; blood glucose input missing",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseInput = nil },
				),
				Entry("units mmol/L; blood glucose input out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseInput = pointer.FromFloat64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/bgInput", NewMeta()),
				),
				Entry("units mmol/L; blood glucose input in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseInput = pointer.FromFloat64(0.0) },
				),
				Entry("units mmol/L; blood glucose input in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseInput = pointer.FromFloat64(55.0) },
				),
				Entry("units mmol/L; blood glucose input out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseInput = pointer.FromFloat64(55.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/bgInput", NewMeta()),
				),
				Entry("units mmol/L; blood glucose target missing",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseTarget = nil },
				),
				Entry("units mmol/L; blood glucose target invalid",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) {
						datum.BloodGlucoseTarget = dataBloodGlucose.NewTarget()
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgTarget/target", NewMeta()),
				),
				Entry("units mmol/L; blood glucose target valid",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) {
						datum.BloodGlucoseTarget = testDataBloodGlucose.NewTarget(units)
					},
				),
				Entry("units mmol/L; bolus missing",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.Bolus = nil },
				),
				Entry("units mmol/L; bolus id missing",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.BolusID = nil },
				),
				Entry("units mmol/L; carbohydrate input missing",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.CarbohydrateInput = nil },
				),
				Entry("units mmol/L; carbohydrate input out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.CarbohydrateInput = pointer.FromFloat64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/carbInput", NewMeta()),
				),
				Entry("units mmol/L; carbohydrate input in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.CarbohydrateInput = pointer.FromFloat64(0.0) },
				),
				Entry("units mmol/L; carbohydrate input in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) {
						datum.CarbohydrateInput = pointer.FromFloat64(1000.0)
					},
				),
				Entry("units mmol/L; carbohydrate input out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) {
						datum.CarbohydrateInput = pointer.FromFloat64(1000.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/carbInput", NewMeta()),
				),
				Entry("units mmol/L; insulin carbohydrate ratio missing",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinCarbohydrateRatio = nil },
				),
				Entry("units mmol/L; insulin carbohydrate ratio out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinCarbohydrateRatio = pointer.FromFloat64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 250.0), "/insulinCarbRatio", NewMeta()),
				),
				Entry("units mmol/L; insulin carbohydrate ratio in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinCarbohydrateRatio = pointer.FromFloat64(0.0)
					},
				),
				Entry("units mmol/L; insulin carbohydrate ratio in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinCarbohydrateRatio = pointer.FromFloat64(250.0)
					},
				),
				Entry("units mmol/L; insulin carbohydrate ratio out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinCarbohydrateRatio = pointer.FromFloat64(250.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, 0.0, 250.0), "/insulinCarbRatio", NewMeta()),
				),
				Entry("units mmol/L; insulin on board missing",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = nil },
				),
				Entry("units mmol/L; insulin on board out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = pointer.FromFloat64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 250.0), "/insulinOnBoard", NewMeta()),
				),
				Entry("units mmol/L; insulin on board in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = pointer.FromFloat64(0.0) },
				),
				Entry("units mmol/L; insulin on board in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = pointer.FromFloat64(250.0) },
				),
				Entry("units mmol/L; insulin on board out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = pointer.FromFloat64(250.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, 0.0, 250.0), "/insulinOnBoard", NewMeta()),
				),
				Entry("units mmol/L; insulin sensitivity missing",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinSensitivity = nil },
				),
				Entry("units mmol/L; insulin sensitivity out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinSensitivity = pointer.FromFloat64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/insulinSensitivity", NewMeta()),
				),
				Entry("units mmol/L; insulin sensitivity in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinSensitivity = pointer.FromFloat64(0.0) },
				),
				Entry("units mmol/L; insulin sensitivity in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinSensitivity = pointer.FromFloat64(55.0)
					},
				),
				Entry("units mmol/L; insulin sensitivity out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinSensitivity = pointer.FromFloat64(55.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/insulinSensitivity", NewMeta()),
				),
				Entry("units mmol/L; recommended missing",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.Recommended = nil },
				),
				Entry("units mmol/L; recommended invalid",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) {
						datum.Recommended.Carbohydrate = pointer.FromFloat64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/recommended/carb", NewMeta()),
				),
				Entry("units mmol/L; recommended valid",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.Recommended = NewRecommended() },
				),
				Entry("units mmol/l; blood glucose input missing",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseInput = nil },
				),
				Entry("units mmol/l; blood glucose input out of range (lower)",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseInput = pointer.FromFloat64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/bgInput", NewMeta()),
				),
				Entry("units mmol/l; blood glucose input in range (lower)",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseInput = pointer.FromFloat64(0.0) },
				),
				Entry("units mmol/l; blood glucose input in range (upper)",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseInput = pointer.FromFloat64(55.0) },
				),
				Entry("units mmol/l; blood glucose input out of range (upper)",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseInput = pointer.FromFloat64(55.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/bgInput", NewMeta()),
				),
				Entry("units mmol/l; blood glucose target missing",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseTarget = nil },
				),
				Entry("units mmol/l; blood glucose target invalid",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) {
						datum.BloodGlucoseTarget = dataBloodGlucose.NewTarget()
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgTarget/target", NewMeta()),
				),
				Entry("units mmol/l; blood glucose target valid",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) {
						datum.BloodGlucoseTarget = testDataBloodGlucose.NewTarget(units)
					},
				),
				Entry("units mmol/l; bolus missing",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) { datum.Bolus = nil },
				),
				Entry("units mmol/l; bolus id missing",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) { datum.BolusID = nil },
				),
				Entry("units mmol/l; carbohydrate input missing",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) { datum.CarbohydrateInput = nil },
				),
				Entry("units mmol/l; carbohydrate input out of range (lower)",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) { datum.CarbohydrateInput = pointer.FromFloat64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/carbInput", NewMeta()),
				),
				Entry("units mmol/l; carbohydrate input in range (lower)",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) { datum.CarbohydrateInput = pointer.FromFloat64(0.0) },
				),
				Entry("units mmol/l; carbohydrate input in range (upper)",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) {
						datum.CarbohydrateInput = pointer.FromFloat64(1000.0)
					},
				),
				Entry("units mmol/l; carbohydrate input out of range (upper)",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) {
						datum.CarbohydrateInput = pointer.FromFloat64(1000.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/carbInput", NewMeta()),
				),
				Entry("units mmol/l; insulin carbohydrate ratio missing",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinCarbohydrateRatio = nil },
				),
				Entry("units mmol/l; insulin carbohydrate ratio out of range (lower)",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinCarbohydrateRatio = pointer.FromFloat64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 250.0), "/insulinCarbRatio", NewMeta()),
				),
				Entry("units mmol/l; insulin carbohydrate ratio in range (lower)",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinCarbohydrateRatio = pointer.FromFloat64(0.0)
					},
				),
				Entry("units mmol/l; insulin carbohydrate ratio in range (upper)",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinCarbohydrateRatio = pointer.FromFloat64(250.0)
					},
				),
				Entry("units mmol/l; insulin carbohydrate ratio out of range (upper)",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinCarbohydrateRatio = pointer.FromFloat64(250.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, 0.0, 250.0), "/insulinCarbRatio", NewMeta()),
				),
				Entry("units mmol/l; insulin on board missing",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = nil },
				),
				Entry("units mmol/l; insulin on board out of range (lower)",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = pointer.FromFloat64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 250.0), "/insulinOnBoard", NewMeta()),
				),
				Entry("units mmol/l; insulin on board in range (lower)",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = pointer.FromFloat64(0.0) },
				),
				Entry("units mmol/l; insulin on board in range (upper)",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = pointer.FromFloat64(250.0) },
				),
				Entry("units mmol/l; insulin on board out of range (upper)",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = pointer.FromFloat64(250.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, 0.0, 250.0), "/insulinOnBoard", NewMeta()),
				),
				Entry("units mmol/l; insulin sensitivity missing",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinSensitivity = nil },
				),
				Entry("units mmol/l; insulin sensitivity out of range (lower)",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinSensitivity = pointer.FromFloat64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/insulinSensitivity", NewMeta()),
				),
				Entry("units mmol/l; insulin sensitivity in range (lower)",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinSensitivity = pointer.FromFloat64(0.0) },
				),
				Entry("units mmol/l; insulin sensitivity in range (upper)",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinSensitivity = pointer.FromFloat64(55.0)
					},
				),
				Entry("units mmol/l; insulin sensitivity out of range (upper)",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinSensitivity = pointer.FromFloat64(55.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/insulinSensitivity", NewMeta()),
				),
				Entry("units mmol/l; recommended missing",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) { datum.Recommended = nil },
				),
				Entry("units mmol/l; recommended invalid",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) {
						datum.Recommended.Carbohydrate = pointer.FromFloat64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/recommended/carb", NewMeta()),
				),
				Entry("units mmol/l; recommended valid",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) { datum.Recommended = NewRecommended() },
				),
				Entry("units mg/dL; blood glucose input missing",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseInput = nil },
				),
				Entry("units mg/dL; blood glucose input out of range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseInput = pointer.FromFloat64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/bgInput", NewMeta()),
				),
				Entry("units mg/dL; blood glucose input in range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseInput = pointer.FromFloat64(0.0) },
				),
				Entry("units mg/dL; blood glucose input in range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) {
						datum.BloodGlucoseInput = pointer.FromFloat64(1000.0)
					},
				),
				Entry("units mg/dL; blood glucose input out of range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) {
						datum.BloodGlucoseInput = pointer.FromFloat64(1000.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/bgInput", NewMeta()),
				),
				Entry("units mg/dL; blood glucose target missing",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseTarget = nil },
				),
				Entry("units mg/dL; blood glucose target invalid",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) {
						datum.BloodGlucoseTarget = dataBloodGlucose.NewTarget()
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgTarget/target", NewMeta()),
				),
				Entry("units mg/dL; blood glucose target valid",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) {
						datum.BloodGlucoseTarget = testDataBloodGlucose.NewTarget(units)
					},
				),
				Entry("units mg/dL; bolus missing",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) { datum.Bolus = nil },
				),
				Entry("units mg/dL; bolus id missing",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) { datum.BolusID = nil },
				),
				Entry("units mg/dL; carbohydrate input missing",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) { datum.CarbohydrateInput = nil },
				),
				Entry("units mg/dL; carbohydrate input out of range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) { datum.CarbohydrateInput = pointer.FromFloat64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/carbInput", NewMeta()),
				),
				Entry("units mg/dL; carbohydrate input in range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) { datum.CarbohydrateInput = pointer.FromFloat64(0.0) },
				),
				Entry("units mg/dL; carbohydrate input in range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) {
						datum.CarbohydrateInput = pointer.FromFloat64(1000.0)
					},
				),
				Entry("units mg/dL; carbohydrate input out of range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) {
						datum.CarbohydrateInput = pointer.FromFloat64(1000.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/carbInput", NewMeta()),
				),
				Entry("units mg/dL; insulin carbohydrate ratio missing",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinCarbohydrateRatio = nil },
				),
				Entry("units mg/dL; insulin carbohydrate ratio out of range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinCarbohydrateRatio = pointer.FromFloat64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 250.0), "/insulinCarbRatio", NewMeta()),
				),
				Entry("units mg/dL; insulin carbohydrate ratio in range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinCarbohydrateRatio = pointer.FromFloat64(0.0)
					},
				),
				Entry("units mg/dL; insulin carbohydrate ratio in range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinCarbohydrateRatio = pointer.FromFloat64(250.0)
					},
				),
				Entry("units mg/dL; insulin carbohydrate ratio out of range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinCarbohydrateRatio = pointer.FromFloat64(250.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, 0.0, 250.0), "/insulinCarbRatio", NewMeta()),
				),
				Entry("units mg/dL; insulin on board missing",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = nil },
				),
				Entry("units mg/dL; insulin on board out of range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = pointer.FromFloat64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 250.0), "/insulinOnBoard", NewMeta()),
				),
				Entry("units mg/dL; insulin on board in range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = pointer.FromFloat64(0.0) },
				),
				Entry("units mg/dL; insulin on board in range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = pointer.FromFloat64(250.0) },
				),
				Entry("units mg/dL; insulin on board out of range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = pointer.FromFloat64(250.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, 0.0, 250.0), "/insulinOnBoard", NewMeta()),
				),
				Entry("units mg/dL; insulin sensitivity missing",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinSensitivity = nil },
				),
				Entry("units mg/dL; insulin sensitivity out of range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinSensitivity = pointer.FromFloat64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/insulinSensitivity", NewMeta()),
				),
				Entry("units mg/dL; insulin sensitivity in range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinSensitivity = pointer.FromFloat64(0.0) },
				),
				Entry("units mg/dL; insulin sensitivity in range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinSensitivity = pointer.FromFloat64(1000.0)
					},
				),
				Entry("units mg/dL; insulin sensitivity out of range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinSensitivity = pointer.FromFloat64(1000.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/insulinSensitivity", NewMeta()),
				),
				Entry("units mg/dL; recommended missing",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) { datum.Recommended = nil },
				),
				Entry("units mg/dL; recommended invalid",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) {
						datum.Recommended.Carbohydrate = pointer.FromFloat64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/recommended/carb", NewMeta()),
				),
				Entry("units mg/dL; recommended valid",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) { datum.Recommended = NewRecommended() },
				),
				Entry("units mg/dl; blood glucose input missing",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseInput = nil },
				),
				Entry("units mg/dl; blood glucose input out of range (lower)",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseInput = pointer.FromFloat64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/bgInput", NewMeta()),
				),
				Entry("units mg/dl; blood glucose input in range (lower)",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseInput = pointer.FromFloat64(0.0) },
				),
				Entry("units mg/dl; blood glucose input in range (upper)",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) {
						datum.BloodGlucoseInput = pointer.FromFloat64(1000.0)
					},
				),
				Entry("units mg/dl; blood glucose input out of range (upper)",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) {
						datum.BloodGlucoseInput = pointer.FromFloat64(1000.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/bgInput", NewMeta()),
				),
				Entry("units mg/dl; blood glucose target missing",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) { datum.BloodGlucoseTarget = nil },
				),
				Entry("units mg/dl; blood glucose target invalid",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) {
						datum.BloodGlucoseTarget = dataBloodGlucose.NewTarget()
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgTarget/target", NewMeta()),
				),
				Entry("units mg/dl; blood glucose target valid",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) {
						datum.BloodGlucoseTarget = testDataBloodGlucose.NewTarget(units)
					},
				),
				Entry("units mg/dl; bolus missing",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) { datum.Bolus = nil },
				),
				Entry("units mg/dl; bolus id missing",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) { datum.BolusID = nil },
				),
				Entry("units mg/dl; carbohydrate input missing",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) { datum.CarbohydrateInput = nil },
				),
				Entry("units mg/dl; carbohydrate input out of range (lower)",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) { datum.CarbohydrateInput = pointer.FromFloat64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/carbInput", NewMeta()),
				),
				Entry("units mg/dl; carbohydrate input in range (lower)",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) { datum.CarbohydrateInput = pointer.FromFloat64(0.0) },
				),
				Entry("units mg/dl; carbohydrate input in range (upper)",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) {
						datum.CarbohydrateInput = pointer.FromFloat64(1000.0)
					},
				),
				Entry("units mg/dl; carbohydrate input out of range (upper)",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) {
						datum.CarbohydrateInput = pointer.FromFloat64(1000.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/carbInput", NewMeta()),
				),
				Entry("units mg/dl; insulin carbohydrate ratio missing",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinCarbohydrateRatio = nil },
				),
				Entry("units mg/dl; insulin carbohydrate ratio out of range (lower)",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinCarbohydrateRatio = pointer.FromFloat64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 250.0), "/insulinCarbRatio", NewMeta()),
				),
				Entry("units mg/dl; insulin carbohydrate ratio in range (lower)",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinCarbohydrateRatio = pointer.FromFloat64(0.0)
					},
				),
				Entry("units mg/dl; insulin carbohydrate ratio in range (upper)",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinCarbohydrateRatio = pointer.FromFloat64(250.0)
					},
				),
				Entry("units mg/dl; insulin carbohydrate ratio out of range (upper)",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinCarbohydrateRatio = pointer.FromFloat64(250.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, 0.0, 250.0), "/insulinCarbRatio", NewMeta()),
				),
				Entry("units mg/dl; insulin on board missing",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = nil },
				),
				Entry("units mg/dl; insulin on board out of range (lower)",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = pointer.FromFloat64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 250.0), "/insulinOnBoard", NewMeta()),
				),
				Entry("units mg/dl; insulin on board in range (lower)",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = pointer.FromFloat64(0.0) },
				),
				Entry("units mg/dl; insulin on board in range (upper)",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = pointer.FromFloat64(250.0) },
				),
				Entry("units mg/dl; insulin on board out of range (upper)",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinOnBoard = pointer.FromFloat64(250.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, 0.0, 250.0), "/insulinOnBoard", NewMeta()),
				),
				Entry("units mg/dl; insulin sensitivity missing",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinSensitivity = nil },
				),
				Entry("units mg/dl; insulin sensitivity out of range (lower)",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinSensitivity = pointer.FromFloat64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/insulinSensitivity", NewMeta()),
				),
				Entry("units mg/dl; insulin sensitivity in range (lower)",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) { datum.InsulinSensitivity = pointer.FromFloat64(0.0) },
				),
				Entry("units mg/dl; insulin sensitivity in range (upper)",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinSensitivity = pointer.FromFloat64(1000.0)
					},
				),
				Entry("units mg/dl; insulin sensitivity out of range (upper)",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) {
						datum.InsulinSensitivity = pointer.FromFloat64(1000.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/insulinSensitivity", NewMeta()),
				),
				Entry("units mg/dl; recommended missing",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) { datum.Recommended = nil },
				),
				Entry("units mg/dl; recommended invalid",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) {
						datum.Recommended.Carbohydrate = pointer.FromFloat64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/recommended/carb", NewMeta()),
				),
				Entry("units mg/dl; recommended valid",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) { datum.Recommended = NewRecommended() },
				),
			)

			DescribeTable("validates the datum with origin external",
				func(units *string, mutator func(datum *calculator.Calculator, units *string), expectedErrors ...error) {
					datum := NewCalculator(units)
					mutator(datum, units)
					testDataTypes.ValidateWithOrigin(datum, structure.OriginExternal, expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) {},
				),
				Entry("units missing; bolus missing",
					nil,
					func(datum *calculator.Calculator, units *string) { datum.Bolus = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; bolus combination invalid",
					nil,
					func(datum *calculator.Calculator, units *string) {
						bolus := testDataTypesBolusCombination.NewCombination()
						bolus.Extended = nil
						bolus.Normal = nil
						datum.Bolus = data.DatumAsPointer(bolus)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/normal", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; bolus combination valid",
					nil,
					func(datum *calculator.Calculator, units *string) {
						datum.Bolus = data.DatumAsPointer(testDataTypesBolusCombination.NewCombination())
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; bolus extended invalid",
					nil,
					func(datum *calculator.Calculator, units *string) {
						bolus := testDataTypesBolusExtended.NewExtended()
						bolus.Extended = nil
						datum.Bolus = data.DatumAsPointer(bolus)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; bolus extended valid",
					nil,
					func(datum *calculator.Calculator, units *string) {
						datum.Bolus = data.DatumAsPointer(testDataTypesBolusExtended.NewExtended())
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; bolus normal invalid",
					nil,
					func(datum *calculator.Calculator, units *string) {
						bolus := testDataTypesBolusNormal.NewNormal()
						bolus.Normal = nil
						datum.Bolus = data.DatumAsPointer(bolus)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/normal", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; bolus normal valid",
					nil,
					func(datum *calculator.Calculator, units *string) {
						datum.Bolus = data.DatumAsPointer(testDataTypesBolusNormal.NewNormal())
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; bolus id missing",
					nil,
					func(datum *calculator.Calculator, units *string) { datum.BolusID = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units invalid; bolus missing",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) { datum.Bolus = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; bolus combination invalid",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) {
						bolus := testDataTypesBolusCombination.NewCombination()
						bolus.Extended = nil
						bolus.Normal = nil
						datum.Bolus = data.DatumAsPointer(bolus)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/normal", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; bolus combination valid",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) {
						datum.Bolus = data.DatumAsPointer(testDataTypesBolusCombination.NewCombination())
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; bolus extended invalid",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) {
						bolus := testDataTypesBolusExtended.NewExtended()
						bolus.Extended = nil
						datum.Bolus = data.DatumAsPointer(bolus)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; bolus extended valid",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) {
						datum.Bolus = data.DatumAsPointer(testDataTypesBolusExtended.NewExtended())
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; bolus normal invalid",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) {
						bolus := testDataTypesBolusNormal.NewNormal()
						bolus.Normal = nil
						datum.Bolus = data.DatumAsPointer(bolus)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/normal", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; bolus normal valid",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) {
						datum.Bolus = data.DatumAsPointer(testDataTypesBolusNormal.NewNormal())
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; bolus id missing",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) { datum.BolusID = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units mmol/L; bolus missing",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.Bolus = nil },
				),
				Entry("units mmol/L; bolus combination invalid",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) {
						bolus := testDataTypesBolusCombination.NewCombination()
						bolus.Extended = nil
						bolus.Normal = nil
						datum.Bolus = data.DatumAsPointer(bolus)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/normal", NewMeta()),
				),
				Entry("units mmol/L; bolus combination valid",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) {
						datum.Bolus = data.DatumAsPointer(testDataTypesBolusCombination.NewCombination())
					},
				),
				Entry("units mmol/L; bolus extended invalid",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) {
						bolus := testDataTypesBolusExtended.NewExtended()
						bolus.Extended = nil
						datum.Bolus = data.DatumAsPointer(bolus)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/extended", NewMeta()),
				),
				Entry("units mmol/L; bolus extended valid",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) {
						datum.Bolus = data.DatumAsPointer(testDataTypesBolusExtended.NewExtended())
					},
				),
				Entry("units mmol/L; bolus normal invalid",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) {
						bolus := testDataTypesBolusNormal.NewNormal()
						bolus.Normal = nil
						datum.Bolus = data.DatumAsPointer(bolus)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/normal", NewMeta()),
				),
				Entry("units mmol/L; bolus normal valid",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) {
						datum.Bolus = data.DatumAsPointer(testDataTypesBolusNormal.NewNormal())
					},
				),
				Entry("units mmol/L; bolus id missing",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.BolusID = nil },
				),
				Entry("units mmol/l; bolus missing",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) { datum.Bolus = nil },
				),
				Entry("units mmol/l; bolus combination invalid",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) {
						bolus := testDataTypesBolusCombination.NewCombination()
						bolus.Extended = nil
						bolus.Normal = nil
						datum.Bolus = data.DatumAsPointer(bolus)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/normal", NewMeta()),
				),
				Entry("units mmol/l; bolus combination valid",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) {
						datum.Bolus = data.DatumAsPointer(testDataTypesBolusCombination.NewCombination())
					},
				),
				Entry("units mmol/l; bolus extended invalid",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) {
						bolus := testDataTypesBolusExtended.NewExtended()
						bolus.Extended = nil
						datum.Bolus = data.DatumAsPointer(bolus)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/extended", NewMeta()),
				),
				Entry("units mmol/l; bolus extended valid",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) {
						datum.Bolus = data.DatumAsPointer(testDataTypesBolusExtended.NewExtended())
					},
				),
				Entry("units mmol/l; bolus normal invalid",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) {
						bolus := testDataTypesBolusNormal.NewNormal()
						bolus.Normal = nil
						datum.Bolus = data.DatumAsPointer(bolus)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/normal", NewMeta()),
				),
				Entry("units mmol/l; bolus normal valid",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) {
						datum.Bolus = data.DatumAsPointer(testDataTypesBolusNormal.NewNormal())
					},
				),
				Entry("units mmol/l; bolus id missing",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) { datum.BolusID = nil },
				),
				Entry("units mg/dL; bolus missing",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) { datum.Bolus = nil },
				),
				Entry("units mg/dL; bolus combination invalid",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) {
						bolus := testDataTypesBolusCombination.NewCombination()
						bolus.Extended = nil
						bolus.Normal = nil
						datum.Bolus = data.DatumAsPointer(bolus)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/normal", NewMeta()),
				),
				Entry("units mg/dL; bolus combination valid",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) {
						datum.Bolus = data.DatumAsPointer(testDataTypesBolusCombination.NewCombination())
					},
				),
				Entry("units mg/dL; bolus extended invalid",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) {
						bolus := testDataTypesBolusExtended.NewExtended()
						bolus.Extended = nil
						datum.Bolus = data.DatumAsPointer(bolus)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/extended", NewMeta()),
				),
				Entry("units mg/dL; bolus extended valid",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) {
						datum.Bolus = data.DatumAsPointer(testDataTypesBolusExtended.NewExtended())
					},
				),
				Entry("units mg/dL; bolus normal invalid",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) {
						bolus := testDataTypesBolusNormal.NewNormal()
						bolus.Normal = nil
						datum.Bolus = data.DatumAsPointer(bolus)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/normal", NewMeta()),
				),
				Entry("units mg/dL; bolus normal valid",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) {
						datum.Bolus = data.DatumAsPointer(testDataTypesBolusNormal.NewNormal())
					},
				),
				Entry("units mg/dL; bolus id missing",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) { datum.BolusID = nil },
				),
				Entry("units mg/dl; bolus missing",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) { datum.Bolus = nil },
				),
				Entry("units mg/dl; bolus combination invalid",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) {
						bolus := testDataTypesBolusCombination.NewCombination()
						bolus.Extended = nil
						bolus.Normal = nil
						datum.Bolus = data.DatumAsPointer(bolus)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/normal", NewMeta()),
				),
				Entry("units mg/dl; bolus combination valid",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) {
						datum.Bolus = data.DatumAsPointer(testDataTypesBolusCombination.NewCombination())
					},
				),
				Entry("units mg/dl; bolus extended invalid",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) {
						bolus := testDataTypesBolusExtended.NewExtended()
						bolus.Extended = nil
						datum.Bolus = data.DatumAsPointer(bolus)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/extended", NewMeta()),
				),
				Entry("units mg/dl; bolus extended valid",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) {
						datum.Bolus = data.DatumAsPointer(testDataTypesBolusExtended.NewExtended())
					},
				),
				Entry("units mg/dl; bolus normal invalid",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) {
						bolus := testDataTypesBolusNormal.NewNormal()
						bolus.Normal = nil
						datum.Bolus = data.DatumAsPointer(bolus)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/normal", NewMeta()),
				),
				Entry("units mg/dl; bolus normal valid",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) {
						datum.Bolus = data.DatumAsPointer(testDataTypesBolusNormal.NewNormal())
					},
				),
				Entry("units mg/dl; bolus id missing",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) { datum.BolusID = nil },
				),
			)

			DescribeTable("validates the datum with origin internal/store",
				func(units *string, mutator func(datum *calculator.Calculator, units *string), expectedErrors ...error) {
					datum := NewCalculatorWithBolusID(units)
					mutator(datum, units)
					testDataTypes.ValidateWithOrigin(datum, structure.OriginInternal, expectedErrors...)
					testDataTypes.ValidateWithOrigin(datum, structure.OriginStore, expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) {},
				),
				Entry("units missing; bolus missing",
					nil,
					func(datum *calculator.Calculator, units *string) { datum.Bolus = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; bolus exists",
					nil,
					func(datum *calculator.Calculator, units *string) {
						datum.Bolus = data.DatumAsPointer(testDataTypesBolusCombination.NewCombination())
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/bolus", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; bolus id missing",
					nil,
					func(datum *calculator.Calculator, units *string) { datum.BolusID = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; bolus id empty",
					nil,
					func(datum *calculator.Calculator, units *string) { datum.BolusID = pointer.FromString("") },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/bolusId", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; bolus id exists",
					nil,
					func(datum *calculator.Calculator, units *string) {
						datum.BolusID = pointer.FromString(dataTest.RandomID())
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units invalid; bolus missing",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) { datum.Bolus = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; bolus exists",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) {
						datum.Bolus = data.DatumAsPointer(testDataTypesBolusCombination.NewCombination())
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/bolus", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; bolus id missing",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) { datum.BolusID = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; bolus id empty",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) { datum.BolusID = pointer.FromString("") },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/bolusId", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; bolus id exists",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) {
						datum.BolusID = pointer.FromString(dataTest.RandomID())
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units mmol/L; bolus missing",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.Bolus = nil },
				),
				Entry("units mmol/L; bolus exists",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) {
						datum.Bolus = data.DatumAsPointer(testDataTypesBolusCombination.NewCombination())
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/bolus", NewMeta()),
				),
				Entry("units mmol/L; bolus id missing",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.BolusID = nil },
				),
				Entry("units mmol/L; bolus id empty",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) { datum.BolusID = pointer.FromString("") },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/bolusId", NewMeta()),
				),
				Entry("units mmol/L; bolus id exists",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) {
						datum.BolusID = pointer.FromString(dataTest.RandomID())
					},
				),
				Entry("units mmol/l; bolus missing",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) { datum.Bolus = nil },
				),
				Entry("units mmol/l; bolus exists",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) {
						datum.Bolus = data.DatumAsPointer(testDataTypesBolusCombination.NewCombination())
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/bolus", NewMeta()),
				),
				Entry("units mmol/l; bolus id missing",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) { datum.BolusID = nil },
				),
				Entry("units mmol/l; bolus id empty",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) { datum.BolusID = pointer.FromString("") },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/bolusId", NewMeta()),
				),
				Entry("units mmol/l; bolus id exists",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) {
						datum.BolusID = pointer.FromString(dataTest.RandomID())
					},
				),
				Entry("units mg/dL; bolus missing",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) { datum.Bolus = nil },
				),
				Entry("units mg/dL; bolus exists",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) {
						datum.Bolus = data.DatumAsPointer(testDataTypesBolusCombination.NewCombination())
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/bolus", NewMeta()),
				),
				Entry("units mg/dL; bolus id missing",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) { datum.BolusID = nil },
				),
				Entry("units mg/dL; bolus id empty",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) { datum.BolusID = pointer.FromString("") },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/bolusId", NewMeta()),
				),
				Entry("units mg/dL; bolus id exists",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) {
						datum.BolusID = pointer.FromString(dataTest.RandomID())
					},
				),
				Entry("units mg/dl; bolus missing",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) { datum.Bolus = nil },
				),
				Entry("units mg/dl; bolus exists",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) {
						datum.Bolus = data.DatumAsPointer(testDataTypesBolusCombination.NewCombination())
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/bolus", NewMeta()),
				),
				Entry("units mg/dl; bolus id missing",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) { datum.BolusID = nil },
				),
				Entry("units mg/dl; bolus id empty",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) { datum.BolusID = pointer.FromString("") },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/bolusId", NewMeta()),
				),
				Entry("units mg/dl; bolus id exists",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) {
						datum.BolusID = pointer.FromString(dataTest.RandomID())
					},
				),
			)
		})

		Context("Normalize", func() {
			It("does not modify datum if bolus is missing", func() {
				datum := NewCalculatorWithBolusID(pointer.FromString("mmol/L"))
				expectedDatum := CloneCalculator(datum)
				normalizer := dataNormalizer.New()
				Expect(normalizer).ToNot(BeNil())
				datum.Normalize(normalizer)
				Expect(normalizer.Error()).To(BeNil())
				Expect(normalizer.Data()).To(BeEmpty())
				Expect(datum).To(Equal(expectedDatum))
			})

			It("normalizes the datum and replaces combination bolus with bolus id", func() {
				datumBolus := testDataTypesBolusCombination.NewCombination()
				datum := NewCalculatorWithBolusID(pointer.FromString("mmol/L"))
				datum.Bolus = data.DatumAsPointer(datumBolus)
				expectedDatum := CloneCalculator(datum)
				normalizer := dataNormalizer.New()
				Expect(normalizer).ToNot(BeNil())
				datum.Normalize(normalizer)
				Expect(normalizer.Error()).To(BeNil())
				Expect(normalizer.Data()).To(Equal([]data.Datum{datumBolus}))
				expectedDatum.Bolus = nil
				expectedDatum.BolusID = pointer.FromString(*datumBolus.ID)
				Expect(datum).To(Equal(expectedDatum))
			})

			It("normalizes the datum and replaces extended bolus with bolus id", func() {
				datumBolus := testDataTypesBolusExtended.NewExtended()
				datum := NewCalculatorWithBolusID(pointer.FromString("mmol/L"))
				datum.Bolus = data.DatumAsPointer(datumBolus)
				expectedDatum := CloneCalculator(datum)
				normalizer := dataNormalizer.New()
				Expect(normalizer).ToNot(BeNil())
				datum.Normalize(normalizer)
				Expect(normalizer.Error()).To(BeNil())
				Expect(normalizer.Data()).To(Equal([]data.Datum{datumBolus}))
				expectedDatum.Bolus = nil
				expectedDatum.BolusID = pointer.FromString(*datumBolus.ID)
				Expect(datum).To(Equal(expectedDatum))
			})

			It("normalizes the datum and replaces normal bolus with bolus id", func() {
				datumBolus := testDataTypesBolusNormal.NewNormal()
				datum := NewCalculatorWithBolusID(pointer.FromString("mmol/L"))
				datum.Bolus = data.DatumAsPointer(datumBolus)
				expectedDatum := CloneCalculator(datum)
				normalizer := dataNormalizer.New()
				Expect(normalizer).ToNot(BeNil())
				datum.Normalize(normalizer)
				Expect(normalizer.Error()).To(BeNil())
				Expect(normalizer.Data()).To(Equal([]data.Datum{datumBolus}))
				expectedDatum.Bolus = nil
				expectedDatum.BolusID = pointer.FromString(*datumBolus.ID)
				Expect(datum).To(Equal(expectedDatum))
			})

			DescribeTable("normalizes the datum",
				func(units *string, mutator func(datum *calculator.Calculator, units *string), expectator func(datum *calculator.Calculator, expectedDatum *calculator.Calculator, units *string)) {
					for _, origin := range structure.Origins() {
						datum := NewCalculator(units)
						mutator(datum, units)
						expectedDatum := CloneCalculator(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum, units)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *calculator.Calculator, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *calculator.Calculator, units *string) {},
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *calculator.Calculator, units *string), expectator func(datum *calculator.Calculator, expectedDatum *calculator.Calculator, units *string)) {
					datum := NewCalculator(units)
					mutator(datum, units)
					expectedDatum := CloneCalculator(datum)
					normalizer := dataNormalizer.New()
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(structure.OriginExternal))
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					if expectator != nil {
						expectator(datum, expectedDatum, units)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) {},
					func(datum *calculator.Calculator, expectedDatum *calculator.Calculator, units *string) {
						testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					},
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) {},
					func(datum *calculator.Calculator, expectedDatum *calculator.Calculator, units *string) {
						testDataBloodGlucose.ExpectNormalizedValue(datum.BloodGlucoseInput, expectedDatum.BloodGlucoseInput, units)
						testDataBloodGlucose.ExpectNormalizedTarget(datum.BloodGlucoseTarget, expectedDatum.BloodGlucoseTarget, units)
						testDataBloodGlucose.ExpectNormalizedValue(datum.InsulinSensitivity, expectedDatum.InsulinSensitivity, units)
						testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) {},
					func(datum *calculator.Calculator, expectedDatum *calculator.Calculator, units *string) {
						testDataBloodGlucose.ExpectNormalizedValue(datum.BloodGlucoseInput, expectedDatum.BloodGlucoseInput, units)
						testDataBloodGlucose.ExpectNormalizedTarget(datum.BloodGlucoseTarget, expectedDatum.BloodGlucoseTarget, units)
						testDataBloodGlucose.ExpectNormalizedValue(datum.InsulinSensitivity, expectedDatum.InsulinSensitivity, units)
						testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *calculator.Calculator, units *string), expectator func(datum *calculator.Calculator, expectedDatum *calculator.Calculator, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := NewCalculator(units)
						mutator(datum, units)
						expectedDatum := CloneCalculator(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum, units)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *calculator.Calculator, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *calculator.Calculator, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *calculator.Calculator, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *calculator.Calculator, units *string) {},
					nil,
				),
			)
		})
	})
})
