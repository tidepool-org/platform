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

func NewBolusCalculator() *pump.BolusCalculator {
	datum := pump.NewBolusCalculator()
	datum.Enabled = pointer.FromBool(test.RandomBool())
	datum.Insulin = NewBolusCalculatorInsulin()
	return datum
}

func CloneBolusCalculator(datum *pump.BolusCalculator) *pump.BolusCalculator {
	if datum == nil {
		return nil
	}
	clone := pump.NewBolusCalculator()
	clone.Enabled = test.CloneBool(datum.Enabled)
	clone.Insulin = CloneBolusCalculatorInsulin(datum.Insulin)
	return clone
}

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
					datum := NewBolusCalculator()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.BolusCalculator) {},
				),
				Entry("enabled missing",
					func(datum *pump.BolusCalculator) { datum.Enabled = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
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
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/insulin/units"),
				),
				Entry("insulin valid",
					func(datum *pump.BolusCalculator) { datum.Insulin = NewBolusCalculatorInsulin() },
				),
				Entry("multiple errors",
					func(datum *pump.BolusCalculator) {
						datum.Enabled = nil
						datum.Insulin.Units = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/insulin/units"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.BolusCalculator)) {
					for _, origin := range structure.Origins() {
						datum := NewBolusCalculator()
						mutator(datum)
						expectedDatum := CloneBolusCalculator(datum)
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
