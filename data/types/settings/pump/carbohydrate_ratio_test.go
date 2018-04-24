package pump_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewCarbohydrateRatio() *pump.CarbohydrateRatio {
	datum := pump.NewCarbohydrateRatio()
	datum.Amount = pointer.Float64(test.RandomFloat64FromRange(pump.CarbohydrateRatioAmountMinimum, pump.CarbohydrateRatioAmountMaximum))
	datum.Start = pointer.Int(test.RandomIntFromRange(pump.CarbohydrateRatioStartMinimum, pump.CarbohydrateRatioStartMaximum))
	return datum
}

func CloneCarbohydrateRatio(datum *pump.CarbohydrateRatio) *pump.CarbohydrateRatio {
	if datum == nil {
		return nil
	}
	clone := pump.NewCarbohydrateRatio()
	clone.Amount = test.CloneFloat64(datum.Amount)
	clone.Start = test.CloneInt(datum.Start)
	return clone
}

func NewCarbohydrateRatioArray() *pump.CarbohydrateRatioArray {
	datum := pump.NewCarbohydrateRatioArray()
	*datum = append(*datum, NewCarbohydrateRatio())
	return datum
}

func CloneCarbohydrateRatioArray(datumArray *pump.CarbohydrateRatioArray) *pump.CarbohydrateRatioArray {
	if datumArray == nil {
		return nil
	}
	clone := pump.NewCarbohydrateRatioArray()
	for _, datum := range *datumArray {
		*clone = append(*clone, CloneCarbohydrateRatio(datum))
	}
	return clone
}

var _ = Describe("CarbohydrateRatio", func() {
	It("CarbohydrateRatioAmountMaximum is expected", func() {
		Expect(pump.CarbohydrateRatioAmountMaximum).To(Equal(250.0))
	})

	It("CarbohydrateRatioAmountMinimum is expected", func() {
		Expect(pump.CarbohydrateRatioAmountMinimum).To(Equal(0.0))
	})

	It("CarbohydrateRatioStartMaximum is expected", func() {
		Expect(pump.CarbohydrateRatioStartMaximum).To(Equal(86400000))
	})

	It("CarbohydrateRatioStartMinimum is expected", func() {
		Expect(pump.CarbohydrateRatioStartMinimum).To(Equal(0))
	})

	Context("ParseCarbohydrateRatio", func() {
		// TODO
	})

	Context("NewCarbohydrateRatio", func() {
		It("is successful", func() {
			Expect(pump.NewCarbohydrateRatio()).To(Equal(&pump.CarbohydrateRatio{}))
		})
	})

	Context("CarbohydrateRatio", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pump.CarbohydrateRatio), expectedErrors ...error) {
					datum := NewCarbohydrateRatio()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.CarbohydrateRatio) {},
				),
				Entry("amount missing",
					func(datum *pump.CarbohydrateRatio) { datum.Amount = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
				Entry("amount out of range (lower)",
					func(datum *pump.CarbohydrateRatio) { datum.Amount = pointer.Float64(-0.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 250), "/amount"),
				),
				Entry("amount in range (lower)",
					func(datum *pump.CarbohydrateRatio) { datum.Amount = pointer.Float64(0.0) },
				),
				Entry("amount in range (upper)",
					func(datum *pump.CarbohydrateRatio) { datum.Amount = pointer.Float64(250.0) },
				),
				Entry("amount out of range (upper)",
					func(datum *pump.CarbohydrateRatio) { datum.Amount = pointer.Float64(250.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(250.1, 0, 250), "/amount"),
				),
				Entry("start missing",
					func(datum *pump.CarbohydrateRatio) { datum.Start = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/start"),
				),
				Entry("start out of range (lower)",
					func(datum *pump.CarbohydrateRatio) { datum.Start = pointer.Int(-1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/start"),
				),
				Entry("start in range (lower)",
					func(datum *pump.CarbohydrateRatio) { datum.Start = pointer.Int(0) },
				),
				Entry("start in range (upper)",
					func(datum *pump.CarbohydrateRatio) { datum.Start = pointer.Int(86400000) },
				),
				Entry("start out of range (upper)",
					func(datum *pump.CarbohydrateRatio) { datum.Start = pointer.Int(86400001) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/start"),
				),
				Entry("multiple errors",
					func(datum *pump.CarbohydrateRatio) {
						datum.Amount = nil
						datum.Start = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/start"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.CarbohydrateRatio)) {
					for _, origin := range structure.Origins() {
						datum := NewCarbohydrateRatio()
						mutator(datum)
						expectedDatum := CloneCarbohydrateRatio(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *pump.CarbohydrateRatio) {},
				),
				Entry("does not modify the datum; amount missing",
					func(datum *pump.CarbohydrateRatio) { datum.Amount = nil },
				),
				Entry("does not modify the datum; start missing",
					func(datum *pump.CarbohydrateRatio) { datum.Start = nil },
				),
			)
		})
	})

	Context("ParseCarbohydrateRatioArray", func() {
		// TODO
	})

	Context("NewCarbohydrateRatioArray", func() {
		It("is successful", func() {
			Expect(pump.NewCarbohydrateRatioArray()).To(Equal(&pump.CarbohydrateRatioArray{}))
		})
	})

	Context("CarbohydrateRatioArray", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pump.CarbohydrateRatioArray), expectedErrors ...error) {
					datum := pump.NewCarbohydrateRatioArray()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.CarbohydrateRatioArray) {},
				),
				Entry("empty",
					func(datum *pump.CarbohydrateRatioArray) { *datum = *pump.NewCarbohydrateRatioArray() },
				),
				Entry("nil",
					func(datum *pump.CarbohydrateRatioArray) { *datum = append(*datum, nil) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
				),
				Entry("single invalid",
					func(datum *pump.CarbohydrateRatioArray) {
						invalid := NewCarbohydrateRatio()
						invalid.Amount = nil
						*datum = append(*datum, invalid)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0/amount"),
				),
				Entry("single valid",
					func(datum *pump.CarbohydrateRatioArray) {
						*datum = append(*datum, NewCarbohydrateRatio())
					},
				),
				Entry("multiple invalid",
					func(datum *pump.CarbohydrateRatioArray) {
						invalid := NewCarbohydrateRatio()
						invalid.Amount = nil
						*datum = append(*datum, NewCarbohydrateRatio(), invalid, NewCarbohydrateRatio())
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/amount"),
				),
				Entry("multiple valid",
					func(datum *pump.CarbohydrateRatioArray) {
						*datum = append(*datum, NewCarbohydrateRatio(), NewCarbohydrateRatio(), NewCarbohydrateRatio())
					},
				),
				Entry("multiple errors",
					func(datum *pump.CarbohydrateRatioArray) {
						invalid := NewCarbohydrateRatio()
						invalid.Amount = nil
						*datum = append(*datum, nil, invalid, NewCarbohydrateRatio())
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/amount"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.CarbohydrateRatioArray)) {
					for _, origin := range structure.Origins() {
						datum := NewCarbohydrateRatioArray()
						mutator(datum)
						expectedDatum := CloneCarbohydrateRatioArray(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *pump.CarbohydrateRatioArray) {},
				),
				Entry("does not modify the datum; amount missing",
					func(datum *pump.CarbohydrateRatioArray) { (*datum)[0].Amount = nil },
				),
				Entry("does not modify the datum; start missing",
					func(datum *pump.CarbohydrateRatioArray) { (*datum)[0].Start = nil },
				),
			)
		})
	})
})
