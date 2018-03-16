package insulin_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/insulin"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewDose() *insulin.Dose {
	datum := insulin.NewDose()
	datum.Correction = pointer.Float64(test.RandomFloat64FromRange(insulin.DoseCorrectionMinimum, insulin.DoseCorrectionMaximum))
	datum.Food = pointer.Float64(test.RandomFloat64FromRange(insulin.DoseFoodMinimum, insulin.DoseFoodMaximum))
	datum.Total = pointer.Float64(test.RandomFloat64FromRange(insulin.DoseTotalMinimum, insulin.DoseTotalMaximum))
	datum.Units = pointer.String(test.RandomStringFromArray(insulin.DoseUnits()))
	return datum
}

func CloneDose(datum *insulin.Dose) *insulin.Dose {
	if datum == nil {
		return nil
	}
	clone := insulin.NewDose()
	clone.Correction = test.CloneFloat64(datum.Correction)
	clone.Food = test.CloneFloat64(datum.Food)
	clone.Total = test.CloneFloat64(datum.Total)
	clone.Units = test.CloneString(datum.Units)
	return clone
}

var _ = Describe("Dose", func() {
	It("DoseCorrectionMaximum is expected", func() {
		Expect(insulin.DoseCorrectionMaximum).To(Equal(250.0))
	})

	It("DoseCorrectionMinimum is expected", func() {
		Expect(insulin.DoseCorrectionMinimum).To(Equal(-250.0))
	})

	It("DoseFoodMaximum is expected", func() {
		Expect(insulin.DoseFoodMaximum).To(Equal(250.0))
	})

	It("DoseFoodMinimum is expected", func() {
		Expect(insulin.DoseFoodMinimum).To(Equal(0.0))
	})

	It("DoseTotalMaximum is expected", func() {
		Expect(insulin.DoseTotalMaximum).To(Equal(250.0))
	})

	It("DoseTotalMinimum is expected", func() {
		Expect(insulin.DoseTotalMinimum).To(Equal(0.0))
	})

	It("DoseUnitsUnits is expected", func() {
		Expect(insulin.DoseUnitsUnits).To(Equal("units"))
	})

	It("DoseUnits returns expected", func() {
		Expect(insulin.DoseUnits()).To(Equal([]string{"units"}))
	})

	Context("ParseDose", func() {
		// TODO
	})

	Context("NewDose", func() {
		It("is successful", func() {
			Expect(insulin.NewDose()).To(Equal(&insulin.Dose{}))
		})
	})

	Context("Dose", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *insulin.Dose), expectedErrors ...error) {
					datum := NewDose()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *insulin.Dose) {},
				),
				Entry("correction missing",
					func(datum *insulin.Dose) { datum.Correction = nil },
				),
				Entry("correction out of range (lower)",
					func(datum *insulin.Dose) { datum.Correction = pointer.Float64(-250.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-250.1, -250, 250), "/correction"),
				),
				Entry("correction in range (lower)",
					func(datum *insulin.Dose) { datum.Correction = pointer.Float64(-250.0) },
				),
				Entry("correction in range (upper)",
					func(datum *insulin.Dose) { datum.Correction = pointer.Float64(250.0) },
				),
				Entry("correction out of range (upper)",
					func(datum *insulin.Dose) { datum.Correction = pointer.Float64(250.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(250.1, -250, 250), "/correction"),
				),
				Entry("food missing",
					func(datum *insulin.Dose) { datum.Food = nil },
				),
				Entry("food out of range (lower)",
					func(datum *insulin.Dose) { datum.Food = pointer.Float64(-0.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 250), "/food"),
				),
				Entry("food in range (lower)",
					func(datum *insulin.Dose) { datum.Food = pointer.Float64(0.0) },
				),
				Entry("food in range (upper)",
					func(datum *insulin.Dose) { datum.Food = pointer.Float64(250.0) },
				),
				Entry("food out of range (upper)",
					func(datum *insulin.Dose) { datum.Food = pointer.Float64(250.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(250.1, 0, 250), "/food"),
				),
				Entry("total missing",
					func(datum *insulin.Dose) { datum.Total = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/total"),
				),
				Entry("total out of range (lower)",
					func(datum *insulin.Dose) { datum.Total = pointer.Float64(-0.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 250), "/total"),
				),
				Entry("total in range (lower)",
					func(datum *insulin.Dose) { datum.Total = pointer.Float64(0.0) },
				),
				Entry("total in range (upper)",
					func(datum *insulin.Dose) { datum.Total = pointer.Float64(250.0) },
				),
				Entry("total out of range (upper)",
					func(datum *insulin.Dose) { datum.Total = pointer.Float64(250.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(250.1, 0, 250), "/total"),
				),
				Entry("units missing",
					func(datum *insulin.Dose) { datum.Units = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid",
					func(datum *insulin.Dose) { datum.Units = pointer.String("invalid") },
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"units"}), "/units"),
				),
				Entry("units units",
					func(datum *insulin.Dose) { datum.Units = pointer.String("units") },
				),
				Entry("multiple errors",
					func(datum *insulin.Dose) {
						datum.Total = nil
						datum.Units = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/total"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *insulin.Dose)) {
					for _, origin := range structure.Origins() {
						datum := NewDose()
						mutator(datum)
						expectedDatum := CloneDose(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *insulin.Dose) {},
				),
				Entry("does not modify the datum; correction nil",
					func(datum *insulin.Dose) { datum.Correction = nil },
				),
				Entry("does not modify the datum; food nil",
					func(datum *insulin.Dose) { datum.Food = nil },
				),
				Entry("does not modify the datum; total nil",
					func(datum *insulin.Dose) { datum.Total = nil },
				),
				Entry("does not modify the datum; units nil",
					func(datum *insulin.Dose) { datum.Units = nil },
				),
			)
		})
	})
})
