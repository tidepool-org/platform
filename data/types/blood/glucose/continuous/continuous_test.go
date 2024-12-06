package continuous_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	dataTypesBloodGlucoseTest "github.com/tidepool-org/platform/data/types/blood/glucose/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/metadata"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewMeta() interface{} {
	return &types.Meta{
		Type: "cbg",
	}
}

func NewContinuous(units *string) *continuous.Continuous {
	datum := continuous.New()
	datum.Glucose = *dataTypesBloodGlucoseTest.NewGlucose(units)
	datum.Type = "cbg"
	return datum
}

func CloneContinuous(datum *continuous.Continuous) *continuous.Continuous {
	if datum == nil {
		return nil
	}
	clone := continuous.New()
	clone.Glucose = *dataTypesBloodGlucoseTest.CloneGlucose(&datum.Glucose)
	clone.Trend = pointer.CloneString(datum.Trend)
	clone.TrendRate = pointer.CloneFloat64(datum.TrendRate)
	return clone
}

var _ = Describe("Continuous", func() {
	It("Type is expected", func() {
		Expect(continuous.Type).To(Equal("cbg"))
	})

	Context("New", func() {
		It("returns the expected datum", func() {
			datum := continuous.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("cbg"))
			Expect(datum.Units).To(BeNil())
			Expect(datum.Value).To(BeNil())
		})
	})

	Context("Validate", func() {
		DescribeTable("validates the datum",
			func(units *string, mutator func(datum *continuous.Continuous, units *string), expectedErrors ...error) {
				datum := NewContinuous(units)
				mutator(datum, units)
				dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
			},
			Entry("succeeds",
				pointer.FromString("mmol/L"),
				func(datum *continuous.Continuous, units *string) {},
			),
			Entry("type missing",
				pointer.FromString("mmol/L"),
				func(datum *continuous.Continuous, units *string) { datum.Type = "" },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &types.Meta{}),
			),
			Entry("type invalid",
				pointer.FromString("mmol/L"),
				func(datum *continuous.Continuous, units *string) { datum.Type = "invalidType" },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "cbg"), "/type", &types.Meta{Type: "invalidType"}),
			),
			Entry("type cbg",
				pointer.FromString("mmol/L"),
				func(datum *continuous.Continuous, units *string) { datum.Type = "cbg" },
			),
			Entry("units missing; value missing",
				nil,
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
			),
			Entry("units missing; value out of range (lower)",
				nil,
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
			),
			Entry("units missing; value in range (lower)",
				nil,
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.FromFloat64(0.0) },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
			),
			Entry("units missing; value in range (upper)",
				nil,
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.FromFloat64(55.0) },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
			),
			Entry("units missing; value out of range (upper)",
				nil,
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.FromFloat64(1000.1) },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
			),
			Entry("units invalid; value missing",
				pointer.FromString("invalid"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
			),
			Entry("units invalid; value out of range (lower)",
				pointer.FromString("invalid"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
			),
			Entry("units invalid; value in range (lower)",
				pointer.FromString("invalid"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.FromFloat64(0.0) },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
			),
			Entry("units invalid; value in range (upper)",
				pointer.FromString("invalid"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.FromFloat64(55.0) },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
			),
			Entry("units invalid; value out of range (upper)",
				pointer.FromString("invalid"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.FromFloat64(1000.1) },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
			),
			Entry("units mmol/L; value missing",
				pointer.FromString("mmol/L"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
			),
			Entry("units mmol/L; value out of range (lower)",
				pointer.FromString("mmol/L"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/value", NewMeta()),
			),
			Entry("units mmol/L; value in range (lower)",
				pointer.FromString("mmol/L"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.FromFloat64(0.0) },
			),
			Entry("units mmol/L; value in range (upper)",
				pointer.FromString("mmol/L"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.FromFloat64(55.0) },
			),
			Entry("units mmol/L; value out of range (upper)",
				pointer.FromString("mmol/L"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.FromFloat64(55.1) },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/value", NewMeta()),
			),
			Entry("units mmol/l; value missing",
				pointer.FromString("mmol/l"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
			),
			Entry("units mmol/l; value out of range (lower)",
				pointer.FromString("mmol/l"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/value", NewMeta()),
			),
			Entry("units mmol/l; value in range (lower)",
				pointer.FromString("mmol/l"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.FromFloat64(0.0) },
			),
			Entry("units mmol/l; value in range (upper)",
				pointer.FromString("mmol/l"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.FromFloat64(55.0) },
			),
			Entry("units mmol/l; value out of range (upper)",
				pointer.FromString("mmol/l"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.FromFloat64(55.1) },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/value", NewMeta()),
			),
			Entry("units mg/dL; value missing",
				pointer.FromString("mg/dL"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
			),
			Entry("units mg/dL; value out of range (lower)",
				pointer.FromString("mg/dL"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/value", NewMeta()),
			),
			Entry("units mg/dL; value in range (lower)",
				pointer.FromString("mg/dL"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.FromFloat64(0.0) },
			),
			Entry("units mg/dL; value in range (upper)",
				pointer.FromString("mg/dL"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.FromFloat64(1000.0) },
			),
			Entry("units mg/dL; value out of range (upper)",
				pointer.FromString("mg/dL"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.FromFloat64(1000.1) },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/value", NewMeta()),
			),
			Entry("units mg/dl; value missing",
				pointer.FromString("mg/dl"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
			),
			Entry("units mg/dl; value out of range (lower)",
				pointer.FromString("mg/dl"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/value", NewMeta()),
			),
			Entry("units mg/dl; value in range (lower)",
				pointer.FromString("mg/dl"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.FromFloat64(0.0) },
			),
			Entry("units mg/dl; value in range (upper)",
				pointer.FromString("mg/dl"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.FromFloat64(1000.0) },
			),
			Entry("units mg/dl; value out of range (upper)",
				pointer.FromString("mg/dl"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.FromFloat64(1000.1) },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/value", NewMeta()),
			),
			Entry("trend rate missing",
				pointer.FromString("mg/dl"),
				func(datum *continuous.Continuous, units *string) { datum.TrendRate = nil },
			),
			Entry("trend rate out of range (lower)",
				pointer.FromString("mg/dl"),
				func(datum *continuous.Continuous, units *string) { datum.TrendRate = pointer.FromFloat64(-100.1) },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-100.1, -100.0, 100.0), "/trendRate", NewMeta()),
			),
			Entry("trend rate in range (lower)",
				pointer.FromString("mg/dl"),
				func(datum *continuous.Continuous, units *string) { datum.TrendRate = pointer.FromFloat64(-100.0) },
			),
			Entry("trend rate in range (upper)",
				pointer.FromString("mg/dl"),
				func(datum *continuous.Continuous, units *string) { datum.TrendRate = pointer.FromFloat64(100.0) },
			),
			Entry("trend rate out of range (upper)",
				pointer.FromString("mg/dl"),
				func(datum *continuous.Continuous, units *string) { datum.TrendRate = pointer.FromFloat64(100.1) },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, -100.0, 100.0), "/trendRate", NewMeta()),
			),
			Entry("multiple errors",
				nil,
				func(datum *continuous.Continuous, units *string) {
					datum.Type = ""
					datum.Value = nil
					datum.Trend = nil
					datum.TrendRate = pointer.FromFloat64(-100.1)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &types.Meta{}),
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", &types.Meta{}),
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", &types.Meta{}),
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-100.1, -100.0, 100.0), "/trendRate", &types.Meta{}),
			),
		)
	})

	Context("Normalize", func() {
		DescribeTable("normalizes the datum",
			func(units *string, mutator func(datum *continuous.Continuous, units *string), expectator func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string)) {

				for _, origin := range structure.Origins() {
					datum := NewContinuous(units)
					mutator(datum, units)
					expectedDatum := CloneContinuous(datum)
					normalizer := dataNormalizer.New(logTest.NewLogger())
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
				func(datum *continuous.Continuous, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units missing; value missing",
				nil,
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units invalid",
				pointer.FromString("invalid"),
				func(datum *continuous.Continuous, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units invalid; value missing",
				pointer.FromString("invalid"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; trend missing",
				nil,
				func(datum *continuous.Continuous, units *string) { datum.Trend = nil },
				nil,
			),
			Entry("does not modify the datum; trend rate missing",
				nil,
				func(datum *continuous.Continuous, units *string) { datum.TrendRate = nil },
				nil,
			),
		)

		DescribeTable("normalizes the datum with origin external",
			func(units *string, mutator func(datum *continuous.Continuous, units *string), expectator func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string, value *float64)) {
				datum := NewContinuous(units)
				mutator(datum, units)
				originalValue := pointer.CloneFloat64(datum.Value)
				expectedDatum := CloneContinuous(datum)
				normalizer := dataNormalizer.New(logTest.NewLogger())
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
			Entry("does not modify the datum; units invalid; value missing",
				pointer.FromString("invalid"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units mmol/L",
				pointer.FromString("mmol/L"),
				func(datum *continuous.Continuous, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units mmol/L; value missing",
				pointer.FromString("mmol/L"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				nil,
			),
			Entry("modifies the datum; units mmol/l",
				pointer.FromString("mmol/l"),
				func(datum *continuous.Continuous, units *string) {},
				func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string, value *float64) {
					dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					dataBloodGlucoseTest.ExpectRaw(datum.Raw, &metadata.Metadata{"units": *units, "value": *value})
				},
			),
			Entry("modifies the datum; units mmol/l; value missing",
				pointer.FromString("mmol/l"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string, value *float64) {
					dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					dataBloodGlucoseTest.ExpectRaw(datum.Raw, &metadata.Metadata{"units": *units, "value": nil})
				},
			),
			Entry("modifies the datum; units mg/dL",
				pointer.FromString("mg/dL"),
				func(datum *continuous.Continuous, units *string) {},
				func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string, value *float64) {
					dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					dataBloodGlucoseTest.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
					dataBloodGlucoseTest.ExpectRaw(datum.Raw, &metadata.Metadata{"units": *units, "value": *value})
				},
			),
			Entry("modifies the datum; units mg/dL; value missing",
				pointer.FromString("mg/dL"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string, value *float64) {
					dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					dataBloodGlucoseTest.ExpectRaw(datum.Raw, &metadata.Metadata{"units": *units, "value": nil})
				},
			),
			Entry("modifies the datum; units mg/dl",
				pointer.FromString("mg/dl"),
				func(datum *continuous.Continuous, units *string) {},
				func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string, value *float64) {
					dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					dataBloodGlucoseTest.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
					dataBloodGlucoseTest.ExpectRaw(datum.Raw, &metadata.Metadata{"units": *units, "value": *value})
				},
			),
			Entry("modifies the datum; units mg/dl; value missing",
				pointer.FromString("mg/dl"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string, value *float64) {
					dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					dataBloodGlucoseTest.ExpectRaw(datum.Raw, &metadata.Metadata{"units": *units, "value": nil})
				},
			),
		)

		DescribeTable("normalizes the datum with origin internal/store",
			func(units *string, mutator func(datum *continuous.Continuous, units *string), expectator func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string)) {
				for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
					datum := NewContinuous(units)
					mutator(datum, units)
					expectedDatum := CloneContinuous(datum)
					normalizer := dataNormalizer.New(logTest.NewLogger())
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
				func(datum *continuous.Continuous, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units missing; value missing",
				nil,
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units invalid",
				pointer.FromString("invalid"),
				func(datum *continuous.Continuous, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units invalid; value missing",
				pointer.FromString("invalid"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units mmol/L",
				pointer.FromString("mmol/L"),
				func(datum *continuous.Continuous, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units mmol/L; value missing",
				pointer.FromString("mmol/L"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units mmol/l",
				pointer.FromString("mmol/l"),
				func(datum *continuous.Continuous, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units mmol/l; value missing",
				pointer.FromString("mmol/l"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units mg/dL",
				pointer.FromString("mg/dL"),
				func(datum *continuous.Continuous, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units mg/dL; value missing",
				pointer.FromString("mg/dL"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units mg/dl",
				pointer.FromString("mg/dl"),
				func(datum *continuous.Continuous, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units mg/dl; value missing",
				pointer.FromString("mg/dl"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				nil,
			),
		)
	})
})
