package pump_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

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

func NewPump(unitsBloodGlucose *string) *pump.Pump {
	scheduleName := testDataTypesBasal.NewScheduleName()
	datum := pump.New()
	datum.Base = *testDataTypes.NewBase()
	datum.Type = "pumpSettings"
	datum.ActiveScheduleName = pointer.String(scheduleName)
	datum.BasalSchedules = pump.NewBasalScheduleArrayMap()
	datum.BasalSchedules.Set(scheduleName, NewBasalScheduleArray())
	datum.BloodGlucoseTargets = NewBloodGlucoseTargetArray(unitsBloodGlucose)
	datum.CarbohydrateRatios = NewCarbohydrateRatioArray()
	datum.InsulinSensitivities = NewInsulinSensitivityArray(unitsBloodGlucose)
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
	clone.BasalSchedules = CloneBasalScheduleArrayMap(datum.BasalSchedules)
	clone.BloodGlucoseTargets = CloneBloodGlucoseTargetArray(datum.BloodGlucoseTargets)
	clone.CarbohydrateRatios = CloneCarbohydrateRatioArray(datum.CarbohydrateRatios)
	clone.InsulinSensitivities = CloneInsulinSensitivityArray(datum.InsulinSensitivities)
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
			Expect(datum.BasalSchedules).To(BeNil())
			Expect(datum.BloodGlucoseTargets).To(BeNil())
			Expect(datum.CarbohydrateRatios).To(BeNil())
			Expect(datum.InsulinSensitivities).To(BeNil())
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
						datum.BasalSchedules.Set("one", invalidBasalScheduleArray)
						(*datum.BloodGlucoseTargets)[0].Target = *dataBloodGlucose.NewTarget()
						(*datum.CarbohydrateRatios)[0].Amount = nil
						(*datum.InsulinSensitivities)[0].Amount = nil
						datum.Units = NewUnits(pointer.String("invalid"))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "pumpSettings"), "/type", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/activeSchedule", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/basalSchedules/one/0/rate", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgTarget/0/target", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/carbRatio/0/amount", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/insulinSensitivity/0/amount", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units/bg", &types.Meta{Type: "invalidType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(unitsBloodGlucose *string, mutator func(datum *pump.Pump, unitsBloodGlucose *string), expectator func(datum *pump.Pump, expectedDatum *pump.Pump, unitsBloodGlucose *string)) {
					for _, origin := range structure.Origins() {
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
				Entry("does not modify the datum",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *pump.Pump, unitsBloodGlucose *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.String("invalid"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {},
					nil,
				),
			)

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
				Entry("does not modify the datum; units mmol/L",
					pointer.String("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {},
					nil,
				),
				Entry("modifies the datum; units mmol/l",
					pointer.String("mmol/l"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {},
					func(datum *pump.Pump, expectedDatum *pump.Pump, unitsBloodGlucose *string) {
						testDataBloodGlucose.ExpectNormalizedUnits(datum.Units.BloodGlucose, expectedDatum.Units.BloodGlucose)
					},
				),
				Entry("modifies the datum; units mg/dL",
					pointer.String("mg/dL"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {},
					func(datum *pump.Pump, expectedDatum *pump.Pump, unitsBloodGlucose *string) {
						testDataBloodGlucose.ExpectNormalizedTarget(&(*datum.BloodGlucoseTargets)[0].Target, &(*expectedDatum.BloodGlucoseTargets)[0].Target, unitsBloodGlucose)
						testDataBloodGlucose.ExpectNormalizedValue((*datum.InsulinSensitivities)[0].Amount, (*expectedDatum.InsulinSensitivities)[0].Amount, unitsBloodGlucose)
						testDataBloodGlucose.ExpectNormalizedUnits(datum.Units.BloodGlucose, expectedDatum.Units.BloodGlucose)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.String("mg/dl"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {},
					func(datum *pump.Pump, expectedDatum *pump.Pump, unitsBloodGlucose *string) {
						testDataBloodGlucose.ExpectNormalizedTarget(&(*datum.BloodGlucoseTargets)[0].Target, &(*expectedDatum.BloodGlucoseTargets)[0].Target, unitsBloodGlucose)
						testDataBloodGlucose.ExpectNormalizedValue((*datum.InsulinSensitivities)[0].Amount, (*expectedDatum.InsulinSensitivities)[0].Amount, unitsBloodGlucose)
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
