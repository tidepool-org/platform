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

func NewBolusExtended() *pump.BolusExtended {
	datum := pump.NewBolusExtended()
	datum.Enabled = pointer.FromBool(test.RandomBool())
	return datum
}

func CloneBolusExtended(datum *pump.BolusExtended) *pump.BolusExtended {
	if datum == nil {
		return nil
	}
	clone := pump.NewBolusExtended()
	clone.Enabled = test.CloneBool(datum.Enabled)
	return clone
}

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
					datum := NewBolusExtended()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.BolusExtended) {},
				),
				Entry("enabled missing",
					func(datum *pump.BolusExtended) { datum.Enabled = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
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
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.BolusExtended)) {
					for _, origin := range structure.Origins() {
						datum := NewBolusExtended()
						mutator(datum)
						expectedDatum := CloneBolusExtended(datum)
						normalizer := dataNormalizer.New()
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
