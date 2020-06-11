package pump_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	pumpTest "github.com/tidepool-org/platform/data/types/settings/pump/test"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("BolusCalculator", func() {
	Context("ParseBolusCalculator", func() {
		// TODO
	})

	Context("NewBolusCalculator", func() {
		It("is successful", func() {
			Expect(pump.NewBolusCalculator()).To(Equal(&pump.BolusCalculator{}))
		})
	})

	Context("BolusCalculator", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pump.BolusCalculator), expectedErrors ...error) {
					datum := pumpTest.NewBolusCalculator()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.BolusCalculator) {},
				),
				Entry("enabled missing",
					func(datum *pump.BolusCalculator) { datum.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled false",
					func(datum *pump.BolusCalculator) { datum.Enabled = pointer.FromBool(false) },
				),
				Entry("enabled true",
					func(datum *pump.BolusCalculator) { datum.Enabled = pointer.FromBool(true) },
				),
				Entry("insulin missing",
					func(datum *pump.BolusCalculator) { datum.Insulin = nil },
				),
				Entry("insulin invalid",
					func(datum *pump.BolusCalculator) { datum.Insulin.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/insulin/units"),
				),
				Entry("insulin valid",
					func(datum *pump.BolusCalculator) { datum.Insulin = pumpTest.NewBolusCalculatorInsulin() },
				),
				Entry("multiple errors",
					func(datum *pump.BolusCalculator) {
						datum.Enabled = nil
						datum.Insulin.Units = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/insulin/units"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.BolusCalculator)) {
					for _, origin := range structure.Origins() {
						datum := pumpTest.NewBolusCalculator()
						mutator(datum)
						expectedDatum := pumpTest.CloneBolusCalculator(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *pump.BolusCalculator) {},
				),
				Entry("does not modify the datum; enabled missing",
					func(datum *pump.BolusCalculator) { datum.Enabled = nil },
				),
				Entry("does not modify the datum; insulin missing",
					func(datum *pump.BolusCalculator) { datum.Insulin = nil },
				),
			)
		})
	})
})
