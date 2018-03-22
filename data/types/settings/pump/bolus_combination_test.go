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

func NewBolusCombination() *pump.BolusCombination {
	datum := pump.NewBolusCombination()
	datum.Enabled = pointer.Bool(test.RandomBool())
	return datum
}

func CloneBolusCombination(datum *pump.BolusCombination) *pump.BolusCombination {
	if datum == nil {
		return nil
	}
	clone := pump.NewBolusCombination()
	clone.Enabled = test.CloneBool(datum.Enabled)
	return clone
}

var _ = Describe("BolusCombination", func() {
	Context("ParseBolusCombination", func() {
		// TODO
	})

	Context("NewBolusCombination", func() {
		It("is successful", func() {
			Expect(pump.NewBolusCombination()).To(Equal(&pump.BolusCombination{}))
		})
	})

	Context("BolusCombination", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pump.BolusCombination), expectedErrors ...error) {
					datum := NewBolusCombination()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.BolusCombination) {},
				),
				Entry("enabled missing",
					func(datum *pump.BolusCombination) { datum.Enabled = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled false",
					func(datum *pump.BolusCombination) { datum.Enabled = pointer.Bool(false) },
				),
				Entry("enabled true",
					func(datum *pump.BolusCombination) { datum.Enabled = pointer.Bool(true) },
				),
				Entry("multiple errors",
					func(datum *pump.BolusCombination) {
						datum.Enabled = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.BolusCombination)) {
					for _, origin := range structure.Origins() {
						datum := NewBolusCombination()
						mutator(datum)
						expectedDatum := CloneBolusCombination(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *pump.BolusCombination) {},
				),
				Entry("does not modify the datum; enabled missing",
					func(datum *pump.BolusCombination) { datum.Enabled = nil },
				),
			)
		})
	})
})
