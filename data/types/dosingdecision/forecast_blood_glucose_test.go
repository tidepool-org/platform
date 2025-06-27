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

var _ = Describe("ForecastBloodGlucose", func() {
	It("ForecastBloodGlucoseArrayLengthMaximum is expected", func() {
		Expect(dataTypesDosingDecision.ForecastBloodGlucoseArrayLengthMaximum).To(Equal(1440))
	})

	Context("ForecastBloodGlucose", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesDosingDecision.ForecastBloodGlucose)) {
				units := pointer.FromString(test.RandomStringFromArray(dataBloodGlucose.Units()))
				datum := dataTypesDosingDecisionTest.RandomForecastBloodGlucose(units)
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesDosingDecisionTest.NewObjectFromForecastBloodGlucose(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesDosingDecisionTest.NewObjectFromForecastBloodGlucose(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesDosingDecision.ForecastBloodGlucose) {},
			),
			Entry("empty",
				func(datum *dataTypesDosingDecision.ForecastBloodGlucose) {
					*datum = *dataTypesDosingDecision.NewForecastBloodGlucose()
				},
			),
			Entry("all",
				func(datum *dataTypesDosingDecision.ForecastBloodGlucose) {
					datum.Time = pointer.FromTime(test.RandomTime())
					datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(pointer.FromString(test.RandomStringFromArray(dataBloodGlucose.Units())))))
				},
			),
		)

		Context("ParseBloodGlucose", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesDosingDecision.ParseBloodGlucose(structureParser.NewObject(logTest.NewLogger(), nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				units := pointer.FromString(test.RandomStringFromArray(dataBloodGlucose.Units()))
				datum := dataTypesDosingDecisionTest.RandomBloodGlucose(units)
				object := dataTypesDosingDecisionTest.NewObjectFromBloodGlucose(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(logTest.NewLogger(), &object)
				Expect(dataTypesDosingDecision.ParseBloodGlucose(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewBloodGlucose", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesDosingDecision.NewBloodGlucose()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Time).To(BeNil())
				Expect(datum.Value).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.ForecastBloodGlucose), expectedErrors ...error) {
					units := pointer.FromString(test.RandomStringFromArray(dataBloodGlucose.Units()))
					expectedDatum := dataTypesDosingDecisionTest.RandomForecastBloodGlucose(units)
					object := dataTypesDosingDecisionTest.NewObjectFromForecastBloodGlucose(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesDosingDecision.NewForecastBloodGlucose()
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.ForecastBloodGlucose) {
					},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.ForecastBloodGlucose) {
						object["time"] = true
						object["value"] = true
						expectedDatum.Time = nil
						expectedDatum.Value = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/time"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotFloat64(true), "/value"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(units *string, mutator func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string), expectedErrors ...error) {
					datum := dataTypesDosingDecisionTest.RandomForecastBloodGlucose(units)
					mutator(datum, units)
					dataTypesTest.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithStringAdapter(datum, units), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string) {},
				),
				Entry("time missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string) {
						datum.Time = nil
					},
				),
				Entry("time exists",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string) {
						datum.Time = pointer.FromTime(test.RandomTime())
					},
				),
				Entry("value missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string) { datum.Value = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units mmol/L; value; out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string) {
						datum.Value = pointer.FromFloat64(-999999)
					},
				),
				Entry("units mmol/L; value; out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string) {
						datum.Value = pointer.FromFloat64(999999)
					},
				),
				Entry("units mmol/l; value; out of range (lower)",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string) {
						datum.Value = pointer.FromFloat64(-999999)
					},
				),
				Entry("units mmol/l; value; out of range (upper)",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string) {
						datum.Value = pointer.FromFloat64(999999)
					},
				),
				Entry("units mg/dL; value; out of range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string) {
						datum.Value = pointer.FromFloat64(-999999)
					},
				),
				Entry("units mg/dL; value; out of range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string) {
						datum.Value = pointer.FromFloat64(999999)
					},
				),
				Entry("units mg/dl; value; out of range (lower)",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string) {
						datum.Value = pointer.FromFloat64(-999999)
					},
				),
				Entry("units mg/dl; value; out of range (upper)",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string) {
						datum.Value = pointer.FromFloat64(999999)
					},
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string) {
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string), expectator func(datum *dataTypesDosingDecision.ForecastBloodGlucose, expectedDatum *dataTypesDosingDecision.ForecastBloodGlucose, units *string)) {
					datum := dataTypesDosingDecisionTest.RandomForecastBloodGlucose(units)
					mutator(datum, units)
					expectedDatum := dataTypesDosingDecisionTest.CloneForecastBloodGlucose(datum)
					normalizer := dataNormalizer.New(logTest.NewLogger())
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(structure.OriginExternal), units)
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					if expectator != nil {
						expectator(datum, expectedDatum, units)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string) {},
					func(datum *dataTypesDosingDecision.ForecastBloodGlucose, expectedDatum *dataTypesDosingDecision.ForecastBloodGlucose, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string) {},
					func(datum *dataTypesDosingDecision.ForecastBloodGlucose, expectedDatum *dataTypesDosingDecision.ForecastBloodGlucose, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string), expectator func(datum *dataTypesDosingDecision.ForecastBloodGlucose, expectedDatum *dataTypesDosingDecision.ForecastBloodGlucose, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := dataTypesDosingDecisionTest.RandomForecastBloodGlucose(units)
						mutator(datum, units)
						expectedDatum := dataTypesDosingDecisionTest.CloneForecastBloodGlucose(datum)
						normalizer := dataNormalizer.New(logTest.NewLogger())
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin), units)
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
					func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesDosingDecision.ForecastBloodGlucose, units *string) {},
					nil,
				),
			)
		})
	})

	Context("BloodGlucoseArray", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesDosingDecision.BloodGlucoseArray)) {
				units := pointer.FromString(test.RandomStringFromArray(dataBloodGlucose.Units()))
				datum := dataTypesDosingDecisionTest.RandomBloodGlucoseArray(units)
				mutator(datum)
				test.ExpectSerializedArrayJSON(dataTypesDosingDecisionTest.AnonymizeBloodGlucoseArray(datum), dataTypesDosingDecisionTest.NewArrayFromBloodGlucoseArray(datum, test.ObjectFormatJSON))
				test.ExpectSerializedArrayBSON(dataTypesDosingDecisionTest.AnonymizeBloodGlucoseArray(datum), dataTypesDosingDecisionTest.NewArrayFromBloodGlucoseArray(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesDosingDecision.BloodGlucoseArray) {},
			),
			Entry("empty",
				func(datum *dataTypesDosingDecision.BloodGlucoseArray) {
					*datum = *dataTypesDosingDecision.NewBloodGlucoseArray()
				},
			),
		)

		Context("ParseBloodGlucoseArray", func() {
			It("returns nil when the array is missing", func() {
				Expect(dataTypesDosingDecision.ParseBloodGlucoseArray(structureParser.NewArray(logTest.NewLogger(), nil))).To(BeNil())
			})

			It("returns new datum when the array is valid", func() {
				units := pointer.FromString(test.RandomStringFromArray(dataBloodGlucose.Units()))
				datum := dataTypesDosingDecisionTest.RandomBloodGlucoseArray(units)
				array := dataTypesDosingDecisionTest.NewArrayFromBloodGlucoseArray(datum, test.ObjectFormatJSON)
				parser := structureParser.NewArray(logTest.NewLogger(), &array)
				Expect(dataTypesDosingDecision.ParseBloodGlucoseArray(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewBloodGlucoseArray", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesDosingDecision.NewBloodGlucoseArray()
				Expect(datum).ToNot(BeNil())
				Expect(*datum).To(BeEmpty())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object []interface{}, expectedDatum *dataTypesDosingDecision.BloodGlucoseArray), expectedErrors ...error) {
					units := pointer.FromString(test.RandomStringFromArray(dataBloodGlucose.Units()))
					expectedDatum := dataTypesDosingDecisionTest.RandomBloodGlucoseArray(units)
					array := dataTypesDosingDecisionTest.NewArrayFromBloodGlucoseArray(expectedDatum, test.ObjectFormatJSON)
					mutator(array, expectedDatum)
					datum := dataTypesDosingDecision.NewBloodGlucoseArray()
					errorsTest.ExpectEqual(structureParser.NewArray(logTest.NewLogger(), &array).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object []interface{}, expectedDatum *dataTypesDosingDecision.BloodGlucoseArray) {},
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *dataTypesDosingDecision.BloodGlucoseArray, units *string), expectedErrors ...error) {
					datum := dataTypesDosingDecision.NewBloodGlucoseArray()
					mutator(datum, units)
					dataTypesTest.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithStringAdapter(datum, units), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.BloodGlucoseArray, units *string) {},
					structureValidator.ErrorValueEmpty(),
				),
				Entry("empty",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.BloodGlucoseArray, units *string) {
						*datum = *dataTypesDosingDecision.NewBloodGlucoseArray()
					},
					structureValidator.ErrorValueEmpty(),
				),
				Entry("nil",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.BloodGlucoseArray, units *string) { *datum = append(*datum, nil) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
				),
				Entry("single invalid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.BloodGlucoseArray, units *string) {
						invalid := dataTypesDosingDecisionTest.RandomBloodGlucose(units)
						invalid.Value = nil
						*datum = append(*datum, invalid)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0/value"),
				),
				Entry("single valid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.BloodGlucoseArray, units *string) {
						*datum = append(*datum, dataTypesDosingDecisionTest.RandomBloodGlucose(units))
					},
				),
				Entry("multiple invalid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.BloodGlucoseArray, units *string) {
						invalid := dataTypesDosingDecisionTest.RandomBloodGlucose(units)
						invalid.Value = nil
						*datum = append(*datum, dataTypesDosingDecisionTest.RandomBloodGlucose(units), invalid, dataTypesDosingDecisionTest.RandomBloodGlucose(units))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/value"),
				),
				Entry("multiple valid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.BloodGlucoseArray, units *string) {
						*datum = append(*datum, dataTypesDosingDecisionTest.RandomBloodGlucose(units), dataTypesDosingDecisionTest.RandomBloodGlucose(units), dataTypesDosingDecisionTest.RandomBloodGlucose(units))
					},
				),
				Entry("multiple; length in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.BloodGlucoseArray, units *string) {
						for len(*datum) < 1440 {
							*datum = append(*datum, dataTypesDosingDecisionTest.RandomBloodGlucose(units))
						}
					},
				),
				Entry("multiple; length out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.BloodGlucoseArray, units *string) {
						for len(*datum) < 1441 {
							*datum = append(*datum, dataTypesDosingDecisionTest.RandomBloodGlucose(units))
						}
					},
					structureValidator.ErrorLengthNotLessThanOrEqualTo(1441, 1440),
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.BloodGlucoseArray, units *string) {
						invalid := dataTypesDosingDecisionTest.RandomBloodGlucose(units)
						invalid.Value = nil
						*datum = append(*datum, nil, invalid, dataTypesDosingDecisionTest.RandomBloodGlucose(units))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/value"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *dataTypesDosingDecision.BloodGlucoseArray, units *string), expectator func(datum *dataTypesDosingDecision.BloodGlucoseArray, expectedDatum *dataTypesDosingDecision.BloodGlucoseArray, units *string)) {
					datum := dataTypesDosingDecisionTest.RandomBloodGlucoseArray(units)
					mutator(datum, units)
					expectedDatum := dataTypesDosingDecisionTest.CloneBloodGlucoseArray(datum)
					normalizer := dataNormalizer.New(logTest.NewLogger())
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(structure.OriginExternal), units)
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					if expectator != nil {
						expectator(datum, expectedDatum, units)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *dataTypesDosingDecision.BloodGlucoseArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *dataTypesDosingDecision.BloodGlucoseArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.BloodGlucoseArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesDosingDecision.BloodGlucoseArray, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesDosingDecision.BloodGlucoseArray, units *string) {},
					func(datum *dataTypesDosingDecision.BloodGlucoseArray, expectedDatum *dataTypesDosingDecision.BloodGlucoseArray, units *string) {
						for index := range *datum {
							dataBloodGlucoseTest.ExpectNormalizedValue((*datum)[index].Value, (*expectedDatum)[index].Value, units)
						}
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesDosingDecision.BloodGlucoseArray, units *string) {},
					func(datum *dataTypesDosingDecision.BloodGlucoseArray, expectedDatum *dataTypesDosingDecision.BloodGlucoseArray, units *string) {
						for index := range *datum {
							dataBloodGlucoseTest.ExpectNormalizedValue((*datum)[index].Value, (*expectedDatum)[index].Value, units)
						}
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *dataTypesDosingDecision.BloodGlucoseArray, units *string), expectator func(datum *dataTypesDosingDecision.BloodGlucoseArray, expectedDatum *dataTypesDosingDecision.BloodGlucoseArray, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := dataTypesDosingDecisionTest.RandomBloodGlucoseArray(units)
						mutator(datum, units)
						expectedDatum := dataTypesDosingDecisionTest.CloneBloodGlucoseArray(datum)
						normalizer := dataNormalizer.New(logTest.NewLogger())
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin), units)
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
					func(datum *dataTypesDosingDecision.BloodGlucoseArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *dataTypesDosingDecision.BloodGlucoseArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.BloodGlucoseArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesDosingDecision.BloodGlucoseArray, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesDosingDecision.BloodGlucoseArray, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesDosingDecision.BloodGlucoseArray, units *string) {},
					nil,
				),
			)
		})
	})
})
