package pump_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewBolus() *pump.Bolus {
	datum := pump.NewBolus()
	datum.AmountMaximum = NewBolusAmountMaximum()
	datum.Extended = NewBolusExtended()
	return datum
}

func CloneBolus(datum *pump.Bolus) *pump.Bolus {
	if datum == nil {
		return nil
	}
	clone := pump.NewBolus()
	clone.AmountMaximum = CloneBolusAmountMaximum(datum.AmountMaximum)
	clone.Extended = CloneBolusExtended(datum.Extended)
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
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.Bolus) {},
				),
				Entry("amount maximum missing",
					func(datum *pump.Bolus) { datum.AmountMaximum = nil },
				),
				Entry("amount maximum invalid",
					func(datum *pump.Bolus) { datum.AmountMaximum.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amountMaximum/units"),
				),
				Entry("amount maximum valid",
					func(datum *pump.Bolus) { datum.AmountMaximum = NewBolusAmountMaximum() },
				),
				Entry("extended missing",
					func(datum *pump.Bolus) { datum.Extended = nil },
				),
				Entry("extended invalid",
					func(datum *pump.Bolus) { datum.Extended.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/extended/enabled"),
				),
				Entry("extended valid",
					func(datum *pump.Bolus) { datum.Extended = NewBolusExtended() },
				),
				Entry("multiple errors",
					func(datum *pump.Bolus) {
						datum.AmountMaximum.Units = nil
						datum.Extended.Enabled = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amountMaximum/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/extended/enabled"),
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
				Entry("does not modify the datum; extended missing",
					func(datum *pump.Bolus) { datum.Extended = nil },
				),
			)
		})
	})
})
