package pump_test

import (
	"math/rand"
	"sort"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types"
	dataTypesBasalTest "github.com/tidepool-org/platform/data/types/basal/test"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
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
	return test.RandomStringFromRange(minimumLength, maximumLength)
}

func NewManufacturers(minimumLength int, maximumLength int) []string {
	result := make([]string, minimumLength+rand.Intn(maximumLength-minimumLength+1))
	for index := range result {
		result[index] = NewManufacturer(1, 100)
	}
	return result
}

func NewPump(unitsBloodGlucose *string) *pump.Pump {
	scheduleName := dataTypesBasalTest.NewScheduleName()
	datum := pump.New()
	datum.Base = *dataTypesTest.NewBase()
	datum.Type = "pumpSettings"
	datum.ActiveScheduleName = pointer.FromString(scheduleName)
	datum.Basal = NewBasal()
	datum.BasalRateSchedules = pump.NewBasalRateStartArrayMap()
	datum.BasalRateSchedules.Set(scheduleName, NewBasalRateStartArray())
	datum.BloodGlucoseTargetSchedules = pump.NewBloodGlucoseTargetStartArrayMap()
	datum.BloodGlucoseTargetSchedules.Set(scheduleName, NewBloodGlucoseTargetStartArray(unitsBloodGlucose))
	datum.Bolus = NewBolus()
	datum.CarbohydrateRatioSchedules = pump.NewCarbohydrateRatioStartArrayMap()
	datum.CarbohydrateRatioSchedules.Set(scheduleName, NewCarbohydrateRatioStartArray())
	datum.Display = NewDisplay()
	datum.InsulinSensitivitySchedules = pump.NewInsulinSensitivityStartArrayMap()
	datum.InsulinSensitivitySchedules.Set(scheduleName, NewInsulinSensitivityStartArray(unitsBloodGlucose))
	datum.Manufacturers = pointer.FromStringArray(NewManufacturers(1, 10))
	datum.Model = pointer.FromString(test.RandomStringFromRange(1, 100))
	datum.SerialNumber = pointer.FromString(test.RandomStringFromRange(1, 100))
	datum.Units = NewUnits(unitsBloodGlucose)
	return datum
}

func ClonePump(datum *pump.Pump) *pump.Pump {
	if datum == nil {
		return nil
	}
	clone := pump.New()
	clone.Base = *dataTypesTest.CloneBase(&datum.Base)
	clone.ActiveScheduleName = pointer.CloneString(datum.ActiveScheduleName)
	clone.Basal = CloneBasal(datum.Basal)
	clone.BasalRateSchedule = CloneBasalRateStartArray(datum.BasalRateSchedule)
	clone.BasalRateSchedules = CloneBasalRateStartArrayMap(datum.BasalRateSchedules)
	clone.BloodGlucoseTargetSchedule = CloneBloodGlucoseTargetStartArray(datum.BloodGlucoseTargetSchedule)
	clone.BloodGlucoseTargetSchedules = CloneBloodGlucoseTargetStartArrayMap(datum.BloodGlucoseTargetSchedules)
	clone.Bolus = CloneBolus(datum.Bolus)
	clone.CarbohydrateRatioSchedule = CloneCarbohydrateRatioStartArray(datum.CarbohydrateRatioSchedule)
	clone.CarbohydrateRatioSchedules = CloneCarbohydrateRatioStartArrayMap(datum.CarbohydrateRatioSchedules)
	clone.Display = CloneDisplay(datum.Display)
	clone.InsulinSensitivitySchedule = CloneInsulinSensitivityStartArray(datum.InsulinSensitivitySchedule)
	clone.InsulinSensitivitySchedules = CloneInsulinSensitivityStartArrayMap(datum.InsulinSensitivitySchedules)
	clone.Manufacturers = pointer.CloneStringArray(datum.Manufacturers)
	clone.Model = pointer.CloneString(datum.Model)
	clone.SerialNumber = pointer.CloneString(datum.SerialNumber)
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
			Expect(datum.BasalRateSchedule).To(BeNil())
			Expect(datum.BasalRateSchedules).To(BeNil())
			Expect(datum.BloodGlucoseTargetSchedule).To(BeNil())
			Expect(datum.BloodGlucoseTargetSchedules).To(BeNil())
			Expect(datum.Bolus).To(BeNil())
			Expect(datum.CarbohydrateRatioSchedule).To(BeNil())
			Expect(datum.CarbohydrateRatioSchedules).To(BeNil())
			Expect(datum.Display).To(BeNil())
			Expect(datum.InsulinSensitivitySchedule).To(BeNil())
			Expect(datum.InsulinSensitivitySchedules).To(BeNil())
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
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {},
				),
				Entry("type missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &types.Meta{}),
				),
				Entry("type invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "pumpSettings"), "/type", &types.Meta{Type: "invalidType"}),
				),
				Entry("type pumpSettings",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Type = "pumpSettings" },
				),
				Entry("active schedule name missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.ActiveScheduleName = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/activeSchedule", NewMeta()),
				),
				Entry("active schedule name empty",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.ActiveScheduleName = pointer.FromString("") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/activeSchedule", NewMeta()),
				),
				Entry("active schedule name valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.ActiveScheduleName = pointer.FromString(dataTypesBasalTest.NewScheduleName())
					},
				),
				Entry("basal missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Basal = nil },
				),
				Entry("basal invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Basal.Temporary.Type = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/basal/temporary/type", NewMeta()),
				),
				Entry("basal valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Basal = NewBasal() },
				),
				Entry("basal rate schedule and basal rate schedules missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.BasalRateSchedule = nil
						datum.BasalRateSchedules = nil
					},
				),
				Entry("basal rate schedule invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						invalidBasalRateSchedule := NewBasalRateStartArray()
						(*invalidBasalRateSchedule)[0].Start = nil
						datum.BasalRateSchedule = invalidBasalRateSchedule
						datum.BasalRateSchedules = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/basalSchedule/0/start", NewMeta()),
				),
				Entry("basal rate schedule valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.BasalRateSchedule = NewBasalRateStartArray()
						datum.BasalRateSchedules = nil
					},
				),
				Entry("basal rate schedules invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						invalidBasalRateSchedule := NewBasalRateStartArray()
						(*invalidBasalRateSchedule)[0].Start = nil
						datum.BasalRateSchedules.Set("one", invalidBasalRateSchedule)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/basalSchedules/one/0/start", NewMeta()),
				),
				Entry("basal rate schedules valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.BasalRateSchedules.Set("one", NewBasalRateStartArray())
					},
				),
				Entry("blood glucose target schedule and blood glucose target schedules missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.BloodGlucoseTargetSchedule = nil
						datum.BloodGlucoseTargetSchedules = nil
					},
				),
				Entry("blood glucose target schedule invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						invalidBloodGlucoseTargetSchedule := NewBloodGlucoseTargetStartArray(unitsBloodGlucose)
						(*invalidBloodGlucoseTargetSchedule)[0].Start = nil
						datum.BloodGlucoseTargetSchedule = invalidBloodGlucoseTargetSchedule
						datum.BloodGlucoseTargetSchedules = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgTarget/0/start", NewMeta()),
				),
				Entry("blood glucose target schedule valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.BloodGlucoseTargetSchedule = NewBloodGlucoseTargetStartArray(unitsBloodGlucose)
						datum.BloodGlucoseTargetSchedules = nil
					},
				),
				Entry("blood glucose target schedules invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						invalidBloodGlucoseTargetSchedule := NewBloodGlucoseTargetStartArray(unitsBloodGlucose)
						(*invalidBloodGlucoseTargetSchedule)[0].Start = nil
						datum.BloodGlucoseTargetSchedules.Set("one", invalidBloodGlucoseTargetSchedule)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgTargets/one/0/start", NewMeta()),
				),
				Entry("blood glucose target schedules valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.BloodGlucoseTargetSchedules.Set("one", NewBloodGlucoseTargetStartArray(unitsBloodGlucose))
					},
				),
				Entry("bolus missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Bolus = nil },
				),
				Entry("bolus invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Bolus.Extended.Enabled = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/extended/enabled", NewMeta()),
				),
				Entry("bolus valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Bolus = NewBolus() },
				),
				Entry("carbohydrate ratio schedule and carbohydrate ratio schedules missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.CarbohydrateRatioSchedule = nil
						datum.CarbohydrateRatioSchedules = nil
					},
				),
				Entry("carbohydrate ratio schedule invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						invalidCarbohydrateRatioSchedule := NewCarbohydrateRatioStartArray()
						(*invalidCarbohydrateRatioSchedule)[0].Start = nil
						datum.CarbohydrateRatioSchedule = invalidCarbohydrateRatioSchedule
						datum.CarbohydrateRatioSchedules = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/carbRatio/0/start", NewMeta()),
				),
				Entry("carbohydrate ratio schedule valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.CarbohydrateRatioSchedule = NewCarbohydrateRatioStartArray()
						datum.CarbohydrateRatioSchedules = nil
					},
				),
				Entry("carbohydrate ratio schedules invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						invalidCarbohydrateRatioSchedule := NewCarbohydrateRatioStartArray()
						(*invalidCarbohydrateRatioSchedule)[0].Start = nil
						datum.CarbohydrateRatioSchedules.Set("one", invalidCarbohydrateRatioSchedule)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/carbRatios/one/0/start", NewMeta()),
				),
				Entry("carbohydrate ratio schedules valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.CarbohydrateRatioSchedules.Set("one", NewCarbohydrateRatioStartArray())
					},
				),
				Entry("display missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Display = nil },
				),
				Entry("display invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Display.BloodGlucose.Units = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/display/bloodGlucose/units", NewMeta()),
				),
				Entry("display valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Display = NewDisplay() },
				),
				Entry("insulin sensitivity schedule and insulin sensitivity schedules missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.InsulinSensitivitySchedule = nil
						datum.InsulinSensitivitySchedules = nil
					},
				),
				Entry("insulin sensitivity schedule invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						invalidInsulinSensitivitySchedule := NewInsulinSensitivityStartArray(unitsBloodGlucose)
						(*invalidInsulinSensitivitySchedule)[0].Start = nil
						datum.InsulinSensitivitySchedule = invalidInsulinSensitivitySchedule
						datum.InsulinSensitivitySchedules = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/insulinSensitivity/0/start", NewMeta()),
				),
				Entry("insulin sensitivity schedule valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.InsulinSensitivitySchedule = NewInsulinSensitivityStartArray(unitsBloodGlucose)
						datum.InsulinSensitivitySchedules = nil
					},
				),
				Entry("insulin sensitivity schedules invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						invalidInsulinSensitivitySchedule := NewInsulinSensitivityStartArray(unitsBloodGlucose)
						(*invalidInsulinSensitivitySchedule)[0].Start = nil
						datum.InsulinSensitivitySchedules.Set("one", invalidInsulinSensitivitySchedule)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/insulinSensitivities/one/0/start", NewMeta()),
				),
				Entry("insulin sensitivity schedules valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.InsulinSensitivitySchedules.Set("one", NewInsulinSensitivityStartArray(unitsBloodGlucose))
					},
				),
				Entry("manufacturers missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Manufacturers = nil },
				),
				Entry("manufacturers empty",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.Manufacturers = pointer.FromStringArray([]string{})
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/manufacturers", NewMeta()),
				),
				Entry("manufacturers length; in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.Manufacturers = pointer.FromStringArray(NewManufacturers(10, 10))
					},
				),
				Entry("manufacturers length; out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.Manufacturers = pointer.FromStringArray(NewManufacturers(11, 11))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(11, 10), "/manufacturers", NewMeta()),
				),
				Entry("manufacturers manufacturer empty",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.Manufacturers = pointer.FromStringArray(append([]string{NewManufacturer(1, 100), "", NewManufacturer(1, 100), ""}, NewManufacturers(0, 6)...))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/manufacturers/1", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/manufacturers/3", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueDuplicate(), "/manufacturers/3", NewMeta()),
				),
				Entry("manufacturers manufacturer length; in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.Manufacturers = pointer.FromStringArray(append([]string{NewManufacturer(100, 100), NewManufacturer(1, 100), NewManufacturer(100, 100)}, NewManufacturers(0, 7)...))
					},
				),
				Entry("manufacturers manufacturer length; out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.Manufacturers = pointer.FromStringArray(append([]string{NewManufacturer(101, 101), NewManufacturer(1, 100), NewManufacturer(101, 101)}, NewManufacturers(0, 7)...))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/manufacturers/0", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/manufacturers/2", NewMeta()),
				),
				Entry("model missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Model = nil },
				),
				Entry("model empty",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Model = pointer.FromString("") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/model", NewMeta()),
				),
				Entry("model length in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.Model = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
				),
				Entry("model length out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.Model = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/model", NewMeta()),
				),
				Entry("serial number missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.SerialNumber = nil },
				),
				Entry("serial number empty",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.SerialNumber = pointer.FromString("") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/serialNumber", NewMeta()),
				),
				Entry("serial number length in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.SerialNumber = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
				),
				Entry("serial number length out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.SerialNumber = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/serialNumber", NewMeta()),
				),
				Entry("units missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Units = nil },
				),
				Entry("units invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.Units.BloodGlucose = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units/bg", NewMeta()),
				),
				Entry("units valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Units = NewUnits(unitsBloodGlucose) },
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.Type = "invalidType"
						datum.ActiveScheduleName = pointer.FromString("")
						datum.Basal.Temporary.Type = nil
						datum.BasalRateSchedules = nil
						datum.BloodGlucoseTargetSchedules = nil
						datum.Bolus.Extended.Enabled = nil
						datum.CarbohydrateRatioSchedules = nil
						datum.Display.BloodGlucose.Units = nil
						datum.InsulinSensitivitySchedules = nil
						datum.Manufacturers = pointer.FromStringArray([]string{})
						datum.Model = pointer.FromString("")
						datum.SerialNumber = pointer.FromString("")
						datum.Units = NewUnits(pointer.FromString("invalid"))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "pumpSettings"), "/type", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/activeSchedule", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/basal/temporary/type", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/extended/enabled", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/display/bloodGlucose/units", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/manufacturers", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/model", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/serialNumber", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units/bg", &types.Meta{Type: "invalidType"}),
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
					pointer.FromString("mmol/L"),
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
					pointer.FromString("invalid"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {},
					func(datum *pump.Pump, expectedDatum *pump.Pump, unitsBloodGlucose *string) {
						sort.Strings(*expectedDatum.Manufacturers)
					},
				),
				Entry("modifies the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {},
					func(datum *pump.Pump, expectedDatum *pump.Pump, unitsBloodGlucose *string) {
						sort.Strings(*expectedDatum.Manufacturers)
					},
				),
				Entry("modifies the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {},
					func(datum *pump.Pump, expectedDatum *pump.Pump, unitsBloodGlucose *string) {
						sort.Strings(*expectedDatum.Manufacturers)
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units.BloodGlucose, expectedDatum.Units.BloodGlucose)
					},
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.BasalRateSchedule = NewBasalRateStartArray()
						datum.BloodGlucoseTargetSchedule = NewBloodGlucoseTargetStartArray(unitsBloodGlucose)
						datum.CarbohydrateRatioSchedule = NewCarbohydrateRatioStartArray()
						datum.InsulinSensitivitySchedule = NewInsulinSensitivityStartArray(unitsBloodGlucose)
					},
					func(datum *pump.Pump, expectedDatum *pump.Pump, unitsBloodGlucose *string) {
						for index := range *datum.BloodGlucoseTargetSchedule {
							dataBloodGlucoseTest.ExpectNormalizedTarget(&(*datum.BloodGlucoseTargetSchedule)[index].Target, &(*expectedDatum.BloodGlucoseTargetSchedule)[index].Target, unitsBloodGlucose)
						}
						for name := range *datum.BloodGlucoseTargetSchedules {
							for index := range *(*datum.BloodGlucoseTargetSchedules)[name] {
								dataBloodGlucoseTest.ExpectNormalizedTarget(&(*(*datum.BloodGlucoseTargetSchedules)[name])[index].Target, &(*(*expectedDatum.BloodGlucoseTargetSchedules)[name])[index].Target, unitsBloodGlucose)
							}
						}
						for index := range *datum.InsulinSensitivitySchedule {
							dataBloodGlucoseTest.ExpectNormalizedValue((*datum.InsulinSensitivitySchedule)[index].Amount, (*expectedDatum.InsulinSensitivitySchedule)[index].Amount, unitsBloodGlucose)
						}
						for name := range *datum.InsulinSensitivitySchedules {
							for index := range *(*datum.InsulinSensitivitySchedules)[name] {
								dataBloodGlucoseTest.ExpectNormalizedValue((*(*datum.InsulinSensitivitySchedules)[name])[index].Amount, (*(*expectedDatum.InsulinSensitivitySchedules)[name])[index].Amount, unitsBloodGlucose)
							}
						}
						sort.Strings(*expectedDatum.Manufacturers)
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units.BloodGlucose, expectedDatum.Units.BloodGlucose)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.BasalRateSchedule = NewBasalRateStartArray()
						datum.BloodGlucoseTargetSchedule = NewBloodGlucoseTargetStartArray(unitsBloodGlucose)
						datum.CarbohydrateRatioSchedule = NewCarbohydrateRatioStartArray()
						datum.InsulinSensitivitySchedule = NewInsulinSensitivityStartArray(unitsBloodGlucose)
					},
					func(datum *pump.Pump, expectedDatum *pump.Pump, unitsBloodGlucose *string) {
						for index := range *datum.BloodGlucoseTargetSchedule {
							dataBloodGlucoseTest.ExpectNormalizedTarget(&(*datum.BloodGlucoseTargetSchedule)[index].Target, &(*expectedDatum.BloodGlucoseTargetSchedule)[index].Target, unitsBloodGlucose)
						}
						for name := range *datum.BloodGlucoseTargetSchedules {
							for index := range *(*datum.BloodGlucoseTargetSchedules)[name] {
								dataBloodGlucoseTest.ExpectNormalizedTarget(&(*(*datum.BloodGlucoseTargetSchedules)[name])[index].Target, &(*(*expectedDatum.BloodGlucoseTargetSchedules)[name])[index].Target, unitsBloodGlucose)
							}
						}
						for index := range *datum.InsulinSensitivitySchedule {
							dataBloodGlucoseTest.ExpectNormalizedValue((*datum.InsulinSensitivitySchedule)[index].Amount, (*expectedDatum.InsulinSensitivitySchedule)[index].Amount, unitsBloodGlucose)
						}
						for name := range *datum.InsulinSensitivitySchedules {
							for index := range *(*datum.InsulinSensitivitySchedules)[name] {
								dataBloodGlucoseTest.ExpectNormalizedValue((*(*datum.InsulinSensitivitySchedules)[name])[index].Amount, (*(*expectedDatum.InsulinSensitivitySchedules)[name])[index].Amount, unitsBloodGlucose)
							}
						}
						sort.Strings(*expectedDatum.Manufacturers)
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units.BloodGlucose, expectedDatum.Units.BloodGlucose)
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
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {},
					nil,
				),
			)
		})
	})
})
