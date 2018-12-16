package insulin_test

import (
	"math"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/insulin"
	dataTypesInsulinTest "github.com/tidepool-org/platform/data/types/insulin/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Concentration", func() {
	It("ConcentrationUnitsUnitsPerML is expected", func() {
		Expect(insulin.ConcentrationUnitsUnitsPerML).To(Equal("Units/mL"))
	})

	It("ConcentrationValueUnitsPerMLMaximum is expected", func() {
		Expect(insulin.ConcentrationValueUnitsPerMLMaximum).To(Equal(10000.0))
	})

	It("ConcentrationValueUnitsPerMLMinimum is expected", func() {
		Expect(insulin.ConcentrationValueUnitsPerMLMinimum).To(Equal(0.0))
	})

	It("ConcentrationUnits returns expected", func() {
		Expect(insulin.ConcentrationUnits()).To(Equal([]string{"Units/mL"}))
	})

	Context("ParseConcentration", func() {
		// TODO
	})

	Context("NewConcentration", func() {
		It("is successful", func() {
			Expect(insulin.NewConcentration()).To(Equal(&insulin.Concentration{}))
		})
	})

	Context("Concentration", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *insulin.Concentration), expectedErrors ...error) {
					datum := dataTypesInsulinTest.NewConcentration()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *insulin.Concentration) {},
				),
				Entry("units missing",
					func(datum *insulin.Concentration) { datum.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid",
					func(datum *insulin.Concentration) { datum.Units = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"Units/mL"}), "/units"),
				),
				Entry("units Units/mL",
					func(datum *insulin.Concentration) {
						datum.Units = pointer.FromString("Units/mL")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units missing; value missing",
					func(datum *insulin.Concentration) {
						datum.Units = nil
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units missing; value out of range (lower)",
					func(datum *insulin.Concentration) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (lower)",
					func(datum *insulin.Concentration) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (upper)",
					func(datum *insulin.Concentration) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(10000.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value out of range (upper)",
					func(datum *insulin.Concentration) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(10000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; value missing",
					func(datum *insulin.Concentration) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"Units/mL"}), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units invalid; value out of range (lower)",
					func(datum *insulin.Concentration) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"Units/mL"}), "/units"),
				),
				Entry("units invalid; value in range (lower)",
					func(datum *insulin.Concentration) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"Units/mL"}), "/units"),
				),
				Entry("units invalid; value in range (upper)",
					func(datum *insulin.Concentration) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(10000.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"Units/mL"}), "/units"),
				),
				Entry("units invalid; value out of range (upper)",
					func(datum *insulin.Concentration) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(10000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"Units/mL"}), "/units"),
				),
				Entry("units Units/mL; value missing",
					func(datum *insulin.Concentration) {
						datum.Units = pointer.FromString("Units/mL")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units Units/mL; value out of range (lower)",
					func(datum *insulin.Concentration) {
						datum.Units = pointer.FromString("Units/mL")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 10000.0), "/value"),
				),
				Entry("units Units/mL; value in range (lower)",
					func(datum *insulin.Concentration) {
						datum.Units = pointer.FromString("Units/mL")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units Units/mL; value in range (upper)",
					func(datum *insulin.Concentration) {
						datum.Units = pointer.FromString("Units/mL")
						datum.Value = pointer.FromFloat64(10000.0)
					},
				),
				Entry("units Units/mL; value out of range (upper)",
					func(datum *insulin.Concentration) {
						datum.Units = pointer.FromString("Units/mL")
						datum.Value = pointer.FromFloat64(10000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(10000.1, 0.0, 10000.0), "/value"),
				),
				Entry("multiple errors",
					func(datum *insulin.Concentration) {
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
				func(mutator func(datum *insulin.Concentration)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesInsulinTest.NewConcentration()
						mutator(datum)
						expectedDatum := dataTypesInsulinTest.CloneConcentration(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *insulin.Concentration) {},
				),
				Entry("does not modify the datum; units missing",
					func(datum *insulin.Concentration) { datum.Units = nil },
				),
				Entry("does not modify the datum; value missing",
					func(datum *insulin.Concentration) { datum.Value = nil },
				),
			)
		})
	})

	Context("ConcentrationValueRangeForUnits", func() {
		It("returns expected range for units missing", func() {
			minimum, maximum := insulin.ConcentrationValueRangeForUnits(nil)
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units invalid", func() {
			minimum, maximum := insulin.ConcentrationValueRangeForUnits(pointer.FromString("invalid"))
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units Units/mL", func() {
			minimum, maximum := insulin.ConcentrationValueRangeForUnits(pointer.FromString("Units/mL"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(10000.0))
		})
	})
})
