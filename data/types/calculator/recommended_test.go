package calculator_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/calculator"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewRecommended() *calculator.Recommended {
	datum := calculator.NewRecommended()
	datum.Carbohydrate = pointer.FromFloat64(test.RandomFloat64FromRange(calculator.CarbohydrateMinimum, calculator.CarbohydrateMaximum))
	datum.Correction = pointer.FromFloat64(test.RandomFloat64FromRange(calculator.CorrectionMinimum, calculator.CorrectionMaximum))
	datum.Net = pointer.FromFloat64(test.RandomFloat64FromRange(calculator.NetMinimum, calculator.NetMaximum))
	return datum
}

func CloneRecommended(datum *calculator.Recommended) *calculator.Recommended {
	if datum == nil {
		return nil
	}
	clone := calculator.NewRecommended()
	clone.Carbohydrate = pointer.CloneFloat64(datum.Carbohydrate)
	clone.Correction = pointer.CloneFloat64(datum.Correction)
	clone.Net = pointer.CloneFloat64(datum.Net)
	return clone
}

var _ = Describe("Recommended", func() {
	It("CarbohydrateMaximum is expected", func() {
		Expect(calculator.CarbohydrateMaximum).To(Equal(100.0))
	})

	It("CarbohydrateMinimum is expected", func() {
		Expect(calculator.CarbohydrateMinimum).To(Equal(0.0))
	})

	It("CorrectionMaximum is expected", func() {
		Expect(calculator.CorrectionMaximum).To(Equal(100.0))
	})

	It("CorrectionMinimum is expected", func() {
		Expect(calculator.CorrectionMinimum).To(Equal(-100.0))
	})

	It("NetMaximum is expected", func() {
		Expect(calculator.NetMaximum).To(Equal(100.0))
	})

	It("NetMinimum is expected", func() {
		Expect(calculator.NetMinimum).To(Equal(-100.0))
	})

	Context("ParseRecommended", func() {
		// TODO
	})

	Context("NewRecommended", func() {
		It("is successful", func() {
			Expect(calculator.NewRecommended()).To(Equal(&calculator.Recommended{}))
		})
	})

	Context("Recommended", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *calculator.Recommended), expectedErrors ...error) {
					datum := NewRecommended()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *calculator.Recommended) {},
				),
				Entry("carbohydrate missing",
					func(datum *calculator.Recommended) { datum.Carbohydrate = nil },
				),
				Entry("carbohydrate out of range (lower)",
					func(datum *calculator.Recommended) { datum.Carbohydrate = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 100), "/carb"),
				),
				Entry("carbohydrate in range (lower)",
					func(datum *calculator.Recommended) { datum.Carbohydrate = pointer.FromFloat64(0.0) },
				),
				Entry("carbohydrate in range (upper)",
					func(datum *calculator.Recommended) { datum.Carbohydrate = pointer.FromFloat64(100.0) },
				),
				Entry("carbohydrate out of range (upper)",
					func(datum *calculator.Recommended) { datum.Carbohydrate = pointer.FromFloat64(100.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, 0, 100), "/carb"),
				),
				Entry("correction missing",
					func(datum *calculator.Recommended) { datum.Correction = nil },
				),
				Entry("correction out of range (lower)",
					func(datum *calculator.Recommended) { datum.Correction = pointer.FromFloat64(-100.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-100.1, -100, 100), "/correction"),
				),
				Entry("correction in range (lower)",
					func(datum *calculator.Recommended) { datum.Correction = pointer.FromFloat64(-100.0) },
				),
				Entry("correction in range (upper)",
					func(datum *calculator.Recommended) { datum.Correction = pointer.FromFloat64(100.0) },
				),
				Entry("correction out of range (upper)",
					func(datum *calculator.Recommended) { datum.Correction = pointer.FromFloat64(100.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, -100, 100), "/correction"),
				),
				Entry("net missing",
					func(datum *calculator.Recommended) { datum.Net = nil },
				),
				Entry("net out of range (lower)",
					func(datum *calculator.Recommended) { datum.Net = pointer.FromFloat64(-100.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-100.1, -100, 100), "/net"),
				),
				Entry("net in range (lower)",
					func(datum *calculator.Recommended) { datum.Net = pointer.FromFloat64(-100.0) },
				),
				Entry("net in range (upper)",
					func(datum *calculator.Recommended) { datum.Net = pointer.FromFloat64(100.0) },
				),
				Entry("net out of range (upper)",
					func(datum *calculator.Recommended) { datum.Net = pointer.FromFloat64(100.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, -100, 100), "/net"),
				),
				Entry("multiple errors",
					func(datum *calculator.Recommended) {
						datum.Carbohydrate = pointer.FromFloat64(-0.1)
						datum.Correction = pointer.FromFloat64(-100.1)
						datum.Net = pointer.FromFloat64(-100.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 100), "/carb"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-100.1, -100, 100), "/correction"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-100.1, -100, 100), "/net"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *calculator.Recommended)) {
					for _, origin := range structure.Origins() {
						datum := NewRecommended()
						mutator(datum)
						expectedDatum := CloneRecommended(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *calculator.Recommended) {},
				),
				Entry("does not modify the datum; carbohydrate missing",
					func(datum *calculator.Recommended) { datum.Carbohydrate = nil },
				),
				Entry("does not modify the datum; correction missing",
					func(datum *calculator.Recommended) { datum.Correction = nil },
				),
				Entry("does not modify the datum; net missing",
					func(datum *calculator.Recommended) { datum.Net = nil },
				),
			)
		})
	})
})
