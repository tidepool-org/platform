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

func NewDisplay() *pump.Display {
	datum := pump.NewDisplay()
	datum.Units = pointer.String(test.RandomStringFromArray(pump.DisplayUnits()))
	return datum
}

func CloneDisplay(datum *pump.Display) *pump.Display {
	if datum == nil {
		return nil
	}
	clone := pump.NewDisplay()
	clone.Units = test.CloneString(datum.Units)
	return clone
}

var _ = Describe("Display", func() {
	It("DisplayUnitsMgPerDL is expected", func() {
		Expect(pump.DisplayUnitsMgPerDL).To(Equal("mg/dL"))
	})

	It("DisplayUnitsMmolPerL is expected", func() {
		Expect(pump.DisplayUnitsMmolPerL).To(Equal("mmol/L"))
	})

	It("DisplayUnits returns expected", func() {
		Expect(pump.DisplayUnits()).To(Equal([]string{"mg/dL", "mmol/L"}))
	})

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
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.Display) {},
				),
				Entry("units missing",
					func(datum *pump.Display) { datum.Units = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid",
					func(datum *pump.Display) { datum.Units = pointer.String("invalid") },
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mg/dL", "mmol/L"}), "/units"),
				),
				Entry("units mg/dL",
					func(datum *pump.Display) { datum.Units = pointer.String("mg/dL") },
				),
				Entry("units mmol/L",
					func(datum *pump.Display) { datum.Units = pointer.String("mmol/L") },
				),
				Entry("multiple errors",
					func(datum *pump.Display) {
						datum.Units = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
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
				Entry("does not modify the datum; units missing",
					func(datum *pump.Display) { datum.Units = nil },
				),
			)
		})
	})
})
