package continuous_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	testDataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	testDataTypesBloodGlucose "github.com/tidepool-org/platform/data/types/blood/glucose/test"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
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
	datum.Glucose = *testDataTypesBloodGlucose.NewGlucose(units)
	datum.Type = "cbg"
	return datum
}

func CloneContinuous(datum *continuous.Continuous) *continuous.Continuous {
	if datum == nil {
		return nil
	}
	clone := continuous.New()
	clone.Glucose = *testDataTypesBloodGlucose.CloneGlucose(&datum.Glucose)
	return clone
}

var _ = Describe("Continuous", func() {
	It("Type is expected", func() {
		Expect(continuous.Type).To(Equal("cbg"))
	})

	Context("NewDatum", func() {
		It("returns the expected datum", func() {
			Expect(continuous.NewDatum()).To(Equal(&continuous.Continuous{}))
		})
	})

	Context("New", func() {
		It("returns the expected datum", func() {
			Expect(continuous.New()).To(Equal(&continuous.Continuous{}))
		})
	})

	Context("Init", func() {
		It("returns the expected datum", func() {
			datum := continuous.Init()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("cbg"))
			Expect(datum.Units).To(BeNil())
			Expect(datum.Value).To(BeNil())
		})
	})

	Context("with new datum", func() {
		var datum *continuous.Continuous

		BeforeEach(func() {
			datum = NewContinuous(pointer.String("mmol/L"))
		})

		Context("Init", func() {
			It("initializes the continuous", func() {
				datum.Init()
				Expect(datum.Type).To(Equal("cbg"))
				Expect(datum.Units).To(BeNil())
				Expect(datum.Value).To(BeNil())
			})
		})
	})

	Context("Validate", func() {
		DescribeTable("validates the datum",
			func(units *string, mutator func(datum *continuous.Continuous, units *string), expectedErrors ...error) {
				datum := NewContinuous(units)
				mutator(datum, units)
				testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
			},
			Entry("succeeds",
				pointer.String("mmol/L"),
				func(datum *continuous.Continuous, units *string) {},
			),
			Entry("type missing",
				pointer.String("mmol/L"),
				func(datum *continuous.Continuous, units *string) { datum.Type = "" },
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &types.Meta{}),
			),
			Entry("type invalid",
				pointer.String("mmol/L"),
				func(datum *continuous.Continuous, units *string) { datum.Type = "invalidType" },
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "cbg"), "/type", &types.Meta{Type: "invalidType"}),
			),
			Entry("type cbg",
				pointer.String("mmol/L"),
				func(datum *continuous.Continuous, units *string) { datum.Type = "cbg" },
			),
			Entry("units missing; value missing",
				nil,
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
			),
			Entry("units missing; value out of range (lower)",
				nil,
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.Float64(-0.1) },
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
			),
			Entry("units missing; value in range (lower)",
				nil,
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.Float64(0.0) },
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
			),
			Entry("units missing; value in range (upper)",
				nil,
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.Float64(55.0) },
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
			),
			Entry("units missing; value out of range (upper)",
				nil,
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.Float64(1000.1) },
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
			),
			Entry("units invalid; value missing",
				pointer.String("invalid"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
			),
			Entry("units invalid; value out of range (lower)",
				pointer.String("invalid"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.Float64(-0.1) },
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
			),
			Entry("units invalid; value in range (lower)",
				pointer.String("invalid"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.Float64(0.0) },
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
			),
			Entry("units invalid; value in range (upper)",
				pointer.String("invalid"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.Float64(55.0) },
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
			),
			Entry("units invalid; value out of range (upper)",
				pointer.String("invalid"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.Float64(1000.1) },
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
			),
			Entry("units mmol/L; value missing",
				pointer.String("mmol/L"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
			),
			Entry("units mmol/L; value out of range (lower)",
				pointer.String("mmol/L"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.Float64(-0.1) },
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/value", NewMeta()),
			),
			Entry("units mmol/L; value in range (lower)",
				pointer.String("mmol/L"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.Float64(0.0) },
			),
			Entry("units mmol/L; value in range (upper)",
				pointer.String("mmol/L"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.Float64(55.0) },
			),
			Entry("units mmol/L; value out of range (upper)",
				pointer.String("mmol/L"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.Float64(55.1) },
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/value", NewMeta()),
			),
			Entry("units mmol/l; value missing",
				pointer.String("mmol/l"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
			),
			Entry("units mmol/l; value out of range (lower)",
				pointer.String("mmol/l"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.Float64(-0.1) },
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/value", NewMeta()),
			),
			Entry("units mmol/l; value in range (lower)",
				pointer.String("mmol/l"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.Float64(0.0) },
			),
			Entry("units mmol/l; value in range (upper)",
				pointer.String("mmol/l"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.Float64(55.0) },
			),
			Entry("units mmol/l; value out of range (upper)",
				pointer.String("mmol/l"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.Float64(55.1) },
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/value", NewMeta()),
			),
			Entry("units mg/dL; value missing",
				pointer.String("mg/dL"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
			),
			Entry("units mg/dL; value out of range (lower)",
				pointer.String("mg/dL"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.Float64(-0.1) },
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/value", NewMeta()),
			),
			Entry("units mg/dL; value in range (lower)",
				pointer.String("mg/dL"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.Float64(0.0) },
			),
			Entry("units mg/dL; value in range (upper)",
				pointer.String("mg/dL"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.Float64(1000.0) },
			),
			Entry("units mg/dL; value out of range (upper)",
				pointer.String("mg/dL"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.Float64(1000.1) },
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/value", NewMeta()),
			),
			Entry("units mg/dl; value missing",
				pointer.String("mg/dl"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
			),
			Entry("units mg/dl; value out of range (lower)",
				pointer.String("mg/dl"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.Float64(-0.1) },
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/value", NewMeta()),
			),
			Entry("units mg/dl; value in range (lower)",
				pointer.String("mg/dl"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.Float64(0.0) },
			),
			Entry("units mg/dl; value in range (upper)",
				pointer.String("mg/dl"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.Float64(1000.0) },
			),
			Entry("units mg/dl; value out of range (upper)",
				pointer.String("mg/dl"),
				func(datum *continuous.Continuous, units *string) { datum.Value = pointer.Float64(1000.1) },
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/value", NewMeta()),
			),
			Entry("multiple errors",
				nil,
				func(datum *continuous.Continuous, units *string) {
					datum.Type = ""
					datum.Value = nil
				},
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &types.Meta{}),
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", &types.Meta{}),
				testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", &types.Meta{}),
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
				func(datum *continuous.Continuous, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units missing; value missing",
				nil,
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units invalid",
				pointer.String("invalid"),
				func(datum *continuous.Continuous, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units invalid; value missing",
				pointer.String("invalid"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				nil,
			),
		)

		DescribeTable("normalizes the datum with origin external",
			func(units *string, mutator func(datum *continuous.Continuous, units *string), expectator func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string)) {
				datum := NewContinuous(units)
				mutator(datum, units)
				expectedDatum := CloneContinuous(datum)
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
			Entry("does not modify the datum; units invalid; value missing",
				pointer.String("invalid"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units mmol/L",
				pointer.String("mmol/L"),
				func(datum *continuous.Continuous, units *string) {},
				func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string) {
					testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					testDataBloodGlucose.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
				},
			),
			Entry("does not modify the datum; units mmol/L; value missing",
				pointer.String("mmol/L"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string) {
					testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
				},
			),
			Entry("modifies the datum; units mmol/l",
				pointer.String("mmol/l"),
				func(datum *continuous.Continuous, units *string) {},
				func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string) {
					testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					testDataBloodGlucose.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
				},
			),
			Entry("modifies the datum; units mmol/l; value missing",
				pointer.String("mmol/l"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string) {
					testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
				},
			),
			Entry("modifies the datum; units mg/dL",
				pointer.String("mg/dL"),
				func(datum *continuous.Continuous, units *string) {},
				func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string) {
					testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					testDataBloodGlucose.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
				},
			),
			Entry("modifies the datum; units mg/dL; value missing",
				pointer.String("mg/dL"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string) {
					testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
				},
			),
			Entry("modifies the datum; units mg/dl",
				pointer.String("mg/dl"),
				func(datum *continuous.Continuous, units *string) {},
				func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string) {
					testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					testDataBloodGlucose.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
				},
			),
			Entry("modifies the datum; units mg/dl; value missing",
				pointer.String("mg/dl"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string) {
					testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
				},
			),
		)

		DescribeTable("normalizes the datum with origin internal/store",
			func(units *string, mutator func(datum *continuous.Continuous, units *string), expectator func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string)) {
				for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
					datum := NewContinuous(units)
					mutator(datum, units)
					expectedDatum := CloneContinuous(datum)
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
				func(datum *continuous.Continuous, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units missing; value missing",
				nil,
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units invalid",
				pointer.String("invalid"),
				func(datum *continuous.Continuous, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units invalid; value missing",
				pointer.String("invalid"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units mmol/L",
				pointer.String("mmol/L"),
				func(datum *continuous.Continuous, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units mmol/L; value missing",
				pointer.String("mmol/L"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units mmol/l",
				pointer.String("mmol/l"),
				func(datum *continuous.Continuous, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units mmol/l; value missing",
				pointer.String("mmol/l"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units mg/dL",
				pointer.String("mg/dL"),
				func(datum *continuous.Continuous, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units mg/dL; value missing",
				pointer.String("mg/dL"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units mg/dl",
				pointer.String("mg/dl"),
				func(datum *continuous.Continuous, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units mg/dl; value missing",
				pointer.String("mg/dl"),
				func(datum *continuous.Continuous, units *string) { datum.Value = nil },
				nil,
			),
		)
	})
})
