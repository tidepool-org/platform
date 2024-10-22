package pump_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	pumpTest "github.com/tidepool-org/platform/data/types/settings/pump/test"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

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
					datum := pumpTest.NewRandomBolus()
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
					func(datum *pump.Bolus) { datum.AmountMaximum = pumpTest.NewBolusAmountMaximum() },
				),
				Entry("extended missing",
					func(datum *pump.Bolus) { datum.Extended = nil },
				),
				Entry("extended invalid",
					func(datum *pump.Bolus) { datum.Extended.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/extended/enabled"),
				),
				Entry("extended valid",
					func(datum *pump.Bolus) { datum.Extended = pumpTest.NewBolusExtended() },
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
						datum := pumpTest.NewRandomBolus()
						mutator(datum)
						expectedDatum := pumpTest.CloneBolus(datum)
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

	Context("BolusMap", func() {
		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pump.BolusMap), expectedErrors ...error) {
					datum := pump.NewBolusMap()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.BolusMap) {},
				),
				Entry("empty",
					func(datum *pump.BolusMap) {
						*datum = *pump.NewBolusMap()
					},
				),

				Entry("single invalid",
					func(datum *pump.BolusMap) {
						invalid := pumpTest.NewRandomBolus()
						invalid.AmountMaximum.Units = nil
						datum.Set("one", invalid)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/one/amountMaximum/units"),
				),
				Entry("single valid",
					func(datum *pump.BolusMap) {
						datum.Set("one", pumpTest.NewRandomBolus())
					},
				),
				Entry("multiple valid",
					func(datum *pump.BolusMap) {
						datum.Set("one", pumpTest.NewRandomBolus())
						datum.Set("two", pumpTest.NewRandomBolus())
						datum.Set("three", pumpTest.NewRandomBolus())
					},
				),
				Entry("multiple errors",
					func(datum *pump.BolusMap) {
						invalid := pumpTest.NewRandomBolus()
						invalid.AmountMaximum.Units = nil

						invalidThree := pumpTest.NewRandomBolus()
						invalidThree.AmountMaximum.Value = nil

						datum.Set("one", invalid)
						datum.Set("two", pumpTest.NewRandomBolus())
						datum.Set("three", invalidThree)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/one/amountMaximum/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/three/amountMaximum/value"),
				),
			)
		})
		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.BolusMap), expectator func(datum *pump.BolusMap, expectedDatum *pump.BolusMap)) {
					for _, origin := range structure.Origins() {
						datum := pumpTest.NewRandomBolusMap(1, 4)
						mutator(datum)
						expectedDatum := pumpTest.CloneBolusMap(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *pump.BolusMap) {},
					nil,
				),
				Entry("does not modify the datum; amountMaximum missing",
					func(datum *pump.BolusMap) {
						for name := range *datum {
							(*(*datum)[name]).AmountMaximum = nil
						}
					},
					nil,
				),
				Entry("does not modify the datum; calculator missing",
					func(datum *pump.BolusMap) {
						for name := range *datum {
							(*(*datum)[name]).Calculator = nil
						}
					},
					nil,
				),
				Entry("does not modify the datum; extended missing",
					func(datum *pump.BolusMap) {
						for name := range *datum {
							(*(*datum)[name]).Extended = nil
						}
					},
					nil,
				),
			)
		})
	})
})
