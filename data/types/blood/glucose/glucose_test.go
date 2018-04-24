package glucose_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	testDataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
	testDataTypesBloodGlucose "github.com/tidepool-org/platform/data/types/blood/glucose/test"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Glucose", func() {
	Context("Validate", func() {
		DescribeTable("validates the datum",
			func(units *string, mutator func(datum *glucose.Glucose, units *string), expectedErrors ...error) {
				datum := testDataTypesBloodGlucose.NewGlucose(units)
				mutator(datum, units)
				testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
			},
			Entry("succeeds",
				pointer.String("mmol/L"),
				func(datum *glucose.Glucose, units *string) {},
			),
			Entry("type missing",
				pointer.String("mmol/L"),
				func(datum *glucose.Glucose, units *string) { datum.Type = "" },
				testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/type"),
			),
			Entry("type exists",
				pointer.String("mmol/L"),
				func(datum *glucose.Glucose, units *string) { datum.Type = testDataTypes.NewType() },
			),
			Entry("units missing; value missing",
				nil,
				func(datum *glucose.Glucose, units *string) {
					datum.Value = nil
				},
				testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
			),
			Entry("units missing; value out of range (lower)",
				nil,
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(-0.1)
				},
				testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
			),
			Entry("units missing; value in range (lower)",
				nil,
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(0.0)
				},
				testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
			),
			Entry("units missing; value in range (upper)",
				nil,
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(55.0)
				},
				testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
			),
			Entry("units missing; value out of range (upper)",
				nil,
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(1000.1)
				},
				testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
			),
			Entry("units invalid; value missing",
				pointer.String("invalid"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = nil
				},
				testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units"),
				testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
			),
			Entry("units invalid; value out of range (lower)",
				pointer.String("invalid"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(-0.1)
				},
				testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units"),
			),
			Entry("units invalid; value in range (lower)",
				pointer.String("invalid"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(0.0)
				},
				testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units"),
			),
			Entry("units invalid; value in range (upper)",
				pointer.String("invalid"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(55.0)
				},
				testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units"),
			),
			Entry("units invalid; value out of range (upper)",
				pointer.String("invalid"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(1000.1)
				},
				testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units"),
			),
			Entry("units mmol/L; value missing",
				pointer.String("mmol/L"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = nil
				},
				testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
			),
			Entry("units mmol/L; value out of range (lower)",
				pointer.String("mmol/L"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(-0.1)
				},
				testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/value"),
			),
			Entry("units mmol/L; value in range (lower)",
				pointer.String("mmol/L"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(0.0)
				},
			),
			Entry("units mmol/L; value in range (upper)",
				pointer.String("mmol/L"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(55.0)
				},
			),
			Entry("units mmol/L; value out of range (upper)",
				pointer.String("mmol/L"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(55.1)
				},
				testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/value"),
			),
			Entry("units mmol/l; value missing",
				pointer.String("mmol/l"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = nil
				},
				testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
			),
			Entry("units mmol/l; value out of range (lower)",
				pointer.String("mmol/l"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(-0.1)
				},
				testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/value"),
			),
			Entry("units mmol/l; value in range (lower)",
				pointer.String("mmol/l"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(0.0)
				},
			),
			Entry("units mmol/l; value in range (upper)",
				pointer.String("mmol/l"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(55.0)
				},
			),
			Entry("units mmol/l; value out of range (upper)",
				pointer.String("mmol/l"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(55.1)
				},
				testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/value"),
			),
			Entry("units mg/dL; value missing",
				pointer.String("mg/dL"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = nil
				},
				testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
			),
			Entry("units mg/dL; value out of range (lower)",
				pointer.String("mg/dL"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(-0.1)
				},
				testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/value"),
			),
			Entry("units mg/dL; value in range (lower)",
				pointer.String("mg/dL"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(0.0)
				},
			),
			Entry("units mg/dL; value in range (upper)",
				pointer.String("mg/dL"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(1000.0)
				},
			),
			Entry("units mg/dL; value out of range (upper)",
				pointer.String("mg/dL"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(1000.1)
				},
				testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/value"),
			),
			Entry("units mg/dl; value missing",
				pointer.String("mg/dl"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = nil
				},
				testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
			),
			Entry("units mg/dl; value out of range (lower)",
				pointer.String("mg/dl"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(-0.1)
				},
				testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/value"),
			),
			Entry("units mg/dl; value in range (lower)",
				pointer.String("mg/dl"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(0.0)
				},
			),
			Entry("units mg/dl; value in range (upper)",
				pointer.String("mg/dl"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(1000.0)
				},
			),
			Entry("units mg/dl; value out of range (upper)",
				pointer.String("mg/dl"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(1000.1)
				},
				testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/value"),
			),
			Entry("multiple errors",
				nil,
				func(datum *glucose.Glucose, units *string) {
					datum.Type = ""
					datum.Value = nil
				},
				testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/type"),
				testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
			),
		)
	})

	Context("Normalize", func() {
		DescribeTable("normalizes the datum",
			func(units *string, mutator func(datum *glucose.Glucose, units *string), expectator func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string)) {
				for _, origin := range structure.Origins() {
					datum := testDataTypesBloodGlucose.NewGlucose(units)
					mutator(datum, units)
					expectedDatum := testDataTypesBloodGlucose.CloneGlucose(datum)
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
				func(datum *glucose.Glucose, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units missing, value missing",
				nil,
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units mmol/L",
				pointer.String("mmol/L"),
				func(datum *glucose.Glucose, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units mmol/L; value missing",
				pointer.String("mmol/L"),
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				nil,
			),
		)

		DescribeTable("normalizes the datum with origin external",
			func(units *string, mutator func(datum *glucose.Glucose, units *string), expectator func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string)) {
				datum := testDataTypesBloodGlucose.NewGlucose(units)
				mutator(datum, units)
				expectedDatum := testDataTypesBloodGlucose.CloneGlucose(datum)
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
			Entry("does not modify the datum; units missing",
				nil,
				func(datum *glucose.Glucose, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units missing, value missing",
				nil,
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units mmol/L",
				pointer.String("mmol/L"),
				func(datum *glucose.Glucose, units *string) {},
				func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string) {
					testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					testDataBloodGlucose.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
				},
			),
			Entry("does not modify the datum; units mmol/L; value missing",
				pointer.String("mmol/L"),
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string) {
					testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
				},
			),
			Entry("modifies the datum; units mmol/l",
				pointer.String("mmol/l"),
				func(datum *glucose.Glucose, units *string) {},
				func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string) {
					testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					testDataBloodGlucose.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
				},
			),
			Entry("modifies the datum; units mmol/l; value missing",
				pointer.String("mmol/l"),
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string) {
					testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
				},
			),
			Entry("modifies the datum; units mg/dL",
				pointer.String("mg/dL"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(units)))
				},
				func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string) {
					testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					testDataBloodGlucose.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
				},
			),
			Entry("modifies the datum; units mg/dL; value missing",
				pointer.String("mg/dL"),
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string) {
					testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
				},
			),
			Entry("modifies the datum; units mg/dl",
				pointer.String("mg/dl"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(units)))
				},
				func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string) {
					testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					testDataBloodGlucose.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
				},
			),
			Entry("modifies the datum; units mg/dl; value missing",
				pointer.String("mg/dl"),
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string) {
					testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
				},
			),
		)

		DescribeTable("normalizes the datum with origin internal/store",
			func(units *string, mutator func(datum *glucose.Glucose, units *string), expectator func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string)) {
				for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
					datum := testDataTypesBloodGlucose.NewGlucose(units)
					mutator(datum, units)
					expectedDatum := testDataTypesBloodGlucose.CloneGlucose(datum)
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
				func(datum *glucose.Glucose, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units missing, value missing",
				nil,
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units mmol/L",
				pointer.String("mmol/L"),
				func(datum *glucose.Glucose, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units mmol/L; value missing",
				pointer.String("mmol/L"),
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units mmol/l",
				pointer.String("mmol/l"),
				func(datum *glucose.Glucose, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units mmol/l; value missing",
				pointer.String("mmol/l"),
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units mg/dL",
				pointer.String("mg/dL"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(units)))
				},
				nil,
			),
			Entry("does not modify the datum; units mg/dL; value missing",
				pointer.String("mg/dL"),
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units mg/dl",
				pointer.String("mg/dl"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.Float64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(units)))
				},
				nil,
			),
			Entry("does not modify the datum; units mg/dl; value missing",
				pointer.String("mg/dl"),
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				nil,
			),
		)
	})
})
