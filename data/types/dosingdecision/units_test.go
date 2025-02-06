package dosingdecision_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypesDosingDecision "github.com/tidepool-org/platform/data/types/dosingdecision"
	dataTypesDosingDecisionTest "github.com/tidepool-org/platform/data/types/dosingdecision/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Units", func() {
	It("CarbohydrateUnitsExchanges is expected", func() {
		Expect(dataTypesDosingDecision.CarbohydrateUnitsExchanges).To(Equal("exchanges"))
	})

	It("CarbohydrateUnitsGrams is expected", func() {
		Expect(dataTypesDosingDecision.CarbohydrateUnitsGrams).To(Equal("grams"))
	})

	It("InsulinUnitsUnits is expected", func() {
		Expect(dataTypesDosingDecision.InsulinUnitsUnits).To(Equal("Units"))
	})

	It("CarbohydrateUnits returns expected", func() {
		Expect(dataTypesDosingDecision.CarbohydrateUnits()).To(Equal([]string{"exchanges", "grams"}))
	})

	It("InsulinUnits returns expected", func() {
		Expect(dataTypesDosingDecision.InsulinUnits()).To(Equal([]string{"Units"}))
	})

	Context("Units", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesDosingDecision.Units)) {
				datum := dataTypesDosingDecisionTest.RandomUnits(pointer.FromString("mg/dL"))
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesDosingDecisionTest.NewObjectFromUnits(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesDosingDecisionTest.NewObjectFromUnits(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesDosingDecision.Units) {},
			),
			Entry("empty",
				func(datum *dataTypesDosingDecision.Units) {
					*datum = *dataTypesDosingDecision.NewUnits()
				},
			),
			Entry("all",
				func(datum *dataTypesDosingDecision.Units) {
					datum.BloodGlucose = pointer.FromString(test.RandomStringFromArray(dataBloodGlucose.Units()))
					datum.Carbohydrate = pointer.FromString(test.RandomStringFromArray(dataTypesDosingDecision.CarbohydrateUnits()))
					datum.Insulin = pointer.FromString(test.RandomStringFromArray(dataTypesDosingDecision.InsulinUnits()))
				},
			),
		)

		Context("ParseUnits", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesDosingDecision.ParseUnits(structureParser.NewObject(logTest.NewLogger(), nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesDosingDecisionTest.RandomUnits(pointer.FromString("mg/dL"))
				object := dataTypesDosingDecisionTest.NewObjectFromUnits(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(logTest.NewLogger(), &object)
				Expect(dataTypesDosingDecision.ParseUnits(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewUnits", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesDosingDecision.NewUnits()
				Expect(datum).ToNot(BeNil())
				Expect(datum.BloodGlucose).To(BeNil())
				Expect(datum.Carbohydrate).To(BeNil())
				Expect(datum.Insulin).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.Units), expectedErrors ...error) {
					expectedDatum := dataTypesDosingDecisionTest.RandomUnits(pointer.FromString("mg/dL"))
					object := dataTypesDosingDecisionTest.NewObjectFromUnits(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesDosingDecision.NewUnits()
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.Units) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.Units) {
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
				func(mutator func(datum *dataTypesDosingDecision.Units), expectedErrors ...error) {
					datum := dataTypesDosingDecisionTest.RandomUnits(pointer.FromString("mmol/L"))
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesDosingDecision.Units) {},
				),
				Entry("blood glucose missing",
					func(datum *dataTypesDosingDecision.Units) { datum.BloodGlucose = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/bg"),
				),
				Entry("blood glucose invalid",
					func(datum *dataTypesDosingDecision.Units) { datum.BloodGlucose = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/bg"),
				),
				Entry("blood glucose mmol/L",
					func(datum *dataTypesDosingDecision.Units) { datum.BloodGlucose = pointer.FromString("mmol/L") },
				),
				Entry("blood glucose mmol/l",
					func(datum *dataTypesDosingDecision.Units) { datum.BloodGlucose = pointer.FromString("mmol/l") },
				),
				Entry("blood glucose mg/dL",
					func(datum *dataTypesDosingDecision.Units) { datum.BloodGlucose = pointer.FromString("mg/dL") },
				),
				Entry("blood glucose mg/dl",
					func(datum *dataTypesDosingDecision.Units) { datum.BloodGlucose = pointer.FromString("mg/dl") },
				),
				Entry("carbohydrate missing",
					func(datum *dataTypesDosingDecision.Units) { datum.Carbohydrate = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/carb"),
				),
				Entry("carbohydrate invalid",
					func(datum *dataTypesDosingDecision.Units) { datum.Carbohydrate = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"exchanges", "grams"}), "/carb"),
				),
				Entry("carbohydrate exchanges",
					func(datum *dataTypesDosingDecision.Units) { datum.Carbohydrate = pointer.FromString("exchanges") },
				),
				Entry("carbohydrate grams",
					func(datum *dataTypesDosingDecision.Units) { datum.Carbohydrate = pointer.FromString("grams") },
				),
				Entry("insulin missing",
					func(datum *dataTypesDosingDecision.Units) { datum.Insulin = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/insulin"),
				),
				Entry("insulin invalid",
					func(datum *dataTypesDosingDecision.Units) { datum.Insulin = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"Units"}), "/insulin"),
				),
				Entry("insulin Units",
					func(datum *dataTypesDosingDecision.Units) { datum.Insulin = pointer.FromString("Units") },
				),
				Entry("multiple errors",
					func(datum *dataTypesDosingDecision.Units) {
						datum.BloodGlucose = nil
						datum.Carbohydrate = nil
						datum.Insulin = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/bg"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/carb"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/insulin"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum with origin external",
				func(mutator func(datum *dataTypesDosingDecision.Units), expectator func(datum *dataTypesDosingDecision.Units, expectedDatum *dataTypesDosingDecision.Units)) {
					datum := dataTypesDosingDecisionTest.RandomUnits(pointer.FromString("mmol/L"))
					mutator(datum)
					expectedDatum := dataTypesDosingDecisionTest.CloneUnits(datum)
					normalizer := dataNormalizer.New(logTest.NewLogger())
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
					func(datum *dataTypesDosingDecision.Units) {},
					nil,
				),
				Entry("does not modify the datum; blood glucose missing",
					func(datum *dataTypesDosingDecision.Units) { datum.BloodGlucose = nil },
					nil,
				),
				Entry("does not modify the datum; blood glucose invalid",
					func(datum *dataTypesDosingDecision.Units) { datum.BloodGlucose = pointer.FromString("invalid") },
					nil,
				),
				Entry("does not modify the datum; blood glucose mmol/L",
					func(datum *dataTypesDosingDecision.Units) { datum.BloodGlucose = pointer.FromString("mmol/L") },
					nil,
				),
				Entry("modifies the datum; blood glucose mmol/l",
					func(datum *dataTypesDosingDecision.Units) { datum.BloodGlucose = pointer.FromString("mmol/l") },
					func(datum *dataTypesDosingDecision.Units, expectedDatum *dataTypesDosingDecision.Units) {
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.BloodGlucose, expectedDatum.BloodGlucose)
					},
				),
				Entry("modifies the datum; blood glucose mg/dL",
					func(datum *dataTypesDosingDecision.Units) { datum.BloodGlucose = pointer.FromString("mg/dL") },
					func(datum *dataTypesDosingDecision.Units, expectedDatum *dataTypesDosingDecision.Units) {
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.BloodGlucose, expectedDatum.BloodGlucose)
					},
				),
				Entry("modifies the datum; blood glucose mg/dl",
					func(datum *dataTypesDosingDecision.Units) { datum.BloodGlucose = pointer.FromString("mg/dl") },
					func(datum *dataTypesDosingDecision.Units, expectedDatum *dataTypesDosingDecision.Units) {
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.BloodGlucose, expectedDatum.BloodGlucose)
					},
				),
				Entry("does not modify the datum; carbohydrate missing",
					func(datum *dataTypesDosingDecision.Units) { datum.Carbohydrate = nil },
					nil,
				),
				Entry("does not modify the datum; carbohydrate invalid",
					func(datum *dataTypesDosingDecision.Units) { datum.Carbohydrate = pointer.FromString("invalid") },
					nil,
				),
				Entry("does not modify the datum; carbohydrate exchanges",
					func(datum *dataTypesDosingDecision.Units) { datum.Carbohydrate = pointer.FromString("exchanges") },
					nil,
				),
				Entry("does not modify the datum; carbohydrate grams",
					func(datum *dataTypesDosingDecision.Units) { datum.Carbohydrate = pointer.FromString("grams") },
					nil,
				),
				Entry("does not modify the datum; insulin missing",
					func(datum *dataTypesDosingDecision.Units) { datum.Insulin = nil },
					nil,
				),
				Entry("does not modify the datum; insulin invalid",
					func(datum *dataTypesDosingDecision.Units) { datum.Insulin = pointer.FromString("invalid") },
					nil,
				),
				Entry("does not modify the datum; insulin Units",
					func(datum *dataTypesDosingDecision.Units) { datum.Insulin = pointer.FromString("Units") },
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(mutator func(datum *dataTypesDosingDecision.Units), expectator func(datum *dataTypesDosingDecision.Units, expectedDatum *dataTypesDosingDecision.Units)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := dataTypesDosingDecisionTest.RandomUnits(pointer.FromString("mmol/L"))
						mutator(datum)
						expectedDatum := dataTypesDosingDecisionTest.CloneUnits(datum)
						normalizer := dataNormalizer.New(logTest.NewLogger())
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
					func(datum *dataTypesDosingDecision.Units) {},
					nil,
				),
				Entry("does not modify the datum; blood glucose missing",
					func(datum *dataTypesDosingDecision.Units) { datum.BloodGlucose = nil },
					nil,
				),
				Entry("does not modify the datum; blood glucose invalid",
					func(datum *dataTypesDosingDecision.Units) { datum.BloodGlucose = pointer.FromString("invalid") },
					nil,
				),
				Entry("does not modify the datum; blood glucose mmol/L",
					func(datum *dataTypesDosingDecision.Units) { datum.BloodGlucose = pointer.FromString("mmol/L") },
					nil,
				),
				Entry("does not modify the datum; blood glucose mmol/l",
					func(datum *dataTypesDosingDecision.Units) { datum.BloodGlucose = pointer.FromString("mmol/l") },
					nil,
				),
				Entry("does not modify the datum; blood glucose mg/dL",
					func(datum *dataTypesDosingDecision.Units) { datum.BloodGlucose = pointer.FromString("mg/dL") },
					nil,
				),
				Entry("does not modify the datum; blood glucose mg/dl",
					func(datum *dataTypesDosingDecision.Units) { datum.BloodGlucose = pointer.FromString("mg/dl") },
					nil,
				),
				Entry("does not modify the datum; carbohydrate missing",
					func(datum *dataTypesDosingDecision.Units) { datum.Carbohydrate = nil },
					nil,
				),
				Entry("does not modify the datum; carbohydrate invalid",
					func(datum *dataTypesDosingDecision.Units) { datum.Carbohydrate = pointer.FromString("invalid") },
					nil,
				),
				Entry("does not modify the datum; carbohydrate exchanges",
					func(datum *dataTypesDosingDecision.Units) { datum.Carbohydrate = pointer.FromString("exchanges") },
					nil,
				),
				Entry("does not modify the datum; carbohydrate grams",
					func(datum *dataTypesDosingDecision.Units) { datum.Carbohydrate = pointer.FromString("grams") },
					nil,
				),
				Entry("does not modify the datum; insulin missing",
					func(datum *dataTypesDosingDecision.Units) { datum.Insulin = nil },
					nil,
				),
				Entry("does not modify the datum; insulin invalid",
					func(datum *dataTypesDosingDecision.Units) { datum.Insulin = pointer.FromString("invalid") },
					nil,
				),
				Entry("does not modify the datum; insulin Units",
					func(datum *dataTypesDosingDecision.Units) { datum.Insulin = pointer.FromString("Units") },
					nil,
				),
			)
		})
	})
})
