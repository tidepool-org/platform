package pump_test

import (
	"sort"

	pumpTest "github.com/tidepool-org/platform/data/types/settings/pump/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
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
			Expect(datum.AutomatedDelivery).To(BeNil())
			Expect(datum.Basal).To(BeNil())
			Expect(datum.BasalRateSchedule).To(BeNil())
			Expect(datum.BasalRateSchedules).To(BeNil())
			Expect(datum.BloodGlucoseSuspendThreshold).To(BeNil())
			Expect(datum.BloodGlucoseTargetPhysicalActivity).To(BeNil())
			Expect(datum.BloodGlucoseTargetPreprandial).To(BeNil())
			Expect(datum.BloodGlucoseTargetSchedule).To(BeNil())
			Expect(datum.BloodGlucoseTargetSchedules).To(BeNil())
			Expect(datum.Bolus).To(BeNil())
			Expect(datum.CarbohydrateRatioSchedule).To(BeNil())
			Expect(datum.CarbohydrateRatioSchedules).To(BeNil())
			Expect(datum.Display).To(BeNil())
			Expect(datum.InsulinModel).To(BeNil())
			Expect(datum.InsulinSensitivitySchedule).To(BeNil())
			Expect(datum.InsulinSensitivitySchedules).To(BeNil())
			Expect(datum.Manufacturers).To(BeNil())
			Expect(datum.Model).To(BeNil())
			Expect(datum.ScheduleTimeZoneOffset).To(BeNil())
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
					datum := pumpTest.NewPump(unitsBloodGlucose)
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
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/activeSchedule", pumpTest.NewMeta()),
				),
				Entry("active schedule name empty",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.ActiveScheduleName = pointer.FromString("") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/activeSchedule", pumpTest.NewMeta()),
				),
				Entry("active schedule name valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.ActiveScheduleName = pointer.FromString(dataTypesBasalTest.NewScheduleName())
					},
				),
				Entry("automated delivery missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.AutomatedDelivery = nil },
				),
				Entry("automated delivery false",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.AutomatedDelivery = pointer.FromBool(false) },
				),
				Entry("automated delivery true",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.AutomatedDelivery = pointer.FromBool(true) },
				),
				Entry("basal missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Basal = nil },
				),
				Entry("basal invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Basal.Temporary.Type = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/basal/temporary/type", pumpTest.NewMeta()),
				),
				Entry("basal valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Basal = pumpTest.NewBasal() },
				),
				Entry("basal rate schedule and basal rate schedules missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.BasalRateSchedule = nil
						datum.BasalRateSchedules = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/basalSchedule", pumpTest.NewMeta()),
				),
				Entry("basal rate schedule invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						invalidBasalRateSchedule := pumpTest.NewBasalRateStartArray()
						(*invalidBasalRateSchedule)[0].Start = nil
						datum.BasalRateSchedule = invalidBasalRateSchedule
						datum.BasalRateSchedules = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/basalSchedule/0/start", pumpTest.NewMeta()),
				),
				Entry("basal rate schedule valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.BasalRateSchedule = pumpTest.NewBasalRateStartArray()
						datum.BasalRateSchedules = nil
					},
				),
				Entry("basal rate schedules invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						invalidBasalRateSchedule := pumpTest.NewBasalRateStartArray()
						(*invalidBasalRateSchedule)[0].Start = nil
						datum.BasalRateSchedules.Set("one", invalidBasalRateSchedule)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/basalSchedules/one/0/start", pumpTest.NewMeta()),
				),
				Entry("basal rate schedules valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.BasalRateSchedules.Set("one", pumpTest.NewBasalRateStartArray())
					},
				),
				Entry("blood glucose suspend threshold missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.BloodGlucoseSuspendThreshold = nil },
				),
				Entry("blood glucose suspend threshold out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.BloodGlucoseSuspendThreshold = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/bgSuspendThreshold", pumpTest.NewMeta()),
				),
				Entry("blood glucose suspend threshold in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.BloodGlucoseSuspendThreshold = pointer.FromFloat64(0.0)
					},
				),
				Entry("blood glucose suspend threshold in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.BloodGlucoseSuspendThreshold = pointer.FromFloat64(55.0)
					},
				),
				Entry("blood glucose suspend threshold out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.BloodGlucoseSuspendThreshold = pointer.FromFloat64(55.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/bgSuspendThreshold", pumpTest.NewMeta()),
				),
				Entry("blood glucose target physical activity missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.BloodGlucoseTargetPhysicalActivity = nil },
				),
				Entry("blood glucose target physical activity invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.BloodGlucoseTargetPhysicalActivity = dataBloodGlucose.NewTarget()
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgTargetPhysicalActivity/target", pumpTest.NewMeta()),
				),
				Entry("blood glucose target physical activity valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.BloodGlucoseTargetPhysicalActivity = dataBloodGlucoseTest.NewTarget(unitsBloodGlucose)
					},
				),
				Entry("blood glucose target preprandial missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.BloodGlucoseTargetPreprandial = nil },
				),
				Entry("blood glucose target preprandial invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.BloodGlucoseTargetPreprandial = dataBloodGlucose.NewTarget()
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgTargetPreprandial/target", pumpTest.NewMeta()),
				),
				Entry("blood glucose target preprandial valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.BloodGlucoseTargetPreprandial = dataBloodGlucoseTest.NewTarget(unitsBloodGlucose)
					},
				),
				Entry("blood glucose target schedule and blood glucose target schedules missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.BloodGlucoseTargetSchedule = nil
						datum.BloodGlucoseTargetSchedules = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgTarget", pumpTest.NewMeta()),
				),
				Entry("blood glucose target schedule invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						invalidBloodGlucoseTargetSchedule := pumpTest.RandomBloodGlucoseTargetStartArray(unitsBloodGlucose)
						(*invalidBloodGlucoseTargetSchedule)[0].Start = nil
						datum.BloodGlucoseTargetSchedule = invalidBloodGlucoseTargetSchedule
						datum.BloodGlucoseTargetSchedules = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgTarget/0/start", pumpTest.NewMeta()),
				),
				Entry("blood glucose target schedule valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.BloodGlucoseTargetSchedule = pumpTest.RandomBloodGlucoseTargetStartArray(unitsBloodGlucose)
						datum.BloodGlucoseTargetSchedules = nil
					},
				),
				Entry("blood glucose target schedules invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						invalidBloodGlucoseTargetSchedule := pumpTest.RandomBloodGlucoseTargetStartArray(unitsBloodGlucose)
						(*invalidBloodGlucoseTargetSchedule)[0].Start = nil
						datum.BloodGlucoseTargetSchedules.Set("one", invalidBloodGlucoseTargetSchedule)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgTargets/one/0/start", pumpTest.NewMeta()),
				),
				Entry("blood glucose target schedules valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.BloodGlucoseTargetSchedules.Set("one", pumpTest.RandomBloodGlucoseTargetStartArray(unitsBloodGlucose))
					},
				),
				Entry("bolus missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Bolus = nil },
				),
				Entry("bolus invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Bolus.Extended.Enabled = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/extended/enabled", pumpTest.NewMeta()),
				),
				Entry("bolus valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Bolus = pumpTest.NewBolus() },
				),
				Entry("carbohydrate ratio schedule and carbohydrate ratio schedules missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.CarbohydrateRatioSchedule = nil
						datum.CarbohydrateRatioSchedules = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/carbRatio", pumpTest.NewMeta()),
				),
				Entry("carbohydrate ratio schedule invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						invalidCarbohydrateRatioSchedule := pumpTest.NewCarbohydrateRatioStartArray()
						(*invalidCarbohydrateRatioSchedule)[0].Start = nil
						datum.CarbohydrateRatioSchedule = invalidCarbohydrateRatioSchedule
						datum.CarbohydrateRatioSchedules = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/carbRatio/0/start", pumpTest.NewMeta()),
				),
				Entry("carbohydrate ratio schedule valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.CarbohydrateRatioSchedule = pumpTest.NewCarbohydrateRatioStartArray()
						datum.CarbohydrateRatioSchedules = nil
					},
				),
				Entry("carbohydrate ratio schedules invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						invalidCarbohydrateRatioSchedule := pumpTest.NewCarbohydrateRatioStartArray()
						(*invalidCarbohydrateRatioSchedule)[0].Start = nil
						datum.CarbohydrateRatioSchedules.Set("one", invalidCarbohydrateRatioSchedule)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/carbRatios/one/0/start", pumpTest.NewMeta()),
				),
				Entry("carbohydrate ratio schedules valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.CarbohydrateRatioSchedules.Set("one", pumpTest.NewCarbohydrateRatioStartArray())
					},
				),
				Entry("display missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Display = nil },
				),
				Entry("display invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Display.BloodGlucose.Units = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/display/bloodGlucose/units", pumpTest.NewMeta()),
				),
				Entry("display valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Display = pumpTest.NewDisplay() },
				),
				Entry("insulin model missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.InsulinModel = nil },
				),
				Entry("insulin model invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.InsulinModel.ModelType = pointer.FromString("invalid")
						datum.InsulinModel.ModelTypeOther = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"fiasp", "other", "rapidAdult", "rapidChild", "walsh"}), "/insulinModel/modelType", pumpTest.NewMeta()),
				),
				Entry("insulin model valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.InsulinModel = pumpTest.RandomInsulinModel()
					},
				),
				Entry("insulin sensitivity schedule and insulin sensitivity schedules missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.InsulinSensitivitySchedule = nil
						datum.InsulinSensitivitySchedules = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/insulinSensitivity", pumpTest.NewMeta()),
				),
				Entry("insulin sensitivity schedule invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						invalidInsulinSensitivitySchedule := pumpTest.NewInsulinSensitivityStartArray(unitsBloodGlucose)
						(*invalidInsulinSensitivitySchedule)[0].Start = nil
						datum.InsulinSensitivitySchedule = invalidInsulinSensitivitySchedule
						datum.InsulinSensitivitySchedules = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/insulinSensitivity/0/start", pumpTest.NewMeta()),
				),
				Entry("insulin sensitivity schedule valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.InsulinSensitivitySchedule = pumpTest.NewInsulinSensitivityStartArray(unitsBloodGlucose)
						datum.InsulinSensitivitySchedules = nil
					},
				),
				Entry("insulin sensitivity schedules invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						invalidInsulinSensitivitySchedule := pumpTest.NewInsulinSensitivityStartArray(unitsBloodGlucose)
						(*invalidInsulinSensitivitySchedule)[0].Start = nil
						datum.InsulinSensitivitySchedules.Set("one", invalidInsulinSensitivitySchedule)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/insulinSensitivities/one/0/start", pumpTest.NewMeta()),
				),
				Entry("insulin sensitivity schedules valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.InsulinSensitivitySchedules.Set("one", pumpTest.NewInsulinSensitivityStartArray(unitsBloodGlucose))
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
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/manufacturers", pumpTest.NewMeta()),
				),
				Entry("manufacturers length; in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.Manufacturers = pointer.FromStringArray(pumpTest.NewManufacturers(10, 10))
					},
				),
				Entry("manufacturers length; out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.Manufacturers = pointer.FromStringArray(pumpTest.NewManufacturers(11, 11))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(11, 10), "/manufacturers", pumpTest.NewMeta()),
				),
				Entry("manufacturers manufacturer empty",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.Manufacturers = pointer.FromStringArray(append([]string{pumpTest.NewManufacturer(1, 100), "", pumpTest.NewManufacturer(1, 100), ""}, pumpTest.NewManufacturers(0, 6)...))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/manufacturers/1", pumpTest.NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/manufacturers/3", pumpTest.NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueDuplicate(), "/manufacturers/3", pumpTest.NewMeta()),
				),
				Entry("manufacturers manufacturer length; in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.Manufacturers = pointer.FromStringArray(append([]string{pumpTest.NewManufacturer(100, 100), pumpTest.NewManufacturer(1, 100), pumpTest.NewManufacturer(100, 100)}, pumpTest.NewManufacturers(0, 7)...))
					},
				),
				Entry("manufacturers manufacturer length; out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.Manufacturers = pointer.FromStringArray(append([]string{pumpTest.NewManufacturer(101, 101), pumpTest.NewManufacturer(1, 100), pumpTest.NewManufacturer(101, 101)}, pumpTest.NewManufacturers(0, 7)...))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/manufacturers/0", pumpTest.NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/manufacturers/2", pumpTest.NewMeta()),
				),
				Entry("model missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Model = nil },
				),
				Entry("model empty",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.Model = pointer.FromString("") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/model", pumpTest.NewMeta()),
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
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/model", pumpTest.NewMeta()),
				),
				Entry("schedules time zone offset missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.ScheduleTimeZoneOffset = nil },
				),
				Entry("schedules time zone offset out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.ScheduleTimeZoneOffset = pointer.FromInt(pump.ScheduleTimeZoneOffsetMinimum - 1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(pump.ScheduleTimeZoneOffsetMinimum-1, pump.ScheduleTimeZoneOffsetMinimum, pump.ScheduleTimeZoneOffsetMaximum), "/scheduleTimeZoneOffset", pumpTest.NewMeta()),
				),
				Entry("schedules time zone offset in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.ScheduleTimeZoneOffset = pointer.FromInt(pump.ScheduleTimeZoneOffsetMinimum)
					},
				),
				Entry("schedules time zone offset in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.ScheduleTimeZoneOffset = pointer.FromInt(pump.ScheduleTimeZoneOffsetMaximum)
					},
				),
				Entry("schedules time zone offset out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.ScheduleTimeZoneOffset = pointer.FromInt(pump.ScheduleTimeZoneOffsetMaximum + 1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(pump.ScheduleTimeZoneOffsetMaximum+1, pump.ScheduleTimeZoneOffsetMinimum, pump.ScheduleTimeZoneOffsetMaximum), "/scheduleTimeZoneOffset", pumpTest.NewMeta()),
				),
				Entry("serial number missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.SerialNumber = nil },
				),
				Entry("serial number empty",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) { datum.SerialNumber = pointer.FromString("") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/serialNumber", pumpTest.NewMeta()),
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
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/serialNumber", pumpTest.NewMeta()),
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
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units/bg", pumpTest.NewMeta()),
				),
				Entry("units valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.Units = pumpTest.RandomUnits(unitsBloodGlucose)
					},
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *pump.Pump, unitsBloodGlucose *string) {
						datum.Type = "invalidType"
						datum.ActiveScheduleName = pointer.FromString("")
						datum.Basal.Temporary.Type = nil
						datum.BasalRateSchedules = nil
						datum.BloodGlucoseTargetPhysicalActivity = dataBloodGlucose.NewTarget()
						datum.BloodGlucoseTargetPreprandial = dataBloodGlucose.NewTarget()
						datum.BloodGlucoseTargetSchedules = nil
						datum.Bolus.Extended.Enabled = nil
						datum.CarbohydrateRatioSchedules = nil
						datum.Display.BloodGlucose.Units = nil
						datum.InsulinModel.ModelType = pointer.FromString("invalid")
						datum.InsulinModel.ModelTypeOther = nil
						datum.InsulinSensitivitySchedules = nil
						datum.Manufacturers = pointer.FromStringArray([]string{})
						datum.Model = pointer.FromString("")
						datum.ScheduleTimeZoneOffset = pointer.FromInt(pump.ScheduleTimeZoneOffsetMinimum - 1)
						datum.SerialNumber = pointer.FromString("")
						datum.Units = pumpTest.RandomUnits(pointer.FromString("invalid"))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "pumpSettings"), "/type", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/activeSchedule", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/basal/temporary/type", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/basalSchedule", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgTargetPhysicalActivity/target", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgTargetPreprandial/target", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgTarget", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolus/extended/enabled", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/carbRatio", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/display/bloodGlucose/units", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"fiasp", "other", "rapidAdult", "rapidChild", "walsh"}), "/insulinModel/modelType", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/insulinSensitivity", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/manufacturers", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/model", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(pump.ScheduleTimeZoneOffsetMinimum-1, pump.ScheduleTimeZoneOffsetMinimum, pump.ScheduleTimeZoneOffsetMaximum), "/scheduleTimeZoneOffset", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/serialNumber", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units/bg", &types.Meta{Type: "invalidType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum with origin external",
				func(unitsBloodGlucose *string, mutator func(datum *pump.Pump, unitsBloodGlucose *string), expectator func(datum *pump.Pump, expectedDatum *pump.Pump, unitsBloodGlucose *string)) {
					datum := pumpTest.NewPump(unitsBloodGlucose)
					mutator(datum, unitsBloodGlucose)
					expectedDatum := pumpTest.ClonePump(datum)
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
						datum.BasalRateSchedule = pumpTest.NewBasalRateStartArray()
						datum.BloodGlucoseSuspendThreshold = pointer.FromFloat64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(unitsBloodGlucose)))
						datum.BloodGlucoseTargetPhysicalActivity = dataBloodGlucoseTest.NewTarget(unitsBloodGlucose)
						datum.BloodGlucoseTargetPreprandial = dataBloodGlucoseTest.NewTarget(unitsBloodGlucose)
						datum.BloodGlucoseTargetSchedule = pumpTest.RandomBloodGlucoseTargetStartArray(unitsBloodGlucose)
						datum.CarbohydrateRatioSchedule = pumpTest.NewCarbohydrateRatioStartArray()
						datum.InsulinSensitivitySchedule = pumpTest.NewInsulinSensitivityStartArray(unitsBloodGlucose)
					},
					func(datum *pump.Pump, expectedDatum *pump.Pump, unitsBloodGlucose *string) {
						for index := range *datum.BloodGlucoseTargetSchedule {
							dataBloodGlucoseTest.ExpectNormalizedTarget(&(*datum.BloodGlucoseTargetSchedule)[index].Target, &(*expectedDatum.BloodGlucoseTargetSchedule)[index].Target, unitsBloodGlucose)
						}
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.BloodGlucoseSuspendThreshold, expectedDatum.BloodGlucoseSuspendThreshold, unitsBloodGlucose)
						dataBloodGlucoseTest.ExpectNormalizedTarget(datum.BloodGlucoseTargetPhysicalActivity, expectedDatum.BloodGlucoseTargetPhysicalActivity, unitsBloodGlucose)
						dataBloodGlucoseTest.ExpectNormalizedTarget(datum.BloodGlucoseTargetPreprandial, expectedDatum.BloodGlucoseTargetPreprandial, unitsBloodGlucose)
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
						datum.BasalRateSchedule = pumpTest.NewBasalRateStartArray()
						datum.BloodGlucoseSuspendThreshold = pointer.FromFloat64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(unitsBloodGlucose)))
						datum.BloodGlucoseTargetPhysicalActivity = dataBloodGlucoseTest.NewTarget(unitsBloodGlucose)
						datum.BloodGlucoseTargetPreprandial = dataBloodGlucoseTest.NewTarget(unitsBloodGlucose)
						datum.BloodGlucoseTargetSchedule = pumpTest.RandomBloodGlucoseTargetStartArray(unitsBloodGlucose)
						datum.CarbohydrateRatioSchedule = pumpTest.NewCarbohydrateRatioStartArray()
						datum.InsulinSensitivitySchedule = pumpTest.NewInsulinSensitivityStartArray(unitsBloodGlucose)
					},
					func(datum *pump.Pump, expectedDatum *pump.Pump, unitsBloodGlucose *string) {
						for index := range *datum.BloodGlucoseTargetSchedule {
							dataBloodGlucoseTest.ExpectNormalizedTarget(&(*datum.BloodGlucoseTargetSchedule)[index].Target, &(*expectedDatum.BloodGlucoseTargetSchedule)[index].Target, unitsBloodGlucose)
						}
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.BloodGlucoseSuspendThreshold, expectedDatum.BloodGlucoseSuspendThreshold, unitsBloodGlucose)
						dataBloodGlucoseTest.ExpectNormalizedTarget(datum.BloodGlucoseTargetPhysicalActivity, expectedDatum.BloodGlucoseTargetPhysicalActivity, unitsBloodGlucose)
						dataBloodGlucoseTest.ExpectNormalizedTarget(datum.BloodGlucoseTargetPreprandial, expectedDatum.BloodGlucoseTargetPreprandial, unitsBloodGlucose)
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
						datum := pumpTest.NewPump(unitsBloodGlucose)
						mutator(datum, unitsBloodGlucose)
						expectedDatum := pumpTest.ClonePump(datum)
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
