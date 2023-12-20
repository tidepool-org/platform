package pump_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypesSettingsPump "github.com/tidepool-org/platform/data/types/settings/pump"
	dataTypesSettingsPumpTest "github.com/tidepool-org/platform/data/types/settings/pump/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("OverridePreset", func() {
	It("AbbreviationLengthMaximum is expected", func() {
		Expect(dataTypesSettingsPump.AbbreviationLengthMaximum).To(Equal(100))
	})

	It("BasalRateScaleFactorMaximum is expected", func() {
		Expect(dataTypesSettingsPump.BasalRateScaleFactorMaximum).To(Equal(10.0))
	})

	It("BasalRateScaleFactorMinimum is expected", func() {
		Expect(dataTypesSettingsPump.BasalRateScaleFactorMinimum).To(Equal(0.1))
	})

	It("CarbohydrateRatioScaleFactorMaximum is expected", func() {
		Expect(dataTypesSettingsPump.CarbohydrateRatioScaleFactorMaximum).To(Equal(10.0))
	})

	It("CarbohydrateRatioScaleFactorMinimum is expected", func() {
		Expect(dataTypesSettingsPump.CarbohydrateRatioScaleFactorMinimum).To(Equal(0.1))
	})

	It("DurationMaximum is expected", func() {
		Expect(dataTypesSettingsPump.DurationMaximum).To(Equal(604800000))
	})

	It("DurationMinimum is expected", func() {
		Expect(dataTypesSettingsPump.DurationMinimum).To(Equal(0))
	})

	It("InsulinSensitivityScaleFactorMaximum is expected", func() {
		Expect(dataTypesSettingsPump.InsulinSensitivityScaleFactorMaximum).To(Equal(10.0))
	})

	It("InsulinSensitivityScaleFactorMinimum is expected", func() {
		Expect(dataTypesSettingsPump.InsulinSensitivityScaleFactorMinimum).To(Equal(0.1))
	})

	Context("OverridePreset", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesSettingsPump.OverridePreset)) {
				datum := dataTypesSettingsPumpTest.RandomOverridePreset(pointer.FromString(dataBloodGlucoseTest.RandomUnits()))
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesSettingsPumpTest.NewObjectFromOverridePreset(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesSettingsPumpTest.NewObjectFromOverridePreset(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesSettingsPump.OverridePreset) {},
			),
			Entry("empty",
				func(datum *dataTypesSettingsPump.OverridePreset) {
					*datum = *dataTypesSettingsPump.NewOverridePreset()
				},
			),
			Entry("all",
				func(datum *dataTypesSettingsPump.OverridePreset) {
					datum.Abbreviation = pointer.FromString(dataTypesSettingsPumpTest.RandomAbbreviation())
					datum.Duration = pointer.FromInt(dataTypesSettingsPumpTest.RandomDuration())
					datum.BloodGlucoseTarget = dataBloodGlucoseTest.RandomTarget(pointer.FromString(dataBloodGlucoseTest.RandomUnits()))
					datum.BasalRateScaleFactor = pointer.FromFloat64(dataTypesSettingsPumpTest.RandomBasalRateScaleFactor())
					datum.CarbohydrateRatioScaleFactor = pointer.FromFloat64(dataTypesSettingsPumpTest.RandomCarbohydrateRatioScaleFactor())
					datum.InsulinSensitivityScaleFactor = pointer.FromFloat64(dataTypesSettingsPumpTest.RandomInsulinSensitivityScaleFactor())
				},
			),
		)

		Context("ParseOverridePreset", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesSettingsPump.ParseOverridePreset(structureParser.NewObject(nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesSettingsPumpTest.RandomOverridePreset(pointer.FromString(dataBloodGlucoseTest.RandomUnits()))
				object := dataTypesSettingsPumpTest.NewObjectFromOverridePreset(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(&object)
				Expect(dataTypesSettingsPump.ParseOverridePreset(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewOverridePreset", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesSettingsPump.NewOverridePreset()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Abbreviation).To(BeNil())
				Expect(datum.Duration).To(BeNil())
				Expect(datum.BloodGlucoseTarget).To(BeNil())
				Expect(datum.BasalRateScaleFactor).To(BeNil())
				Expect(datum.CarbohydrateRatioScaleFactor).To(BeNil())
				Expect(datum.InsulinSensitivityScaleFactor).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesSettingsPump.OverridePreset), expectedErrors ...error) {
					expectedDatum := dataTypesSettingsPumpTest.RandomOverridePreset(pointer.FromString(dataBloodGlucoseTest.RandomUnits()))
					object := dataTypesSettingsPumpTest.NewObjectFromOverridePreset(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesSettingsPump.NewOverridePreset()
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesSettingsPump.OverridePreset) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesSettingsPump.OverridePreset) {
						object["abbreviation"] = true
						object["duration"] = true
						object["bgTarget"] = true
						object["basalRateScaleFactor"] = true
						object["carbRatioScaleFactor"] = true
						object["insulinSensitivityScaleFactor"] = true
						expectedDatum.Abbreviation = nil
						expectedDatum.Duration = nil
						expectedDatum.BloodGlucoseTarget = nil
						expectedDatum.BasalRateScaleFactor = nil
						expectedDatum.CarbohydrateRatioScaleFactor = nil
						expectedDatum.InsulinSensitivityScaleFactor = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/abbreviation"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/duration"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/bgTarget"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotFloat64(true), "/basalRateScaleFactor"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotFloat64(true), "/carbRatioScaleFactor"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotFloat64(true), "/insulinSensitivityScaleFactor"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(unitsBloodGlucose *string, mutator func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string), expectedErrors ...error) {
					datum := dataTypesSettingsPumpTest.RandomOverridePreset(unitsBloodGlucose)
					mutator(datum, unitsBloodGlucose)
					dataTypesTest.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithStringAdapter(datum, unitsBloodGlucose), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {},
				),
				Entry("abbreviation missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) { datum.Abbreviation = nil },
				),
				Entry("abbreviation empty",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.Abbreviation = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/abbreviation"),
				),
				Entry("abbreviation length in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.Abbreviation = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
				),
				Entry("abbreviation length out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.Abbreviation = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/abbreviation"),
				),
				Entry("duration missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) { datum.Duration = nil },
				),
				Entry("duration; out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.Duration = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, dataTypesSettingsPump.DurationMaximum), "/duration"),
				),
				Entry("duration; in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.Duration = pointer.FromInt(0)
					},
				),
				Entry("duration; in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.Duration = pointer.FromInt(dataTypesSettingsPump.DurationMaximum)
					},
				),
				Entry("duration; out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.Duration = pointer.FromInt(604800001)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(604800001, 0, dataTypesSettingsPump.DurationMaximum), "/duration"),
				),
				Entry("units mmol/L; blood glucose target missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, units *string) { datum.BloodGlucoseTarget = nil },
				),
				Entry("units mmol/L; blood glucose target invalid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, units *string) {
						datum.BloodGlucoseTarget = dataBloodGlucose.NewTarget()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/bgTarget/target"),
				),
				Entry("units mmol/L; blood glucose target valid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.BloodGlucoseTarget = dataBloodGlucoseTest.RandomTarget(unitsBloodGlucose)
					},
				),
				Entry("units mmol/l; blood glucose target missing",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsPump.OverridePreset, units *string) { datum.BloodGlucoseTarget = nil },
				),
				Entry("units mmol/l; blood glucose target invalid",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsPump.OverridePreset, units *string) {
						datum.BloodGlucoseTarget = dataBloodGlucose.NewTarget()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/bgTarget/target"),
				),
				Entry("units mmol/l; blood glucose target valid",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.BloodGlucoseTarget = dataBloodGlucoseTest.RandomTarget(unitsBloodGlucose)
					},
				),
				Entry("units mg/dL; blood glucose target missing",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsPump.OverridePreset, units *string) { datum.BloodGlucoseTarget = nil },
				),
				Entry("units mg/dL; blood glucose target invalid",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsPump.OverridePreset, units *string) {
						datum.BloodGlucoseTarget = dataBloodGlucose.NewTarget()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/bgTarget/target"),
				),
				Entry("units mg/dL; blood glucose target valid",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.BloodGlucoseTarget = dataBloodGlucoseTest.RandomTarget(unitsBloodGlucose)
					},
				),
				Entry("units mg/dl; blood glucose target missing",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsPump.OverridePreset, units *string) { datum.BloodGlucoseTarget = nil },
				),
				Entry("units mg/dl; blood glucose target invalid",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsPump.OverridePreset, units *string) {
						datum.BloodGlucoseTarget = dataBloodGlucose.NewTarget()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/bgTarget/target"),
				),
				Entry("units mg/dl; blood glucose target valid",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.BloodGlucoseTarget = dataBloodGlucoseTest.RandomTarget(unitsBloodGlucose)
					},
				),
				Entry("basal rate scale factor missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.BasalRateScaleFactor = nil
					},
				),
				Entry("basal rate scale factor; out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.BasalRateScaleFactor = pointer.FromFloat64(0.09)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(0.09, 0.1, 10.0), "/basalRateScaleFactor"),
				),
				Entry("basal rate scale factor; in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.BasalRateScaleFactor = pointer.FromFloat64(0.1)
					},
				),
				Entry("basal rate scale factor; in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.BasalRateScaleFactor = pointer.FromFloat64(10.0)
					},
				),
				Entry("basal rate scale factor; out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.BasalRateScaleFactor = pointer.FromFloat64(10.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(10.1, 0.1, 10.0), "/basalRateScaleFactor"),
				),
				Entry("carbohydrate ratio scale factor missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.CarbohydrateRatioScaleFactor = nil
					},
				),
				Entry("carbohydrate ratio scale factor; out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.CarbohydrateRatioScaleFactor = pointer.FromFloat64(0.09)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(0.09, 0.1, 10.0), "/carbRatioScaleFactor"),
				),
				Entry("carbohydrate ratio scale factor; in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.CarbohydrateRatioScaleFactor = pointer.FromFloat64(0.1)
					},
				),
				Entry("carbohydrate ratio scale factor; in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.CarbohydrateRatioScaleFactor = pointer.FromFloat64(10.0)
					},
				),
				Entry("carbohydrate ratio scale factor; out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.CarbohydrateRatioScaleFactor = pointer.FromFloat64(10.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(10.1, 0.1, 10.0), "/carbRatioScaleFactor"),
				),
				Entry("insulin sensitivity scale factor missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.InsulinSensitivityScaleFactor = nil
					},
				),
				Entry("insulin sensitivity scale factor; out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.InsulinSensitivityScaleFactor = pointer.FromFloat64(0.09)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(0.09, 0.1, 10.0), "/insulinSensitivityScaleFactor"),
				),
				Entry("insulin sensitivity scale factor; in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.InsulinSensitivityScaleFactor = pointer.FromFloat64(0.1)
					},
				),
				Entry("insulin sensitivity scale factor; in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.InsulinSensitivityScaleFactor = pointer.FromFloat64(10.0)
					},
				),
				Entry("insulin sensitivity scale factor; out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.InsulinSensitivityScaleFactor = pointer.FromFloat64(10.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(10.1, 0.1, 10.0), "/insulinSensitivityScaleFactor"),
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.Abbreviation = pointer.FromString("")
						datum.Duration = pointer.FromInt(-1)
						datum.BloodGlucoseTarget = nil
						datum.BasalRateScaleFactor = pointer.FromFloat64(0.09)
						datum.CarbohydrateRatioScaleFactor = pointer.FromFloat64(0.09)
						datum.InsulinSensitivityScaleFactor = pointer.FromFloat64(0.09)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/abbreviation"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, dataTypesSettingsPump.DurationMaximum), "/duration"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(0.09, 0.1, 10.0), "/basalRateScaleFactor"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(0.09, 0.1, 10.0), "/carbRatioScaleFactor"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(0.09, 0.1, 10.0), "/insulinSensitivityScaleFactor"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum with origin external",
				func(unitsBloodGlucose *string, mutator func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string), expectator func(datum *dataTypesSettingsPump.OverridePreset, expectedDatum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string)) {
					datum := dataTypesSettingsPumpTest.RandomOverridePreset(unitsBloodGlucose)
					mutator(datum, unitsBloodGlucose)
					expectedDatum := dataTypesSettingsPumpTest.CloneOverridePreset(datum)
					normalizer := dataNormalizer.New()
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(structure.OriginExternal), unitsBloodGlucose)
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					if expectator != nil {
						expectator(datum, expectedDatum, unitsBloodGlucose)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("modifies the datum",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {},
					func(datum *dataTypesSettingsPump.OverridePreset, expectedDatum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
					},
				),
				Entry("modifies the datum; units missing",
					nil,
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {},
					func(datum *dataTypesSettingsPump.OverridePreset, expectedDatum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
					},
				),
				Entry("modifies the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {},
					func(datum *dataTypesSettingsPump.OverridePreset, expectedDatum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
					},
				),
				Entry("modifies the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {},
					func(datum *dataTypesSettingsPump.OverridePreset, expectedDatum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
					},
				),
				Entry("modifies the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {},
					func(datum *dataTypesSettingsPump.OverridePreset, expectedDatum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
					},
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.BloodGlucoseTarget = dataBloodGlucoseTest.RandomTarget(unitsBloodGlucose)
					},
					func(datum *dataTypesSettingsPump.OverridePreset, expectedDatum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						dataBloodGlucoseTest.ExpectNormalizedTarget(datum.BloodGlucoseTarget, expectedDatum.BloodGlucoseTarget, unitsBloodGlucose)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						datum.BloodGlucoseTarget = dataBloodGlucoseTest.RandomTarget(unitsBloodGlucose)
					},
					func(datum *dataTypesSettingsPump.OverridePreset, expectedDatum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {
						dataBloodGlucoseTest.ExpectNormalizedTarget(datum.BloodGlucoseTarget, expectedDatum.BloodGlucoseTarget, unitsBloodGlucose)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(unitsBloodGlucose *string, mutator func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string), expectator func(datum *dataTypesSettingsPump.OverridePreset, expectedDatum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := dataTypesSettingsPumpTest.RandomOverridePreset(unitsBloodGlucose)
						mutator(datum, unitsBloodGlucose)
						expectedDatum := dataTypesSettingsPumpTest.CloneOverridePreset(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin), unitsBloodGlucose)
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
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsPump.OverridePreset, unitsBloodGlucose *string) {},
					nil,
				),
			)
		})
	})

	Context("OverridePresetMap", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesSettingsPump.OverridePresetMap)) {
				datum := dataTypesSettingsPumpTest.RandomOverridePresetMap(pointer.FromString(dataBloodGlucoseTest.RandomUnits()))
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesSettingsPumpTest.NewObjectFromOverridePresetMap(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesSettingsPumpTest.NewObjectFromOverridePresetMap(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesSettingsPump.OverridePresetMap) {},
			),
			Entry("empty",
				func(datum *dataTypesSettingsPump.OverridePresetMap) {
					*datum = *dataTypesSettingsPump.NewOverridePresetMap()
				},
			),
		)

		Context("ParseOverridePresetMap", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesSettingsPump.ParseOverridePresetMap(structureParser.NewObject(nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesSettingsPumpTest.RandomOverridePresetMap(pointer.FromString(dataBloodGlucoseTest.RandomUnits()))
				object := dataTypesSettingsPumpTest.NewObjectFromOverridePresetMap(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(&object)
				Expect(dataTypesSettingsPump.ParseOverridePresetMap(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewOverridePresetMap", func() {
			It("is successful", func() {
				Expect(dataTypesSettingsPump.NewOverridePresetMap()).To(Equal(&dataTypesSettingsPump.OverridePresetMap{}))
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesSettingsPump.OverridePresetMap), expectedErrors ...error) {
					expectedDatum := dataTypesSettingsPumpTest.RandomOverridePresetMap(pointer.FromString(dataBloodGlucoseTest.RandomUnits()))
					object := dataTypesSettingsPumpTest.NewObjectFromOverridePresetMap(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesSettingsPump.NewOverridePresetMap()
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesSettingsPump.OverridePresetMap) {},
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(unitsBloodGlucose *string, mutator func(datum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string), expectedErrors ...error) {
					datum := dataTypesSettingsPump.NewOverridePresetMap()
					mutator(datum, unitsBloodGlucose)
					dataTypesTest.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithStringAdapter(datum, unitsBloodGlucose), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string) {},
				),
				Entry("empty",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string) {
						*datum = *dataTypesSettingsPump.NewOverridePresetMap()
					},
				),
				Entry("empty name",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string) {
						datum.Set("", dataTypesSettingsPumpTest.RandomOverridePreset(unitsBloodGlucose))
					},
				),
				Entry("nil value",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string) {
						datum.Set("", nil)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/"),
				),
				Entry("single invalid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string) {
						invalid := dataTypesSettingsPumpTest.RandomOverridePreset(unitsBloodGlucose)
						invalid.Abbreviation = pointer.FromString("")
						datum.Set("one", invalid)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/one/abbreviation"),
				),
				Entry("single valid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string) {
						datum.Set("one", dataTypesSettingsPumpTest.RandomOverridePreset(unitsBloodGlucose))
					},
				),
				Entry("multiple invalid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string) {
						invalid := dataTypesSettingsPumpTest.RandomOverridePreset(unitsBloodGlucose)
						invalid.Abbreviation = pointer.FromString("")
						datum.Set("one", dataTypesSettingsPumpTest.RandomOverridePreset(unitsBloodGlucose))
						datum.Set("two", invalid)
						datum.Set("three", dataTypesSettingsPumpTest.RandomOverridePreset(unitsBloodGlucose))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/two/abbreviation"),
				),
				Entry("multiple valid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string) {
						datum.Set("one", dataTypesSettingsPumpTest.RandomOverridePreset(unitsBloodGlucose))
						datum.Set("two", dataTypesSettingsPumpTest.RandomOverridePreset(unitsBloodGlucose))
						datum.Set("three", dataTypesSettingsPumpTest.RandomOverridePreset(unitsBloodGlucose))
					},
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string) {
						invalid := dataTypesSettingsPumpTest.RandomOverridePreset(unitsBloodGlucose)
						invalid.Abbreviation = pointer.FromString("")
						datum.Set("one", nil)
						datum.Set("two", invalid)
						datum.Set("three", dataTypesSettingsPumpTest.RandomOverridePreset(unitsBloodGlucose))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/one"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/two/abbreviation"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum with origin external",
				func(unitsBloodGlucose *string, mutator func(datum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string), expectator func(datum *dataTypesSettingsPump.OverridePresetMap, expectedDatum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string)) {
					datum := dataTypesSettingsPumpTest.RandomOverridePresetMap(unitsBloodGlucose)
					mutator(datum, unitsBloodGlucose)
					expectedDatum := dataTypesSettingsPumpTest.CloneOverridePresetMap(datum)
					normalizer := dataNormalizer.New()
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(structure.OriginExternal), unitsBloodGlucose)
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					if expectator != nil {
						expectator(datum, expectedDatum, unitsBloodGlucose)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("modifies the datum",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string) {},
					func(datum *dataTypesSettingsPump.OverridePresetMap, expectedDatum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string) {
					},
				),
				Entry("modifies the datum; units missing",
					nil,
					func(datum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string) {},
					func(datum *dataTypesSettingsPump.OverridePresetMap, expectedDatum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string) {
					},
				),
				Entry("modifies the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string) {},
					func(datum *dataTypesSettingsPump.OverridePresetMap, expectedDatum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string) {
					},
				),
				Entry("modifies the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string) {},
					func(datum *dataTypesSettingsPump.OverridePresetMap, expectedDatum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string) {
					},
				),
				Entry("modifies the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string) {},
					func(datum *dataTypesSettingsPump.OverridePresetMap, expectedDatum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string) {
					},
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string) {
						for name := range *datum {
							datum.Get(name).BloodGlucoseTarget = dataBloodGlucoseTest.RandomTarget(unitsBloodGlucose)
						}
					},
					func(datum *dataTypesSettingsPump.OverridePresetMap, expectedDatum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string) {
						for name := range *datum {
							dataBloodGlucoseTest.ExpectNormalizedTarget(datum.Get(name).BloodGlucoseTarget, expectedDatum.Get(name).BloodGlucoseTarget, unitsBloodGlucose)
						}
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string) {
						for name := range *datum {
							datum.Get(name).BloodGlucoseTarget = dataBloodGlucoseTest.RandomTarget(unitsBloodGlucose)
						}
					},
					func(datum *dataTypesSettingsPump.OverridePresetMap, expectedDatum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string) {
						for name := range *datum {
							dataBloodGlucoseTest.ExpectNormalizedTarget(datum.Get(name).BloodGlucoseTarget, expectedDatum.Get(name).BloodGlucoseTarget, unitsBloodGlucose)
						}
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(unitsBloodGlucose *string, mutator func(datum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string), expectator func(datum *dataTypesSettingsPump.OverridePresetMap, expectedDatum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := dataTypesSettingsPumpTest.RandomOverridePresetMap(unitsBloodGlucose)
						mutator(datum, unitsBloodGlucose)
						expectedDatum := dataTypesSettingsPumpTest.CloneOverridePresetMap(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin), unitsBloodGlucose)
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
					func(datum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsPump.OverridePresetMap, unitsBloodGlucose *string) {},
					nil,
				),
			)
		})
	})
})
