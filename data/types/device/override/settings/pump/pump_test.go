package pump_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypesDevice "github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/device/override/settings/pump"
	dataTypesDeviceOverrideSettingsPump "github.com/tidepool-org/platform/data/types/device/override/settings/pump"
	dataTypesDeviceOverrideSettingsPumpTest "github.com/tidepool-org/platform/data/types/device/override/settings/pump/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &dataTypesDevice.Meta{
		Type:    "deviceEvent",
		SubType: "pumpSettingsOverride",
	}
}

var _ = Describe("Pump", func() {
	It("SubType is expected", func() {
		Expect(dataTypesDeviceOverrideSettingsPump.SubType).To(Equal("pumpSettingsOverride"))
	})

	It("BasalRateScaleFactorMaximum is expected", func() {
		Expect(dataTypesDeviceOverrideSettingsPump.BasalRateScaleFactorMaximum).To(Equal(10.0))
	})

	It("BasalRateScaleFactorMinimum is expected", func() {
		Expect(dataTypesDeviceOverrideSettingsPump.BasalRateScaleFactorMinimum).To(Equal(0.1))
	})

	It("CarbohydrateRatioScaleFactorMaximum is expected", func() {
		Expect(dataTypesDeviceOverrideSettingsPump.CarbohydrateRatioScaleFactorMaximum).To(Equal(10.0))
	})

	It("CarbohydrateRatioScaleFactorMinimum is expected", func() {
		Expect(dataTypesDeviceOverrideSettingsPump.CarbohydrateRatioScaleFactorMinimum).To(Equal(0.1))
	})

	It("DurationMaximum is expected", func() {
		Expect(dataTypesDeviceOverrideSettingsPump.DurationMaximum).To(Equal(604800000))
	})

	It("DurationMinimum is expected", func() {
		Expect(dataTypesDeviceOverrideSettingsPump.DurationMinimum).To(Equal(0))
	})

	It("InsulinSensitivityScaleFactorMaximum is expected", func() {
		Expect(dataTypesDeviceOverrideSettingsPump.InsulinSensitivityScaleFactorMaximum).To(Equal(10.0))
	})

	It("InsulinSensitivityScaleFactorMinimum is expected", func() {
		Expect(dataTypesDeviceOverrideSettingsPump.InsulinSensitivityScaleFactorMinimum).To(Equal(0.1))
	})

	It("MethodAutomatic is expected", func() {
		Expect(dataTypesDeviceOverrideSettingsPump.MethodAutomatic).To(Equal("automatic"))
	})

	It("MethodManual is expected", func() {
		Expect(dataTypesDeviceOverrideSettingsPump.MethodManual).To(Equal("manual"))
	})

	It("OverridePresetLengthMaximum is expected", func() {
		Expect(dataTypesDeviceOverrideSettingsPump.OverridePresetLengthMaximum).To(Equal(100))
	})

	It("OverrideTypeCustom is expected", func() {
		Expect(dataTypesDeviceOverrideSettingsPump.OverrideTypeCustom).To(Equal("custom"))
	})

	It("OverrideTypePhysicalActivity is expected", func() {
		Expect(dataTypesDeviceOverrideSettingsPump.OverrideTypePhysicalActivity).To(Equal("physicalActivity"))
	})

	It("OverrideTypePreprandial is expected", func() {
		Expect(dataTypesDeviceOverrideSettingsPump.OverrideTypePreprandial).To(Equal("preprandial"))
	})

	It("OverrideTypePreset is expected", func() {
		Expect(dataTypesDeviceOverrideSettingsPump.OverrideTypePreset).To(Equal("preset"))
	})

	It("OverrideTypeSleep is expected", func() {
		Expect(dataTypesDeviceOverrideSettingsPump.OverrideTypeSleep).To(Equal("sleep"))
	})

	It("Methods returns expected", func() {
		Expect(dataTypesDeviceOverrideSettingsPump.Methods()).To(Equal([]string{"automatic", "manual"}))
	})

	It("OverrideTypes returns expected", func() {
		Expect(dataTypesDeviceOverrideSettingsPump.OverrideTypes()).To(Equal([]string{"custom", "physicalActivity", "preprandial", "preset", "sleep"}))
	})

	Context("Pump", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesDeviceOverrideSettingsPump.Pump)) {
				datum := dataTypesDeviceOverrideSettingsPumpTest.RandomPump(pointer.FromString(dataBloodGlucoseTest.RandomUnits()))
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesDeviceOverrideSettingsPumpTest.NewObjectFromPump(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesDeviceOverrideSettingsPumpTest.NewObjectFromPump(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesDeviceOverrideSettingsPump.Pump) {},
			),
			Entry("empty",
				func(datum *dataTypesDeviceOverrideSettingsPump.Pump) {
					*datum = *dataTypesDeviceOverrideSettingsPump.New()
				},
			),
			Entry("all",
				func(datum *dataTypesDeviceOverrideSettingsPump.Pump) {
					units := dataTypesDeviceOverrideSettingsPumpTest.RandomUnits()
					datum.OverrideType = pointer.FromString(dataTypesDeviceOverrideSettingsPumpTest.RandomOverrideType())
					datum.OverridePreset = pointer.FromString(test.RandomString())
					datum.Method = pointer.FromString(dataTypesDeviceOverrideSettingsPumpTest.RandomMethod())
					datum.Duration = pointer.FromInt(test.RandomIntFromRange(dataTypesDeviceOverrideSettingsPump.DurationMinimum, dataTypesDeviceOverrideSettingsPump.DurationMaximum-1))
					datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesDeviceOverrideSettingsPump.DurationMaximum))
					datum.BloodGlucoseTarget = dataBloodGlucoseTest.RandomTarget(units.BloodGlucose)
					datum.BasalRateScaleFactor = pointer.FromFloat64(dataTypesDeviceOverrideSettingsPumpTest.RandomBasalRateScaleFactor())
					datum.CarbohydrateRatioScaleFactor = pointer.FromFloat64(dataTypesDeviceOverrideSettingsPumpTest.RandomCarbohydrateRatioScaleFactor())
					datum.InsulinSensitivityScaleFactor = pointer.FromFloat64(dataTypesDeviceOverrideSettingsPumpTest.RandomInsulinSensitivityScaleFactor())
					datum.Units = units
				},
			),
		)

		Context("New", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesDeviceOverrideSettingsPump.New()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Type).To(Equal("deviceEvent"))
				Expect(datum.SubType).To(Equal("pumpSettingsOverride"))
				Expect(datum.OverrideType).To(BeNil())
				Expect(datum.OverridePreset).To(BeNil())
				Expect(datum.Method).To(BeNil())
				Expect(datum.Duration).To(BeNil())
				Expect(datum.DurationExpected).To(BeNil())
				Expect(datum.BloodGlucoseTarget).To(BeNil())
				Expect(datum.BasalRateScaleFactor).To(BeNil())
				Expect(datum.CarbohydrateRatioScaleFactor).To(BeNil())
				Expect(datum.InsulinSensitivityScaleFactor).To(BeNil())
				Expect(datum.Units).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesDeviceOverrideSettingsPump.Pump), expectedErrors ...error) {
					expectedDatum := dataTypesDeviceOverrideSettingsPumpTest.RandomPumpForParser(pointer.FromString(dataBloodGlucoseTest.RandomUnits()))
					object := dataTypesDeviceOverrideSettingsPumpTest.NewObjectFromPump(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesDeviceOverrideSettingsPump.New()
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesDeviceOverrideSettingsPump.Pump) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesDeviceOverrideSettingsPump.Pump) {
						object["overrideType"] = true
						object["overridePreset"] = true
						object["method"] = true
						object["duration"] = true
						object["expectedDuration"] = true
						object["bgTarget"] = true
						object["basalRateScaleFactor"] = true
						object["carbRatioScaleFactor"] = true
						object["insulinSensitivityScaleFactor"] = true
						object["units"] = true
						expectedDatum.OverrideType = nil
						expectedDatum.OverridePreset = nil
						expectedDatum.Method = nil
						expectedDatum.Duration = nil
						expectedDatum.DurationExpected = nil
						expectedDatum.BloodGlucoseTarget = nil
						expectedDatum.BasalRateScaleFactor = nil
						expectedDatum.CarbohydrateRatioScaleFactor = nil
						expectedDatum.InsulinSensitivityScaleFactor = nil
						expectedDatum.Units = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotString(true), "/overrideType", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotString(true), "/overridePreset", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotString(true), "/method", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotInt(true), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotInt(true), "/expectedDuration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotObject(true), "/bgTarget", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotFloat64(true), "/basalRateScaleFactor", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotFloat64(true), "/carbRatioScaleFactor", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotFloat64(true), "/insulinSensitivityScaleFactor", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotObject(true), "/units", NewMeta()),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(unitsBloodGlucose *string, mutator func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string), expectedErrors ...error) {
					datum := dataTypesDeviceOverrideSettingsPumpTest.RandomPump(unitsBloodGlucose)
					mutator(datum, unitsBloodGlucose)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {},
				),
				Entry("sub type missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) { datum.SubType = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &dataTypesDevice.Meta{Type: "deviceEvent", SubType: ""}),
				),
				Entry("sub type invalid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.SubType = "invalidSubType"
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "pumpSettingsOverride"), "/subType", &dataTypesDevice.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
				),
				Entry("sub type pumpSettings",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.SubType = "pumpSettingsOverride"
					},
				),
				Entry("override type missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.OverrideType = nil
						datum.OverridePreset = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/overrideType", NewMeta()),
				),
				Entry("override type invalid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.OverrideType = pointer.FromString("invalid")
						datum.OverridePreset = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"custom", "physicalActivity", "preprandial", "preset", "sleep"}), "/overrideType", NewMeta()),
				),
				Entry("override type custom",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.OverrideType = pointer.FromString(dataTypesDeviceOverrideSettingsPump.OverrideTypeCustom)
						datum.OverridePreset = nil
					},
				),
				Entry("override type physical activity",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.OverrideType = pointer.FromString(dataTypesDeviceOverrideSettingsPump.OverrideTypePhysicalActivity)
						datum.OverridePreset = nil
					},
				),
				Entry("override type preprandial",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.OverrideType = pointer.FromString(dataTypesDeviceOverrideSettingsPump.OverrideTypePreprandial)
						datum.OverridePreset = nil
					},
				),
				Entry("override type preset",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.OverrideType = pointer.FromString(dataTypesDeviceOverrideSettingsPump.OverrideTypePreset)
						datum.OverridePreset = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
				),
				Entry("override type sleep",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.OverrideType = pointer.FromString(dataTypesDeviceOverrideSettingsPump.OverrideTypeSleep)
						datum.OverridePreset = nil
					},
				),
				Entry("override preset exists",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.OverrideType = pointer.FromString(dataTypesDeviceOverrideSettingsPump.OverrideTypeCustom)
						datum.OverridePreset = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/overridePreset", NewMeta()),
				),
				Entry("override preset missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.OverrideType = pointer.FromString(dataTypesDeviceOverrideSettingsPump.OverrideTypePreset)
						datum.OverridePreset = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/overridePreset", NewMeta()),
				),
				Entry("override preset empty",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.OverrideType = pointer.FromString(dataTypesDeviceOverrideSettingsPump.OverrideTypePreset)
						datum.OverridePreset = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/overridePreset", NewMeta()),
				),
				Entry("override preset length; in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.OverrideType = pointer.FromString(dataTypesDeviceOverrideSettingsPump.OverrideTypePreset)
						datum.OverridePreset = pointer.FromString(test.RandomStringFromRange(100, 100))
					},
				),
				Entry("override preset length; out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.OverrideType = pointer.FromString(dataTypesDeviceOverrideSettingsPump.OverrideTypePreset)
						datum.OverridePreset = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/overridePreset", NewMeta()),
				),
				Entry("override preset valid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.OverrideType = pointer.FromString(dataTypesDeviceOverrideSettingsPump.OverrideTypePreset)
						datum.OverridePreset = pointer.FromString(test.RandomStringFromRange(1, dataTypesDeviceOverrideSettingsPump.OverridePresetLengthMaximum))
					},
				),
				Entry("method missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.Method = nil
					},
				),
				Entry("method invalid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.Method = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"automatic", "manual"}), "/method", NewMeta()),
				),
				Entry("method automatic",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.Method = pointer.FromString(dataTypesDeviceOverrideSettingsPump.MethodAutomatic)
					},
				),
				Entry("method manual",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.Method = pointer.FromString(dataTypesDeviceOverrideSettingsPump.MethodManual)
					},
				),
				Entry("duration missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.Duration = nil
						datum.DurationExpected = nil
					},
				),
				Entry("duration; out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, pump.DurationMaximum), "/duration", NewMeta()),
				),
				Entry("duration; in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = nil
					},
				),
				Entry("duration; in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.Duration = pointer.FromInt(604800)
						datum.DurationExpected = nil
					},
				),
				Entry("duration; out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.Duration = pointer.FromInt(pump.DurationMaximum + 1)
						datum.DurationExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(pump.DurationMaximum+1, 0, pump.DurationMaximum), "/duration", NewMeta()),
				),
				Entry("duration expected missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.Duration = pointer.FromInt(test.RandomIntFromRange(0, pump.DurationMaximum))
						datum.DurationExpected = nil
					},
				),
				Entry("duration expected; duration missing; out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, pump.DurationMaximum), "/expectedDuration", NewMeta()),
				),
				Entry("duration expected; duration missing; in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(0)
					},
				),
				Entry("duration expected; duration missing; in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(pump.DurationMaximum)
					},
				),
				Entry("duration expected; duration missing; out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(pump.DurationMaximum + 1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(pump.DurationMaximum+1, 0, pump.DurationMaximum), "/expectedDuration", NewMeta()),
				),
				Entry("duration expected; duration out of range; out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, pump.DurationMaximum), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, pump.DurationMaximum), "/expectedDuration", NewMeta()),
				),
				Entry("duration expected; duration out of range; in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, pump.DurationMaximum), "/duration", NewMeta()),
				),
				Entry("duration expected; out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.Duration = pointer.FromInt(3600)
						datum.DurationExpected = pointer.FromInt(3599)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(3599, 3600, pump.DurationMaximum), "/expectedDuration", NewMeta()),
				),
				Entry("duration expected; in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.Duration = pointer.FromInt(3600)
						datum.DurationExpected = pointer.FromInt(3600)
					},
				),
				Entry("duration expected; in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.Duration = pointer.FromInt(3600)
						datum.DurationExpected = pointer.FromInt(604800)
					},
				),
				Entry("duration expected; out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.Duration = pointer.FromInt(3600)
						datum.DurationExpected = pointer.FromInt(pump.DurationMaximum + 1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(pump.DurationMaximum+1, 3600, pump.DurationMaximum), "/expectedDuration", NewMeta()),
				),
				Entry("units mmol/L; blood glucose target missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, units *string) { datum.BloodGlucoseTarget = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/units", NewMeta()),
				),
				Entry("units mmol/L; blood glucose target invalid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, units *string) {
						datum.BloodGlucoseTarget = dataBloodGlucose.NewTarget()
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgTarget/target", NewMeta()),
				),
				Entry("units mmol/L; blood glucose target valid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.BloodGlucoseTarget = dataBloodGlucoseTest.RandomTarget(unitsBloodGlucose)
					},
				),
				Entry("units mmol/l; blood glucose target missing",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, units *string) { datum.BloodGlucoseTarget = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/units", NewMeta()),
				),
				Entry("units mmol/l; blood glucose target invalid",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, units *string) {
						datum.BloodGlucoseTarget = dataBloodGlucose.NewTarget()
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgTarget/target", NewMeta()),
				),
				Entry("units mmol/l; blood glucose target valid",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.BloodGlucoseTarget = dataBloodGlucoseTest.RandomTarget(unitsBloodGlucose)
					},
				),
				Entry("units mg/dL; blood glucose target missing",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, units *string) { datum.BloodGlucoseTarget = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/units", NewMeta()),
				),
				Entry("units mg/dL; blood glucose target invalid",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, units *string) {
						datum.BloodGlucoseTarget = dataBloodGlucose.NewTarget()
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgTarget/target", NewMeta()),
				),
				Entry("units mg/dL; blood glucose target valid",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.BloodGlucoseTarget = dataBloodGlucoseTest.RandomTarget(unitsBloodGlucose)
					},
				),
				Entry("units mg/dl; blood glucose target missing",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, units *string) { datum.BloodGlucoseTarget = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/units", NewMeta()),
				),
				Entry("units mg/dl; blood glucose target invalid",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, units *string) {
						datum.BloodGlucoseTarget = dataBloodGlucose.NewTarget()
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgTarget/target", NewMeta()),
				),
				Entry("units mg/dl; blood glucose target valid",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.BloodGlucoseTarget = dataBloodGlucoseTest.RandomTarget(unitsBloodGlucose)
					},
				),
				Entry("basal rate scale factor missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.BasalRateScaleFactor = nil
					},
				),
				Entry("basal rate scale factor; out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.BasalRateScaleFactor = pointer.FromFloat64(0.09)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(0.09, 0.1, 10.0), "/basalRateScaleFactor", NewMeta()),
				),
				Entry("basal rate scale factor; in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.BasalRateScaleFactor = pointer.FromFloat64(0.1)
					},
				),
				Entry("basal rate scale factor; in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.BasalRateScaleFactor = pointer.FromFloat64(10.0)
					},
				),
				Entry("basal rate scale factor; out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.BasalRateScaleFactor = pointer.FromFloat64(10.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(10.1, 0.1, 10.0), "/basalRateScaleFactor", NewMeta()),
				),
				Entry("carbohydrate ratio scale factor missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.CarbohydrateRatioScaleFactor = nil
					},
				),
				Entry("carbohydrate ratio scale factor; out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.CarbohydrateRatioScaleFactor = pointer.FromFloat64(0.09)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(0.09, 0.1, 10.0), "/carbRatioScaleFactor", NewMeta()),
				),
				Entry("carbohydrate ratio scale factor; in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.CarbohydrateRatioScaleFactor = pointer.FromFloat64(0.1)
					},
				),
				Entry("carbohydrate ratio scale factor; in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.CarbohydrateRatioScaleFactor = pointer.FromFloat64(10.0)
					},
				),
				Entry("carbohydrate ratio scale factor; out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.CarbohydrateRatioScaleFactor = pointer.FromFloat64(10.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(10.1, 0.1, 10.0), "/carbRatioScaleFactor", NewMeta()),
				),
				Entry("insulin sensitivity scale factor missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.InsulinSensitivityScaleFactor = nil
					},
				),
				Entry("insulin sensitivity scale factor; out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.InsulinSensitivityScaleFactor = pointer.FromFloat64(0.09)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(0.09, 0.1, 10.0), "/insulinSensitivityScaleFactor", NewMeta()),
				),
				Entry("insulin sensitivity scale factor; in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.InsulinSensitivityScaleFactor = pointer.FromFloat64(0.1)
					},
				),
				Entry("insulin sensitivity scale factor; in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.InsulinSensitivityScaleFactor = pointer.FromFloat64(10.0)
					},
				),
				Entry("insulin sensitivity scale factor; out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.InsulinSensitivityScaleFactor = pointer.FromFloat64(10.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(10.1, 0.1, 10.0), "/insulinSensitivityScaleFactor", NewMeta()),
				),
				Entry("units missing; blood glucose target missing",
					nil,
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.BloodGlucoseTarget = nil
					},
				),
				Entry("units missing; blood glucose target valid",
					nil,
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.BloodGlucoseTarget = dataBloodGlucoseTest.RandomTarget(pointer.FromString("mmol/L"))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units invalid; blood glucose target valid",
					pointer.FromString("invalid"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.BloodGlucoseTarget = dataBloodGlucoseTest.RandomTarget(pointer.FromString("mmol/L"))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units/bg", NewMeta()),
				),
				Entry("units valid; blood glucose target missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.BloodGlucoseTarget = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/units", NewMeta()),
				),
				Entry("units valid; blood glucose target valid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.BloodGlucoseTarget = dataBloodGlucoseTest.RandomTarget(unitsBloodGlucose)
					},
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.SubType = "invalidSubType"
						datum.OverrideType = nil
						datum.OverridePreset = pointer.FromString(test.RandomStringFromRange(1, 100))
						datum.Method = pointer.FromString("invalid")
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(-1)
						datum.BloodGlucoseTarget = nil
						datum.BasalRateScaleFactor = pointer.FromFloat64(0.09)
						datum.CarbohydrateRatioScaleFactor = pointer.FromFloat64(0.09)
						datum.InsulinSensitivityScaleFactor = pointer.FromFloat64(0.09)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "pumpSettingsOverride"), "/subType", &dataTypesDevice.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/overrideType", &dataTypesDevice.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/overridePreset", &dataTypesDevice.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"automatic", "manual"}), "/method", &dataTypesDevice.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, pump.DurationMaximum), "/duration", &dataTypesDevice.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, pump.DurationMaximum), "/expectedDuration", &dataTypesDevice.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(0.09, 0.1, 10.0), "/basalRateScaleFactor", &dataTypesDevice.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(0.09, 0.1, 10.0), "/carbRatioScaleFactor", &dataTypesDevice.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(0.09, 0.1, 10.0), "/insulinSensitivityScaleFactor", &dataTypesDevice.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/units", &dataTypesDevice.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum with origin external",
				func(unitsBloodGlucose *string, mutator func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string), expectator func(datum *dataTypesDeviceOverrideSettingsPump.Pump, expectedDatum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string)) {
					datum := dataTypesDeviceOverrideSettingsPumpTest.RandomPump(unitsBloodGlucose)
					mutator(datum, unitsBloodGlucose)
					expectedDatum := dataTypesDeviceOverrideSettingsPumpTest.ClonePump(datum)
					normalizer := dataNormalizer.New()
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(structure.OriginExternal))
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					dataBloodGlucoseTest.SetTargetRaw(expectedDatum.BloodGlucoseTarget, datum.BloodGlucoseTarget)
					dataTypesDeviceOverrideSettingsPumpTest.SetUnitsRaw(expectedDatum.Units, datum.Units)
					if expectator != nil {
						expectator(datum, expectedDatum, unitsBloodGlucose)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("modifies the datum",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {},
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, expectedDatum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
					},
				),
				Entry("modifies the datum; units missing",
					nil,
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {},
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, expectedDatum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
					},
				),
				Entry("modifies the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {},
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, expectedDatum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
					},
				),
				Entry("modifies the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {},
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, expectedDatum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
					},
				),
				Entry("modifies the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {},
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, expectedDatum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units.BloodGlucose, expectedDatum.Units.BloodGlucose)
					},
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.BloodGlucoseTarget = dataBloodGlucoseTest.RandomTarget(unitsBloodGlucose)
					},
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, expectedDatum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						dataBloodGlucoseTest.ExpectNormalizedTarget(datum.BloodGlucoseTarget, expectedDatum.BloodGlucoseTarget, unitsBloodGlucose)
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units.BloodGlucose, expectedDatum.Units.BloodGlucose)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						datum.BloodGlucoseTarget = dataBloodGlucoseTest.RandomTarget(unitsBloodGlucose)
					},
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, expectedDatum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {
						dataBloodGlucoseTest.ExpectNormalizedTarget(datum.BloodGlucoseTarget, expectedDatum.BloodGlucoseTarget, unitsBloodGlucose)
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units.BloodGlucose, expectedDatum.Units.BloodGlucose)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(unitsBloodGlucose *string, mutator func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string), expectator func(datum *dataTypesDeviceOverrideSettingsPump.Pump, expectedDatum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := dataTypesDeviceOverrideSettingsPumpTest.RandomPump(unitsBloodGlucose)
						mutator(datum, unitsBloodGlucose)
						expectedDatum := dataTypesDeviceOverrideSettingsPumpTest.ClonePump(datum)
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
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesDeviceOverrideSettingsPump.Pump, unitsBloodGlucose *string) {},
					nil,
				),
			)
		})
	})
})
