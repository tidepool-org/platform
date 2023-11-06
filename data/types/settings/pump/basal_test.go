package pump_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	pumpTest "github.com/tidepool-org/platform/data/types/settings/pump/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Basal", func() {
	Context("ParseBasal", func() {
		// TODO
	})

	Context("NewBasal", func() {
		It("is successful", func() {
			Expect(pump.NewBasal()).To(Equal(&pump.Basal{}))
		})
	})

	Context("Basal", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pump.Basal), expectedErrors ...error) {
					datum := pumpTest.NewBasal()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.Basal) {},
				),
				Entry("rate maximum missing",
					func(datum *pump.Basal) { datum.RateMaximum = nil },
				),
				Entry("rate maximum invalid",
					func(datum *pump.Basal) { datum.RateMaximum.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rateMaximum/units"),
				),
				Entry("rate maximum valid",
					func(datum *pump.Basal) { datum.RateMaximum = pumpTest.NewBasalRateMaximum() },
				),
				Entry("temporary missing",
					func(datum *pump.Basal) { datum.Temporary = nil },
				),
				Entry("temporary invalid",
					func(datum *pump.Basal) { datum.Temporary.Type = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/temporary/type"),
				),
				Entry("temporary valid",
					func(datum *pump.Basal) { datum.Temporary = pumpTest.NewBasalTemporary() },
				),
				Entry("multiple errors",
					func(datum *pump.Basal) {
						datum.RateMaximum.Units = nil
						datum.Temporary.Type = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rateMaximum/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/temporary/type"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.Basal)) {
					for _, origin := range structure.Origins() {
						datum := pumpTest.NewBasal()
						mutator(datum)
						expectedDatum := pumpTest.CloneBasal(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *pump.Basal) {},
				),
				Entry("does not modify the datum; rate maximum missing",
					func(datum *pump.Basal) { datum.RateMaximum = nil },
				),
				Entry("does not modify the datum; temporary missing",
					func(datum *pump.Basal) { datum.Temporary = nil },
				),
			)
		})
	})
})
