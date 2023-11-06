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

var _ = Describe("Display", func() {
	Context("ParseDisplay", func() {
		// TODO
	})

	Context("NewDisplay", func() {
		It("is successful", func() {
			Expect(pump.NewDisplay()).To(Equal(&pump.Display{}))
		})
	})

	Context("Display", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pump.Display), expectedErrors ...error) {
					datum := pumpTest.NewDisplay()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.Display) {},
				),
				Entry("blood glucose missing",
					func(datum *pump.Display) { datum.BloodGlucose = nil },
				),
				Entry("blood glucose invalid",
					func(datum *pump.Display) { datum.BloodGlucose.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/bloodGlucose/units"),
				),
				Entry("blood glucose valid",
					func(datum *pump.Display) { datum.BloodGlucose = pumpTest.NewDisplayBloodGlucose() },
				),
				Entry("multiple errors",
					func(datum *pump.Display) {
						datum.BloodGlucose.Units = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/bloodGlucose/units"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.Display)) {
					for _, origin := range structure.Origins() {
						datum := pumpTest.NewDisplay()
						mutator(datum)
						expectedDatum := pumpTest.CloneDisplay(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *pump.Display) {},
				),
				Entry("does not modify the datum; blood glucose missing",
					func(datum *pump.Display) { datum.BloodGlucose = nil },
				),
			)
		})
	})
})
