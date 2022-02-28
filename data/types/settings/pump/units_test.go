package pump_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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

var _ = Describe("Units", func() {
	It("CarbohydrateUnitsExchanges is expected", func() {
		Expect(dataTypesSettingsPump.CarbohydrateUnitsExchanges).To(Equal("exchanges"))
	})

	It("CarbohydrateUnitsGrams is expected", func() {
		Expect(dataTypesSettingsPump.CarbohydrateUnitsGrams).To(Equal("grams"))
	})

	It("InsulinUnitsUnits is expected", func() {
		Expect(dataTypesSettingsPump.InsulinUnitsUnits).To(Equal("Units"))
	})

	It("CarbohydrateUnits returns expected", func() {
		Expect(dataTypesSettingsPump.CarbohydrateUnits()).To(Equal([]string{"exchanges", "grams"}))
	})

	It("InsulinUnits returns expected", func() {
		Expect(dataTypesSettingsPump.InsulinUnits()).To(Equal([]string{"Units"}))
	})

	Context("Units", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesSettingsPump.Units)) {
				datum := dataTypesSettingsPumpTest.RandomUnits(pointer.FromString("mg/dL"))
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesSettingsPumpTest.NewObjectFromUnits(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesSettingsPumpTest.NewObjectFromUnits(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesSettingsPump.Units) {},
			),
			Entry("empty",
				func(datum *dataTypesSettingsPump.Units) {
					*datum = *dataTypesSettingsPump.NewUnits()
				},
			),
			Entry("all",
				func(datum *dataTypesSettingsPump.Units) {
					datum.BloodGlucose = pointer.FromString(test.RandomStringFromArray(dataBloodGlucose.Units()))
					datum.Carbohydrate = pointer.FromString(test.RandomStringFromArray(dataTypesSettingsPump.CarbohydrateUnits()))
					datum.Insulin = pointer.FromString(test.RandomStringFromArray(dataTypesSettingsPump.InsulinUnits()))
				},
			),
		)

		Context("ParseUnits", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesSettingsPump.ParseUnits(structureParser.NewObject(nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesSettingsPumpTest.RandomUnits(pointer.FromString("mg/dL"))
				object := dataTypesSettingsPumpTest.NewObjectFromUnits(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(&object)
				Expect(dataTypesSettingsPump.ParseUnits(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewUnits", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesSettingsPump.NewUnits()
				Expect(datum).ToNot(BeNil())
				Expect(datum.BloodGlucose).To(BeNil())
				Expect(datum.Carbohydrate).To(BeNil())
				Expect(datum.Insulin).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesSettingsPump.Units), expectedErrors ...error) {
					expectedDatum := dataTypesSettingsPumpTest.RandomUnits(pointer.FromString("mg/dL"))
					object := dataTypesSettingsPumpTest.NewObjectFromUnits(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesSettingsPump.NewUnits()
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesSettingsPump.Units) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesSettingsPump.Units) {
						object["bg"] = true
						object["carb"] = true
						object["insulin"] = true
						expectedDatum.BloodGlucose = nil
						expectedDatum.Carbohydrate = nil
						expectedDatum.Insulin = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/bg"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/carb"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/insulin"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesSettingsPump.Units), expectedErrors ...error) {
					datum := dataTypesSettingsPumpTest.RandomUnits(pointer.FromString("mmol/L"))
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesSettingsPump.Units) {},
				),
				Entry("blood glucose missing",
					func(datum *dataTypesSettingsPump.Units) { datum.BloodGlucose = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/bg"),
				),
				Entry("blood glucose invalid",
					func(datum *dataTypesSettingsPump.Units) { datum.BloodGlucose = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/bg"),
				),
				Entry("blood glucose mmol/L",
					func(datum *dataTypesSettingsPump.Units) { datum.BloodGlucose = pointer.FromString("mmol/L") },
				),
				Entry("blood glucose mmol/l",
					func(datum *dataTypesSettingsPump.Units) { datum.BloodGlucose = pointer.FromString("mmol/l") },
				),
				Entry("blood glucose mg/dL",
					func(datum *dataTypesSettingsPump.Units) { datum.BloodGlucose = pointer.FromString("mg/dL") },
				),
				Entry("blood glucose mg/dl",
					func(datum *dataTypesSettingsPump.Units) { datum.BloodGlucose = pointer.FromString("mg/dl") },
				),
				Entry("carbohydrate missing",
					func(datum *dataTypesSettingsPump.Units) { datum.Carbohydrate = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/carb"),
				),
				Entry("carbohydrate invalid",
					func(datum *dataTypesSettingsPump.Units) { datum.Carbohydrate = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"exchanges", "grams"}), "/carb"),
				),
				Entry("carbohydrate exchanges",
					func(datum *dataTypesSettingsPump.Units) { datum.Carbohydrate = pointer.FromString("exchanges") },
				),
				Entry("carbohydrate grams",
					func(datum *dataTypesSettingsPump.Units) { datum.Carbohydrate = pointer.FromString("grams") },
				),
				Entry("insulin missing",
					func(datum *dataTypesSettingsPump.Units) { datum.Insulin = nil },
				),
				Entry("insulin invalid",
					func(datum *dataTypesSettingsPump.Units) { datum.Insulin = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"Units"}), "/insulin"),
				),
				Entry("insulin Units",
					func(datum *dataTypesSettingsPump.Units) { datum.Insulin = pointer.FromString("Units") },
				),
				Entry("multiple errors",
					func(datum *dataTypesSettingsPump.Units) {
						datum.BloodGlucose = nil
						datum.Carbohydrate = nil
						datum.Insulin = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/bg"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/carb"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"Units"}), "/insulin"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum with origin external",
				func(mutator func(datum *dataTypesSettingsPump.Units), expectator func(datum *dataTypesSettingsPump.Units, expectedDatum *dataTypesSettingsPump.Units)) {
					datum := dataTypesSettingsPumpTest.RandomUnits(pointer.FromString("mmol/L"))
					mutator(datum)
					expectedDatum := dataTypesSettingsPumpTest.CloneUnits(datum)
					normalizer := dataNormalizer.New()
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(structure.OriginExternal))
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					if expectator != nil {
						expectator(datum, expectedDatum)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("does not modify the datum",
					func(datum *dataTypesSettingsPump.Units) {},
					nil,
				),
				Entry("does not modify the datum; blood glucose missing",
					func(datum *dataTypesSettingsPump.Units) { datum.BloodGlucose = nil },
					nil,
				),
				Entry("does not modify the datum; blood glucose invalid",
					func(datum *dataTypesSettingsPump.Units) { datum.BloodGlucose = pointer.FromString("invalid") },
					nil,
				),
				Entry("does not modify the datum; blood glucose mmol/L",
					func(datum *dataTypesSettingsPump.Units) { datum.BloodGlucose = pointer.FromString("mmol/L") },
					nil,
				),
				Entry("modifies the datum; blood glucose mmol/l",
					func(datum *dataTypesSettingsPump.Units) { datum.BloodGlucose = pointer.FromString("mmol/l") },
					func(datum *dataTypesSettingsPump.Units, expectedDatum *dataTypesSettingsPump.Units) {
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.BloodGlucose, expectedDatum.BloodGlucose)
					},
				),
				Entry("modifies the datum; blood glucose mg/dL",
					func(datum *dataTypesSettingsPump.Units) { datum.BloodGlucose = pointer.FromString("mg/dL") },
					func(datum *dataTypesSettingsPump.Units, expectedDatum *dataTypesSettingsPump.Units) {
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.BloodGlucose, expectedDatum.BloodGlucose)
					},
				),
				Entry("modifies the datum; blood glucose mg/dl",
					func(datum *dataTypesSettingsPump.Units) { datum.BloodGlucose = pointer.FromString("mg/dl") },
					func(datum *dataTypesSettingsPump.Units, expectedDatum *dataTypesSettingsPump.Units) {
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.BloodGlucose, expectedDatum.BloodGlucose)
					},
				),
				Entry("does not modify the datum; carbohydrate missing",
					func(datum *dataTypesSettingsPump.Units) { datum.Carbohydrate = nil },
					nil,
				),
				Entry("does not modify the datum; carbohydrate invalid",
					func(datum *dataTypesSettingsPump.Units) { datum.Carbohydrate = pointer.FromString("invalid") },
					nil,
				),
				Entry("does not modify the datum; carbohydrate exchanges",
					func(datum *dataTypesSettingsPump.Units) { datum.Carbohydrate = pointer.FromString("exchanges") },
					nil,
				),
				Entry("does not modify the datum; carbohydrate grams",
					func(datum *dataTypesSettingsPump.Units) { datum.Carbohydrate = pointer.FromString("grams") },
					nil,
				),
				Entry("does not modify the datum; insulin missing",
					func(datum *dataTypesSettingsPump.Units) { datum.Insulin = nil },
					nil,
				),
				Entry("does not modify the datum; insulin invalid",
					func(datum *dataTypesSettingsPump.Units) { datum.Insulin = pointer.FromString("invalid") },
					nil,
				),
				Entry("does not modify the datum; insulin Units",
					func(datum *dataTypesSettingsPump.Units) { datum.Insulin = pointer.FromString("Units") },
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(mutator func(datum *dataTypesSettingsPump.Units), expectator func(datum *dataTypesSettingsPump.Units, expectedDatum *dataTypesSettingsPump.Units)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := dataTypesSettingsPumpTest.RandomUnits(pointer.FromString("mmol/L"))
						mutator(datum)
						expectedDatum := dataTypesSettingsPumpTest.CloneUnits(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *dataTypesSettingsPump.Units) {},
					nil,
				),
				Entry("does not modify the datum; blood glucose missing",
					func(datum *dataTypesSettingsPump.Units) { datum.BloodGlucose = nil },
					nil,
				),
				Entry("does not modify the datum; blood glucose invalid",
					func(datum *dataTypesSettingsPump.Units) { datum.BloodGlucose = pointer.FromString("invalid") },
					nil,
				),
				Entry("does not modify the datum; blood glucose mmol/L",
					func(datum *dataTypesSettingsPump.Units) { datum.BloodGlucose = pointer.FromString("mmol/L") },
					nil,
				),
				Entry("does not modify the datum; blood glucose mmol/l",
					func(datum *dataTypesSettingsPump.Units) { datum.BloodGlucose = pointer.FromString("mmol/l") },
					nil,
				),
				Entry("does not modify the datum; blood glucose mg/dL",
					func(datum *dataTypesSettingsPump.Units) { datum.BloodGlucose = pointer.FromString("mg/dL") },
					nil,
				),
				Entry("does not modify the datum; blood glucose mg/dl",
					func(datum *dataTypesSettingsPump.Units) { datum.BloodGlucose = pointer.FromString("mg/dl") },
					nil,
				),
				Entry("does not modify the datum; carbohydrate missing",
					func(datum *dataTypesSettingsPump.Units) { datum.Carbohydrate = nil },
					nil,
				),
				Entry("does not modify the datum; carbohydrate invalid",
					func(datum *dataTypesSettingsPump.Units) { datum.Carbohydrate = pointer.FromString("invalid") },
					nil,
				),
				Entry("does not modify the datum; carbohydrate exchanges",
					func(datum *dataTypesSettingsPump.Units) { datum.Carbohydrate = pointer.FromString("exchanges") },
					nil,
				),
				Entry("does not modify the datum; carbohydrate grams",
					func(datum *dataTypesSettingsPump.Units) { datum.Carbohydrate = pointer.FromString("grams") },
					nil,
				),
				Entry("does not modify the datum; insulin missing",
					func(datum *dataTypesSettingsPump.Units) { datum.Insulin = nil },
					nil,
				),
				Entry("does not modify the datum; insulin invalid",
					func(datum *dataTypesSettingsPump.Units) { datum.Insulin = pointer.FromString("invalid") },
					nil,
				),
				Entry("does not modify the datum; insulin Units",
					func(datum *dataTypesSettingsPump.Units) { datum.Insulin = pointer.FromString("Units") },
					nil,
				),
			)
		})
	})
})
