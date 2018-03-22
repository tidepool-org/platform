package pump_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewBolus() *pump.Bolus {
	datum := pump.NewBolus()
	datum.AmountMaximum = NewBolusAmountMaximum()
	datum.Combination = NewBolusCombination()
	return datum
}

func CloneBolus(datum *pump.Bolus) *pump.Bolus {
	if datum == nil {
		return nil
	}
	clone := pump.NewBolus()
	clone.AmountMaximum = CloneBolusAmountMaximum(datum.AmountMaximum)
	clone.Combination = CloneBolusCombination(datum.Combination)
	return clone
}

var _ = Describe("Bolus", func() {
	Context("ParseBolus", func() {
		// TODO
	})

	Context("NewBolus", func() {
		It("is successful", func() {
			Expect(pump.NewBolus()).To(Equal(&pump.Bolus{}))
		})
	})

	Context("Bolus", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pump.Bolus), expectedErrors ...error) {
					datum := NewBolus()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.Bolus) {},
				),
				Entry("amount maximum missing",
					func(datum *pump.Bolus) { datum.AmountMaximum = nil },
				),
				Entry("amount maximum invalid",
					func(datum *pump.Bolus) { datum.AmountMaximum.Units = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amountMaximum/units"),
				),
				Entry("amount maximum valid",
					func(datum *pump.Bolus) { datum.AmountMaximum = NewBolusAmountMaximum() },
				),
				Entry("combination missing",
					func(datum *pump.Bolus) { datum.Combination = nil },
				),
				Entry("combination invalid",
					func(datum *pump.Bolus) { datum.Combination.Enabled = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/combination/enabled"),
				),
				Entry("combination valid",
					func(datum *pump.Bolus) { datum.Combination = NewBolusCombination() },
				),
				Entry("multiple errors",
					func(datum *pump.Bolus) {
						datum.AmountMaximum.Units = nil
						datum.Combination.Enabled = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amountMaximum/units"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/combination/enabled"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.Bolus)) {
					for _, origin := range structure.Origins() {
						datum := NewBolus()
						mutator(datum)
						expectedDatum := CloneBolus(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *pump.Bolus) {},
				),
				Entry("does not modify the datum; amount maximum missing",
					func(datum *pump.Bolus) { datum.AmountMaximum = nil },
				),
				Entry("does not modify the datum; combination missing",
					func(datum *pump.Bolus) { datum.Combination = nil },
				),
			)
		})
	})
})
