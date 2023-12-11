package ketone_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/blood/ketone"
	dataTypesBloodTest "github.com/tidepool-org/platform/data/types/blood/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/test"

	dataBloodKetone "github.com/tidepool-org/platform/data/blood/ketone"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewMeta() interface{} {
	return &types.Meta{
		Type: "bloodKetone",
	}
}

func NewKetone(units *string) *ketone.Ketone {
	datum := ketone.New()
	datum.Blood = *dataTypesBloodTest.NewBlood()
	datum.Type = "bloodKetone"
	datum.Units = units
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dataBloodKetone.ValueRangeForUnits(units)))
	return datum
}

func CloneKetone(datum *ketone.Ketone) *ketone.Ketone {
	if datum == nil {
		return nil
	}
	clone := ketone.New()
	clone.Blood = *dataTypesBloodTest.CloneBlood(&datum.Blood)
	return clone
}

var _ = Describe("Ketone", func() {
	It("Type is expected", func() {
		Expect(ketone.Type).To(Equal("bloodKetone"))
	})

	Context("New", func() {
		It("returns the expected datum", func() {
			datum := ketone.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("bloodKetone"))
			Expect(datum.Units).To(BeNil())
			Expect(datum.Value).To(BeNil())
		})
	})

	Context("Ketone", func() {
		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *ketone.Ketone, units *string), expectedErrors ...error) {
					datum := NewKetone(units)
					mutator(datum, units)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *ketone.Ketone, units *string) {},
				),
				Entry("type missing",
					pointer.FromString("mmol/L"),
					func(datum *ketone.Ketone, units *string) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &types.Meta{}),
				),
				Entry("type invalid",
					pointer.FromString("mmol/L"),
					func(datum *ketone.Ketone, units *string) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "bloodKetone"), "/type", &types.Meta{Type: "invalidType"}),
				),
				Entry("type bloodKetone",
					pointer.FromString("mmol/L"),
					func(datum *ketone.Ketone, units *string) { datum.Type = "bloodKetone" },
				),
				Entry("units missing; value missing",
					nil,
					func(datum *ketone.Ketone, units *string) { datum.Value = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units missing; value out of range (lower)",
					nil,
					func(datum *ketone.Ketone, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; value in range (lower)",
					nil,
					func(datum *ketone.Ketone, units *string) { datum.Value = pointer.FromFloat64(0.) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; value in range (upper)",
					nil,
					func(datum *ketone.Ketone, units *string) { datum.Value = pointer.FromFloat64(10.0) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; value out of range (upper)",
					nil,
					func(datum *ketone.Ketone, units *string) { datum.Value = pointer.FromFloat64(10.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units invalid; value missing",
					pointer.FromString("invalid"),
					func(datum *ketone.Ketone, units *string) { datum.Value = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l"}), "/units", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units invalid; value out of range (lower)",
					pointer.FromString("invalid"),
					func(datum *ketone.Ketone, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l"}), "/units", NewMeta()),
				),
				Entry("units invalid; value in range (lower)",
					pointer.FromString("invalid"),
					func(datum *ketone.Ketone, units *string) { datum.Value = pointer.FromFloat64(0.0) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l"}), "/units", NewMeta()),
				),
				Entry("units invalid; value in range (upper)",
					pointer.FromString("invalid"),
					func(datum *ketone.Ketone, units *string) { datum.Value = pointer.FromFloat64(10.0) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l"}), "/units", NewMeta()),
				),
				Entry("units invalid; value out of range (upper)",
					pointer.FromString("invalid"),
					func(datum *ketone.Ketone, units *string) { datum.Value = pointer.FromFloat64(10.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l"}), "/units", NewMeta()),
				),
				Entry("units mmol/L; value missing",
					pointer.FromString("mmol/L"),
					func(datum *ketone.Ketone, units *string) { datum.Value = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units mmol/L; value out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *ketone.Ketone, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 10.0), "/value", NewMeta()),
				),
				Entry("units mmol/L; value in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *ketone.Ketone, units *string) { datum.Value = pointer.FromFloat64(0.0) },
				),
				Entry("units mmol/L; value in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *ketone.Ketone, units *string) { datum.Value = pointer.FromFloat64(10.0) },
				),
				Entry("units mmol/L; value out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *ketone.Ketone, units *string) { datum.Value = pointer.FromFloat64(10.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(10.1, 0.0, 10.0), "/value", NewMeta()),
				),
				Entry("units mmol/l; value missing",
					pointer.FromString("mmol/l"),
					func(datum *ketone.Ketone, units *string) { datum.Value = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units mmol/l; value out of range (lower)",
					pointer.FromString("mmol/l"),
					func(datum *ketone.Ketone, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 10.0), "/value", NewMeta()),
				),
				Entry("units mmol/l; value in range (lower)",
					pointer.FromString("mmol/l"),
					func(datum *ketone.Ketone, units *string) { datum.Value = pointer.FromFloat64(0.0) },
				),
				Entry("units mmol/l; value in range (upper)",
					pointer.FromString("mmol/l"),
					func(datum *ketone.Ketone, units *string) { datum.Value = pointer.FromFloat64(10.0) },
				),
				Entry("units mmol/l; value out of range (upper)",
					pointer.FromString("mmol/l"),
					func(datum *ketone.Ketone, units *string) { datum.Value = pointer.FromFloat64(10.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(10.1, 0.0, 10.0), "/value", NewMeta()),
				),
				Entry("units mg/dL; value missing",
					pointer.FromString("mg/dL"),
					func(datum *ketone.Ketone, units *string) { datum.Value = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("mg/dL", []string{"mmol/L", "mmol/l"}), "/units", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units mg/dL; value out of range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *ketone.Ketone, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("mg/dL", []string{"mmol/L", "mmol/l"}), "/units", NewMeta()),
				),
				Entry("units mg/dL; value in range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *ketone.Ketone, units *string) { datum.Value = pointer.FromFloat64(0.0) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("mg/dL", []string{"mmol/L", "mmol/l"}), "/units", NewMeta()),
				),
				Entry("units mg/dL; value in range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *ketone.Ketone, units *string) { datum.Value = pointer.FromFloat64(10.0) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("mg/dL", []string{"mmol/L", "mmol/l"}), "/units", NewMeta()),
				),
				Entry("units mg/dL; value out of range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *ketone.Ketone, units *string) { datum.Value = pointer.FromFloat64(10.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("mg/dL", []string{"mmol/L", "mmol/l"}), "/units", NewMeta()),
				),
				Entry("units mg/dl; value missing",
					pointer.FromString("mg/dl"),
					func(datum *ketone.Ketone, units *string) { datum.Value = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("mg/dl", []string{"mmol/L", "mmol/l"}), "/units", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units mg/dl; value out of range (lower)",
					pointer.FromString("mg/dl"),
					func(datum *ketone.Ketone, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("mg/dl", []string{"mmol/L", "mmol/l"}), "/units", NewMeta()),
				),
				Entry("units mg/dl; value in range (lower)",
					pointer.FromString("mg/dl"),
					func(datum *ketone.Ketone, units *string) { datum.Value = pointer.FromFloat64(0.0) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("mg/dl", []string{"mmol/L", "mmol/l"}), "/units", NewMeta()),
				),
				Entry("units mg/dl; value in range (upper)",
					pointer.FromString("mg/dl"),
					func(datum *ketone.Ketone, units *string) { datum.Value = pointer.FromFloat64(10.0) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("mg/dl", []string{"mmol/L", "mmol/l"}), "/units", NewMeta()),
				),
				Entry("units mg/dl; value out of range (upper)",
					pointer.FromString("mg/dl"),
					func(datum *ketone.Ketone, units *string) { datum.Value = pointer.FromFloat64(10.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("mg/dl", []string{"mmol/L", "mmol/l"}), "/units", NewMeta()),
				),
				Entry("multiple errors",
					nil,
					func(datum *ketone.Ketone, units *string) {
						datum.Type = ""
						datum.Value = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &types.Meta{}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", &types.Meta{}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", &types.Meta{}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(units *string, mutator func(datum *ketone.Ketone, units *string), expectator func(datum *ketone.Ketone, expectedDatum *ketone.Ketone, units *string)) {
					for _, origin := range structure.Origins() {
						datum := NewKetone(units)
						mutator(datum, units)
						expectedDatum := CloneKetone(datum)
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
					func(datum *ketone.Ketone, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing; value missing",
					nil,
					func(datum *ketone.Ketone, units *string) { datum.Value = nil },
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *ketone.Ketone, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid; value missing",
					pointer.FromString("invalid"),
					func(datum *ketone.Ketone, units *string) { datum.Value = nil },
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *ketone.Ketone, units *string), expectator func(datum *ketone.Ketone, expectedDatum *ketone.Ketone, units *string)) {
					datum := NewKetone(units)
					mutator(datum, units)
					expectedDatum := CloneKetone(datum)
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
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *ketone.Ketone, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/L; value missing",
					pointer.FromString("mmol/L"),
					func(datum *ketone.Ketone, units *string) { datum.Value = nil },
					nil,
				),
				Entry("modifies the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *ketone.Ketone, units *string) {},
					func(datum *ketone.Ketone, expectedDatum *ketone.Ketone, units *string) {
						Expect(datum.Units).ToNot(BeNil())
						Expect(*datum.Units).To(Equal("mmol/L"))
						expectedDatum.Units = datum.Units
					},
				),
				Entry("modifies the datum; units mmol/l; value missing",
					pointer.FromString("mmol/l"),
					func(datum *ketone.Ketone, units *string) { datum.Value = nil },
					func(datum *ketone.Ketone, expectedDatum *ketone.Ketone, units *string) {
						Expect(datum.Units).ToNot(BeNil())
						Expect(*datum.Units).To(Equal("mmol/L"))
						expectedDatum.Units = datum.Units
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *ketone.Ketone, units *string), expectator func(datum *ketone.Ketone, expectedDatum *ketone.Ketone, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := NewKetone(units)
						mutator(datum, units)
						expectedDatum := CloneKetone(datum)
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
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *ketone.Ketone, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/L; value missing",
					pointer.FromString("mmol/L"),
					func(datum *ketone.Ketone, units *string) { datum.Value = nil },
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *ketone.Ketone, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l; value missing",
					pointer.FromString("mmol/l"),
					func(datum *ketone.Ketone, units *string) { datum.Value = nil },
					nil,
				),
			)
		})
	})
})
