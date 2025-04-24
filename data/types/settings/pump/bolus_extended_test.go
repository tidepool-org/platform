package pump_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	pumpTest "github.com/tidepool-org/platform/data/types/settings/pump/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("BolusExtended", func() {
	Context("ParseBolusExtended", func() {
		// TODO
	})

	Context("NewBolusExtended", func() {
		It("is successful", func() {
			Expect(pump.NewBolusExtended()).To(Equal(&pump.BolusExtended{}))
		})
	})

	Context("BolusExtended", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pump.BolusExtended), expectedErrors ...error) {
					datum := pumpTest.NewBolusExtended()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.BolusExtended) {},
				),
				Entry("enabled missing",
					func(datum *pump.BolusExtended) { datum.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled false",
					func(datum *pump.BolusExtended) { datum.Enabled = pointer.FromBool(false) },
				),
				Entry("enabled true",
					func(datum *pump.BolusExtended) { datum.Enabled = pointer.FromBool(true) },
				),
				Entry("multiple errors",
					func(datum *pump.BolusExtended) {
						datum.Enabled = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.BolusExtended)) {
					for _, origin := range structure.Origins() {
						datum := pumpTest.NewBolusExtended()
						mutator(datum)
						expectedDatum := pumpTest.CloneBolusExtended(datum)
						normalizer := dataNormalizer.New(logTest.NewLogger())
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *pump.BolusExtended) {},
				),
				Entry("does not modify the datum; enabled missing",
					func(datum *pump.BolusExtended) { datum.Enabled = nil },
				),
			)
		})
	})
})
