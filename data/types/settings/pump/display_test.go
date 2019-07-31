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

func NewDisplay() *pump.Display {
	datum := pump.NewDisplay()
	datum.BloodGlucose = NewDisplayBloodGlucose()
	return datum
}

func CloneDisplay(datum *pump.Display) *pump.Display {
	if datum == nil {
		return nil
	}
	clone := pump.NewDisplay()
	clone.BloodGlucose = CloneDisplayBloodGlucose(datum.BloodGlucose)
	return clone
}

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
					datum := NewDisplay()
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
					func(datum *pump.Display) { datum.BloodGlucose = NewDisplayBloodGlucose() },
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
						datum := NewDisplay()
						mutator(datum)
						expectedDatum := CloneDisplay(datum)
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
