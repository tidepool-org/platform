package pump_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"math/rand"
	"sort"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	testDataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types"
	testDataTypesBasal "github.com/tidepool-org/platform/data/types/basal/test"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &types.Meta{
		Type: "pumpSettings",
	}
}

func NewManufacturer(minimumLength int, maximumLength int) string {
	return test.NewText(minimumLength, maximumLength)
}

func NewManufacturers(minimumLength int, maximumLength int) []string {
	result := make([]string, minimumLength+rand.Intn(maximumLength-minimumLength+1))
	for index := range result {
		result[index] = NewManufacturer(1, 100)
	}
	return result
}

func NewPump(unitsBloodGlucose *string) *pump.Pump {
	scheduleName := testDataTypesBasal.NewScheduleName()
	datum := pump.New()
	datum.Base = *testDataTypes.NewBase()
	datum.Type = "pumpSettings"
	datum.ActiveScheduleName = pointer.String(scheduleName)
	datum.Basal = NewBasal()
	datum.BasalSchedules = pump.NewBasalScheduleArrayMap()
	datum.BasalSchedules.Set(scheduleName, NewBasalScheduleArray())
	datum.BloodGlucoseTargets = NewBloodGlucoseTargetArray(unitsBloodGlucose)
	datum.Bolus = NewBolus()
	datum.CarbohydrateRatios = NewCarbohydrateRatioArray()
	datum.Display = NewDisplay()
	datum.Insulin = NewInsulin()
	datum.InsulinSensitivities = NewInsulinSensitivityArray(unitsBloodGlucose)
	datum.Manufacturers = pointer.StringArray(NewManufacturers(1, 10))
	datum.Model = pointer.String(test.NewText(1, 100))
	datum.SerialNumber = pointer.String(test.NewText(1, 100))
	datum.Units = NewUnits(unitsBloodGlucose)
	return datum
}

func ClonePump(datum *pump.Pump) *pump.Pump {
	if datum == nil {
		return nil
	}
	clone := pump.New()
	clone.Base = *testDataTypes.CloneBase(&datum.Base)
	clone.ActiveScheduleName = test.CloneString(datum.ActiveScheduleName)
	clone.Basal = CloneBasal(datum.Basal)
	clone.BasalSchedules = CloneBasalScheduleArrayMap(datum.BasalSchedules)
	clone.BloodGlucoseTargets = CloneBloodGlucoseTargetArray(datum.BloodGlucoseTargets)
	clone.Bolus = CloneBolus(datum.Bolus)
	clone.CarbohydrateRatios = CloneCarbohydrateRatioArray(datum.CarbohydrateRatios)
	clone.Display = CloneDisplay(datum.Display)
	clone.Insulin = CloneInsulin(datum.Insulin)
	clone.InsulinSensitivities = CloneInsulinSensitivityArray(datum.InsulinSensitivities)
	clone.Manufacturers = test.CloneStringArray(datum.Manufacturers)
	clone.Model = test.CloneString(datum.Model)
	clone.SerialNumber = test.CloneString(datum.SerialNumber)
	clone.Units = CloneUnits(datum.Units)
	return clone
}

var _ = Describe("Pump", func() {
	It("Type is expected", func() {
		Expect(pump.Type).To(Equal("pumpSettings"))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := pump.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("pumpSettings"))
			Expect(datum.ActiveScheduleName).To(BeNil())
			Expect(datum.Basal).To(BeNil())
			Expect(datum.BasalSchedules).To(BeNil())
			Expect(datum.BloodGlucoseTargets).To(BeNil())
			Expect(datum.Bolus).To(BeNil())
			Expect(datum.CarbohydrateRatios).To(BeNil())
			Expect(datum.Display).To(BeNil())
			Expect(datum.Insulin).To(BeNil())
			Expect(datum.InsulinSensitivities).To(BeNil())
			Expect(datum.Manufacturers).To(BeNil())
			Expect(datum.Model).To(BeNil())
			Expect(datum.SerialNumber).To(BeNil())
			Expect(datum.Units).To(BeNil())
		})
	})

	Context("Pump", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(unitsBloodGlucose *string, mutator func(datum *pump.Pump, unitsBloodGlucose *string), expectedErrors ...error) {
					datum := NewPump(unitsBloodGlucose)
					mutator(datum, unitsBloodGlucose)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {},
				),
				Entry("type missing",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Type = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &types.Meta{}),
				),
				Entry("type invalid",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Type = "invalidType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "pumpSettings"), "/type", &types.Meta{Type: "invalidType"}),
				),
				Entry("type pumpSettings",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Type = "pumpSettings" },
				),
				Entry("active schedule name missing",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.ActiveScheduleName = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/activeSchedule", NewMeta()),
				),
				Entry("active schedule name empty",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.ActiveScheduleName = pointer.String("") },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/activeSchedule", NewMeta()),
				),
				Entry("active schedule name valid",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.ActiveScheduleName = pointer.String(testDataTypesBasal.NewScheduleName())
					},
				),
				Entry("basal missing",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Basal = nil },
				),
				Entry("basal invalid",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Basal.Temporary.Type = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/basal/temporary/type", NewMeta()),
				),
				Entry("basal valid",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Basal = NewBasal() },
				),
				Entry("basal schedules missing",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.BasalSchedules = nil },
				),
				Entry("basal schedules invalid",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						invalidBasalScheduleArray := NewBasalScheduleArray()
						(*invalidBasalScheduleArray)[0].Rate = nil
						datum.BasalSchedules.Set("one", invalidBasalScheduleArray)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/basalSchedules/one/0/rate", NewMeta()),
				),
				Entry("basal schedules valid",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.BasalSchedules.Set("one", NewBasalScheduleArray())
					},
				),
				Entry("blood glucose targets missing",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.BloodGlucoseTargets = nil },
				),
				Entry("blood glucose targets invalid",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						(*datum.BloodGlucoseTargets)[0].Target = *dataBloodGlucose.NewTarget()
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgTarget/0/target", NewMeta()),
				),
				Entry("blood glucose targets valid",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						(*datum.BloodGlucoseTargets)[0] = NewBloodGlucoseTarget(unitsBloodGlucose)
					},
				),
				Entry("bolus missing",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Bolus = nil },
				),
				Entry("bolus invalid",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Bolus.Combination.Enabled = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/combination/enabled", NewMeta()),
				),
				Entry("bolus valid",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Bolus = NewBolus() },
				),
				Entry("carbohydrate ratios missing",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.CarbohydrateRatios = nil },
				),
				Entry("carbohydrate ratios invalid",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { (*datum.CarbohydrateRatios)[0].Amount = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/carbRatio/0/amount", NewMeta()),
				),
				Entry("carbohydrate ratios valid",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						(*datum.CarbohydrateRatios)[0] = NewCarbohydrateRatio()
					},
				),
				Entry("display missing",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Display = nil },
				),
				Entry("display invalid",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Display.Units = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/display/units", NewMeta()),
				),
				Entry("display valid",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Display = NewDisplay() },
				),
				Entry("insulin missing",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Insulin = nil },
				),
				Entry("insulin invalid",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Insulin.Units = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/insulin/units", NewMeta()),
				),
				Entry("insulin valid",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Insulin = NewInsulin() },
				),
				Entry("insulin sensitivities missing",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.InsulinSensitivities = nil },
				),
				Entry("insulin sensitivities invalid",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { (*datum.InsulinSensitivities)[0].Amount = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/insulinSensitivity/0/amount", NewMeta()),
				),
				Entry("insulin sensitivities valid",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						(*datum.InsulinSensitivities)[0] = NewInsulinSensitivity(unitsBloodGlucose)
					},
				),
				Entry("manufacturers missing",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Manufacturers = nil },
				),
				Entry("manufacturers empty",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.Manufacturers = pointer.StringArray([]string{})
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/manufacturers", NewMeta()),
				),
				Entry("manufacturers length; in range (upper)",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.Manufacturers = pointer.StringArray(NewManufacturers(10, 10))
					},
				),
				Entry("manufacturers length; out of range (upper)",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.Manufacturers = pointer.StringArray(NewManufacturers(11, 11))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(11, 10), "/manufacturers", NewMeta()),
				),
				Entry("manufacturers manufacturer empty",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.Manufacturers = pointer.StringArray(append([]string{NewManufacturer(1, 100), "", NewManufacturer(1, 100), ""}, NewManufacturers(0, 6)...))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/manufacturers/1", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/manufacturers/3", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueDuplicate(), "/manufacturers/3", NewMeta()),
				),
				Entry("manufacturers manufacturer length; in range (upper)",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.Manufacturers = pointer.StringArray(append([]string{NewManufacturer(100, 100), NewManufacturer(1, 100), NewManufacturer(100, 100)}, NewManufacturers(0, 7)...))
					},
				),
				Entry("manufacturers manufacturer length; out of range (upper)",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.Manufacturers = pointer.StringArray(append([]string{NewManufacturer(101, 101), NewManufacturer(1, 100), NewManufacturer(101, 101)}, NewManufacturers(0, 7)...))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/manufacturers/0", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/manufacturers/2", NewMeta()),
				),
				Entry("model missing",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Model = nil },
				),
				Entry("model empty",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Model = pointer.String("") },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/model", NewMeta()),
				),
				Entry("model length in range (upper)",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Model = pointer.String(test.NewText(1, 100)) },
				),
				Entry("model length out of range (upper)",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.Model = pointer.String(test.NewText(101, 101))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/model", NewMeta()),
				),
				Entry("serial number missing",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.SerialNumber = nil },
				),
				Entry("serial number empty",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.SerialNumber = pointer.String("") },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/serialNumber", NewMeta()),
				),
				Entry("serial number length in range (upper)",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.SerialNumber = pointer.String(test.NewText(1, 100))
					},
				),
				Entry("serial number length out of range (upper)",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.SerialNumber = pointer.String(test.NewText(101, 101))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/serialNumber", NewMeta()),
				),
				Entry("units missing",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Units = nil },
				),
				Entry("units invalid",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.Units.BloodGlucose = pointer.String("invalid")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units/bg", NewMeta()),
				),
				Entry("units valid",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Units = NewUnits(unitsBloodGlucose) },
				),
				Entry("multiple errors",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						invalidBasalScheduleArray := NewBasalScheduleArray()
						(*invalidBasalScheduleArray)[0].Rate = nil
						datum.Type = "invalidType"
						datum.ActiveScheduleName = pointer.String("")
						datum.Basal.Temporary.Type = nil
						datum.BasalSchedules.Set("one", invalidBasalScheduleArray)
						(*datum.BloodGlucoseTargets)[0].Target = *dataBloodGlucose.NewTarget()
						datum.Bolus.Combination.Enabled = nil
						(*datum.CarbohydrateRatios)[0].Amount = nil
						datum.Display.Units = nil
						datum.Insulin.Units = nil
						(*datum.InsulinSensitivities)[0].Amount = nil
						datum.Manufacturers = pointer.StringArray([]string{})
						datum.Model = pointer.String("")
						datum.SerialNumber = pointer.String("")
						datum.Units = NewUnits(pointer.String("invalid"))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "pumpSettings"), "/type", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/activeSchedule", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/basal/temporary/type", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/basalSchedules/one/0/rate", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgTarget/0/target", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/combination/enabled", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/carbRatio/0/amount", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/display/units", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/insulin/units", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/insulinSensitivity/0/amount", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/manufacturers", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/model", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/serialNumber", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units/bg", &types.Meta{Type: "invalidType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum with origin external",
				func(unitsBloodGlucose *string, mutator func(datum *pump.Pump, unitsBloodGlucose *string), expectator func(datum *pump.Pump, expectedDatum *pump.Pump, unitsBloodGlucose *string)) {
					datum := NewPump(unitsBloodGlucose)
					mutator(datum, unitsBloodGlucose)
					expectedDatum := ClonePump(datum)
					normalizer := dataNormalizer.New()
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(structure.OriginExternal))
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					if expectator != nil {
						expectator(datum, expectedDatum, unitsBloodGlucose)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("modifies the datum",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {},
					func(datum *pump.Pump, expectedDatum *pump.Pump, unitsBloodGlucose *string) {
						sort.Strings(*expectedDatum.Manufacturers)
					},
				),
				Entry("modifies the datum; units missing",
					nil,
					func(datum *pump.Pump, unitsBloodGlucose *string) {},
					func(datum *pump.Pump, expectedDatum *pump.Pump, unitsBloodGlucose *string) {
						sort.Strings(*expectedDatum.Manufacturers)
					},
				),
				Entry("modifies the datum; units invalid",
					pointer.String("invalid"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {},
					func(datum *pump.Pump, expectedDatum *pump.Pump, unitsBloodGlucose *string) {
						sort.Strings(*expectedDatum.Manufacturers)
					},
				),
				Entry("modifies the datum; units mmol/L",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {},
					func(datum *pump.Pump, expectedDatum *pump.Pump, unitsBloodGlucose *string) {
						sort.Strings(*expectedDatum.Manufacturers)
					},
				),
				Entry("modifies the datum; units mmol/l",
					pointer.String("mmol/l"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {},
					func(datum *pump.Pump, expectedDatum *pump.Pump, unitsBloodGlucose *string) {
						sort.Strings(*expectedDatum.Manufacturers)
						testDataBloodGlucose.ExpectNormalizedUnits(datum.Units.BloodGlucose, expectedDatum.Units.BloodGlucose)
					},
				),
				Entry("modifies the datum; units mg/dL",
					pointer.String("mg/dL"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {},
					func(datum *pump.Pump, expectedDatum *pump.Pump, unitsBloodGlucose *string) {
						testDataBloodGlucose.ExpectNormalizedTarget(&(*datum.BloodGlucoseTargets)[0].Target, &(*expectedDatum.BloodGlucoseTargets)[0].Target, unitsBloodGlucose)
						testDataBloodGlucose.ExpectNormalizedValue((*datum.InsulinSensitivities)[0].Amount, (*expectedDatum.InsulinSensitivities)[0].Amount, unitsBloodGlucose)
						sort.Strings(*expectedDatum.Manufacturers)
						testDataBloodGlucose.ExpectNormalizedUnits(datum.Units.BloodGlucose, expectedDatum.Units.BloodGlucose)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.String("mg/dl"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {},
					func(datum *pump.Pump, expectedDatum *pump.Pump, unitsBloodGlucose *string) {
						testDataBloodGlucose.ExpectNormalizedTarget(&(*datum.BloodGlucoseTargets)[0].Target, &(*expectedDatum.BloodGlucoseTargets)[0].Target, unitsBloodGlucose)
						testDataBloodGlucose.ExpectNormalizedValue((*datum.InsulinSensitivities)[0].Amount, (*expectedDatum.InsulinSensitivities)[0].Amount, unitsBloodGlucose)
						sort.Strings(*expectedDatum.Manufacturers)
						testDataBloodGlucose.ExpectNormalizedUnits(datum.Units.BloodGlucose, expectedDatum.Units.BloodGlucose)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(unitsBloodGlucose *string, mutator func(datum *pump.Pump, unitsBloodGlucose *string), expectator func(datum *pump.Pump, expectedDatum *pump.Pump, unitsBloodGlucose *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := NewPump(unitsBloodGlucose)
						mutator(datum, unitsBloodGlucose)
						expectedDatum := ClonePump(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum, unitsBloodGlucose)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum; units mmol/L",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.String("mmol/l"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dL",
					pointer.String("mg/dL"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dl",
					pointer.String("mg/dl"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {},
					nil,
				),
			)
		})
	})
})
