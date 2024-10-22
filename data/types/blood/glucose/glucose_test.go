package glucose_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
	dataTypesBloodGlucoseTest "github.com/tidepool-org/platform/data/types/blood/glucose/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/metadata"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Glucose", func() {
	Context("New", func() {
		It("creates a new datum with all values initialized", func() {
			typ := dataTypesTest.NewType()
			datum := glucose.New(typ)
			Expect(datum.Type).To(Equal(typ))
			Expect(datum.Units).To(BeNil())
			Expect(datum.Value).To(BeNil())
		})
	})

	Context("with new datum", func() {
		var typ string
		var datum glucose.Glucose

		BeforeEach(func() {
			typ = dataTypesTest.NewType()
			datum = glucose.New(typ)
		})

		Context("Meta", func() {
			It("returns the meta with delivery type", func() {
				Expect(datum.Meta()).To(Equal(&types.Meta{Type: typ}))
			})
		})
	})

	Context("Validate", func() {
		DescribeTable("validates the datum",
			func(units *string, mutator func(datum *glucose.Glucose, units *string), expectedErrors ...error) {
				datum := dataTypesBloodGlucoseTest.NewGlucose(units)
				mutator(datum, units)
				dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
			},
			Entry("succeeds",
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) {},
			),
			Entry("type missing",
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) { datum.Type = "" },
				errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/type"),
			),
			Entry("type exists",
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) { datum.Type = dataTypesTest.NewType() },
			),
			Entry("units missing; value missing",
				nil,
				func(datum *glucose.Glucose, units *string) {
					datum.Value = nil
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
			),
			Entry("units missing; value out of range (lower)",
				nil,
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(-0.1)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
			),
			Entry("units missing; value in range (lower)",
				nil,
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(0.0)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
			),
			Entry("units missing; value in range (upper)",
				nil,
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(55.0)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
			),
			Entry("units missing; value out of range (upper)",
				nil,
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(1000.1)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
			),
			Entry("units invalid; value missing",
				pointer.FromString("invalid"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = nil
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
			),
			Entry("units invalid; value out of range (lower)",
				pointer.FromString("invalid"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(-0.1)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units"),
			),
			Entry("units invalid; value in range (lower)",
				pointer.FromString("invalid"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(0.0)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units"),
			),
			Entry("units invalid; value in range (upper)",
				pointer.FromString("invalid"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(55.0)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units"),
			),
			Entry("units invalid; value out of range (upper)",
				pointer.FromString("invalid"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(1000.1)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units"),
			),
			Entry("units mmol/L; value missing",
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = nil
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
			),
			Entry("units mmol/L; value out of range (lower)",
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(-0.1)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/value"),
			),
			Entry("units mmol/L; value in range (lower)",
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(0.0)
				},
			),
			Entry("units mmol/L; value in range (upper)",
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(55.0)
				},
			),
			Entry("units mmol/L; value out of range (upper)",
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(55.1)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/value"),
			),
			Entry("units mmol/l; value missing",
				pointer.FromString("mmol/l"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = nil
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
			),
			Entry("units mmol/l; value out of range (lower)",
				pointer.FromString("mmol/l"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(-0.1)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/value"),
			),
			Entry("units mmol/l; value in range (lower)",
				pointer.FromString("mmol/l"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(0.0)
				},
			),
			Entry("units mmol/l; value in range (upper)",
				pointer.FromString("mmol/l"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(55.0)
				},
			),
			Entry("units mmol/l; value out of range (upper)",
				pointer.FromString("mmol/l"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(55.1)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/value"),
			),
			Entry("units mg/dL; value missing",
				pointer.FromString("mg/dL"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = nil
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
			),
			Entry("units mg/dL; value out of range (lower)",
				pointer.FromString("mg/dL"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(-0.1)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/value"),
			),
			Entry("units mg/dL; value in range (lower)",
				pointer.FromString("mg/dL"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(0.0)
				},
			),
			Entry("units mg/dL; value in range (upper)",
				pointer.FromString("mg/dL"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(1000.0)
				},
			),
			Entry("units mg/dL; value out of range (upper)",
				pointer.FromString("mg/dL"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(1000.1)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/value"),
			),
			Entry("units mg/dl; value missing",
				pointer.FromString("mg/dl"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = nil
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
			),
			Entry("units mg/dl; value out of range (lower)",
				pointer.FromString("mg/dl"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(-0.1)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/value"),
			),
			Entry("units mg/dl; value in range (lower)",
				pointer.FromString("mg/dl"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(0.0)
				},
			),
			Entry("units mg/dl; value in range (upper)",
				pointer.FromString("mg/dl"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(1000.0)
				},
			),
			Entry("units mg/dl; value out of range (upper)",
				pointer.FromString("mg/dl"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(1000.1)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/value"),
			),
			Entry("multiple errors",
				nil,
				func(datum *glucose.Glucose, units *string) {
					datum.Type = ""
					datum.Value = nil
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/type"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
			),
		)
	})

	Context("Normalize", func() {
		DescribeTable("normalizes the datum",
			func(units *string, mutator func(datum *glucose.Glucose, units *string), expectator func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string)) {

				for _, origin := range structure.Origins() {
					datum := dataTypesBloodGlucoseTest.NewGlucose(units)
					mutator(datum, units)
					expectedDatum := dataTypesBloodGlucoseTest.CloneGlucose(datum)
					normalizer := dataNormalizer.New()
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(origin))
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					expectedDatum.Raw = metadataTest.CloneMetadata(datum.Raw)
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
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units mmol/L; value missing",
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				nil,
			),
		)

		DescribeTable("normalizes the datum with origin external",
			func(units *string, mutator func(datum *glucose.Glucose, units *string), expectator func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string, value *float64)) {
				datum := dataTypesBloodGlucoseTest.NewGlucose(units)
				mutator(datum, units)
				originalValue := pointer.CloneFloat64(datum.Value)
				expectedDatum := dataTypesBloodGlucoseTest.CloneGlucose(datum)
				normalizer := dataNormalizer.New()
				Expect(normalizer).ToNot(BeNil())
				datum.Normalize(normalizer.WithOrigin(structure.OriginExternal))
				Expect(normalizer.Error()).To(BeNil())
				Expect(normalizer.Data()).To(BeEmpty())
				expectedDatum.Raw = metadataTest.CloneMetadata(datum.Raw)
				if expectator != nil {
					expectator(datum, expectedDatum, units, originalValue)
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
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units mmol/L; value missing",
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				nil,
			),
			Entry("modifies the datum; units mmol/l",
				pointer.FromString("mmol/l"),
				func(datum *glucose.Glucose, units *string) {},
				func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string, value *float64) {
					dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					dataBloodGlucoseTest.ExpectRaw(datum.Raw, &metadata.Metadata{"units": *units, "value": *value})
				},
			),
			Entry("modifies the datum; units mmol/l; value missing",
				pointer.FromString("mmol/l"),
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string, value *float64) {
					dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					dataBloodGlucoseTest.ExpectRaw(datum.Raw, &metadata.Metadata{"units": *units, "value": nil})
				},
			),
			Entry("modifies the datum; units mg/dL",
				pointer.FromString("mg/dL"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(units)))
				},
				func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string, value *float64) {
					dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					dataBloodGlucoseTest.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
					dataBloodGlucoseTest.ExpectRaw(datum.Raw, &metadata.Metadata{"units": *units, "value": *value})
				},
			),
			Entry("modifies the datum; units mg/dL; value missing",
				pointer.FromString("mg/dL"),
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string, value *float64) {
					dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					dataBloodGlucoseTest.ExpectRaw(datum.Raw, &metadata.Metadata{"units": *units, "value": nil})
				},
			),
			Entry("modifies the datum; units mg/dl",
				pointer.FromString("mg/dl"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(units)))
				},
				func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string, value *float64) {
					dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					dataBloodGlucoseTest.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
					dataBloodGlucoseTest.ExpectRaw(datum.Raw, &metadata.Metadata{"units": *units, "value": *value})
				},
			),
			Entry("modifies the datum; units mg/dl; value missing",
				pointer.FromString("mg/dl"),
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string, value *float64) {
					dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					dataBloodGlucoseTest.ExpectRaw(datum.Raw, &metadata.Metadata{"units": *units, "value": nil})
				},
			),
		)

		DescribeTable("normalizes the datum with origin internal/store",
			func(units *string, mutator func(datum *glucose.Glucose, units *string), expectator func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string)) {
				for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
					datum := dataTypesBloodGlucoseTest.NewGlucose(units)
					mutator(datum, units)
					expectedDatum := dataTypesBloodGlucoseTest.CloneGlucose(datum)
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
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units mmol/L; value missing",
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units mmol/l",
				pointer.FromString("mmol/l"),
				func(datum *glucose.Glucose, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units mmol/l; value missing",
				pointer.FromString("mmol/l"),
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units mg/dL",
				pointer.FromString("mg/dL"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(units)))
				},
				nil,
			),
			Entry("does not modify the datum; units mg/dL; value missing",
				pointer.FromString("mg/dL"),
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units mg/dl",
				pointer.FromString("mg/dl"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(units)))
				},
				nil,
			),
			Entry("does not modify the datum; units mg/dl; value missing",
				pointer.FromString("mg/dl"),
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				nil,
			),
		)
	})
})
