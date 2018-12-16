package pump_test

import (
	"math"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewBolusAmountMaximum() *pump.BolusAmountMaximum {
	datum := pump.NewBolusAmountMaximum()
	datum.Units = pointer.FromString(test.RandomStringFromArray(pump.BolusAmountMaximumUnits()))
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(pump.BolusAmountMaximumValueRangeForUnits(datum.Units)))
	return datum
}

func CloneBolusAmountMaximum(datum *pump.BolusAmountMaximum) *pump.BolusAmountMaximum {
	if datum == nil {
		return nil
	}
	clone := pump.NewBolusAmountMaximum()
	clone.Units = test.CloneString(datum.Units)
	clone.Value = test.CloneFloat64(datum.Value)
	return clone
}

var _ = Describe("BolusAmountMaximum", func() {
	It("BolusAmountMaximumUnitsUnits is expected", func() {
		Expect(pump.BolusAmountMaximumUnitsUnits).To(Equal("Units"))
	})

	It("BolusAmountMaximumValueUnitsMaximum is expected", func() {
		Expect(pump.BolusAmountMaximumValueUnitsMaximum).To(Equal(100.0))
	})

	It("BolusAmountMaximumValueUnitsMinimum is expected", func() {
		Expect(pump.BolusAmountMaximumValueUnitsMinimum).To(Equal(0.0))
	})

	It("BolusAmountMaximumUnits returns expected", func() {
		Expect(pump.BolusAmountMaximumUnits()).To(Equal([]string{"Units"}))
	})

	Context("ParseBolusAmountMaximum", func() {
		// TODO
	})

	Context("NewBolusAmountMaximum", func() {
		It("is successful", func() {
			Expect(pump.NewBolusAmountMaximum()).To(Equal(&pump.BolusAmountMaximum{}))
		})
	})

	Context("BolusAmountMaximum", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pump.BolusAmountMaximum), expectedErrors ...error) {
					datum := NewBolusAmountMaximum()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.BolusAmountMaximum) {},
				),
				Entry("units missing",
					func(datum *pump.BolusAmountMaximum) { datum.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid",
					func(datum *pump.BolusAmountMaximum) { datum.Units = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"Units"}), "/units"),
				),
				Entry("units Units",
					func(datum *pump.BolusAmountMaximum) {
						datum.Units = pointer.FromString("Units")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units missing; value missing",
					func(datum *pump.BolusAmountMaximum) {
						datum.Units = nil
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units missing; value out of range (lower)",
					func(datum *pump.BolusAmountMaximum) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (lower)",
					func(datum *pump.BolusAmountMaximum) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (upper)",
					func(datum *pump.BolusAmountMaximum) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(100.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value out of range (upper)",
					func(datum *pump.BolusAmountMaximum) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; value missing",
					func(datum *pump.BolusAmountMaximum) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"Units"}), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units invalid; value out of range (lower)",
					func(datum *pump.BolusAmountMaximum) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"Units"}), "/units"),
				),
				Entry("units invalid; value in range (lower)",
					func(datum *pump.BolusAmountMaximum) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"Units"}), "/units"),
				),
				Entry("units invalid; value in range (upper)",
					func(datum *pump.BolusAmountMaximum) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(100.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"Units"}), "/units"),
				),
				Entry("units invalid; value out of range (upper)",
					func(datum *pump.BolusAmountMaximum) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"Units"}), "/units"),
				),
				Entry("units Units; value missing",
					func(datum *pump.BolusAmountMaximum) {
						datum.Units = pointer.FromString("Units")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units Units; value out of range (lower)",
					func(datum *pump.BolusAmountMaximum) {
						datum.Units = pointer.FromString("Units")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/value"),
				),
				Entry("units Units; value in range (lower)",
					func(datum *pump.BolusAmountMaximum) {
						datum.Units = pointer.FromString("Units")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units Units; value in range (upper)",
					func(datum *pump.BolusAmountMaximum) {
						datum.Units = pointer.FromString("Units")
						datum.Value = pointer.FromFloat64(100.0)
					},
				),
				Entry("units Units; value out of range (upper)",
					func(datum *pump.BolusAmountMaximum) {
						datum.Units = pointer.FromString("Units")
						datum.Value = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/value"),
				),
				Entry("multiple errors",
					func(datum *pump.BolusAmountMaximum) {
						datum.Units = nil
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.BolusAmountMaximum)) {
					for _, origin := range structure.Origins() {
						datum := NewBolusAmountMaximum()
						mutator(datum)
						expectedDatum := CloneBolusAmountMaximum(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *pump.BolusAmountMaximum) {},
				),
				Entry("does not modify the datum; units missing",
					func(datum *pump.BolusAmountMaximum) { datum.Units = nil },
				),
				Entry("does not modify the datum; value missing",
					func(datum *pump.BolusAmountMaximum) { datum.Value = nil },
				),
			)
		})
	})

	Context("BolusAmountMaximumValueRangeForUnits", func() {
		It("returns expected range for units missing", func() {
			minimum, maximum := pump.BolusAmountMaximumValueRangeForUnits(nil)
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units invalid", func() {
			minimum, maximum := pump.BolusAmountMaximumValueRangeForUnits(pointer.FromString("invalid"))
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units Units", func() {
			minimum, maximum := pump.BolusAmountMaximumValueRangeForUnits(pointer.FromString("Units"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(100.0))
		})
	})
})
