package physical_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/activity/physical"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewFlight() *physical.Flight {
	datum := physical.NewFlight()
	datum.Count = pointer.Int(test.RandomIntFromRange(0, 10000))
	return datum
}

func CloneFlight(datum *physical.Flight) *physical.Flight {
	if datum == nil {
		return nil
	}
	clone := physical.NewFlight()
	clone.Count = test.CloneInt(datum.Count)
	return clone
}

var _ = Describe("Flight", func() {
	It("FlightCountMaximum is expected", func() {
		Expect(physical.FlightCountMaximum).To(Equal(10000))
	})

	It("FlightCountMinimum is expected", func() {
		Expect(physical.FlightCountMinimum).To(Equal(0))
	})

	Context("ParseFlight", func() {
		// TODO
	})

	Context("NewFlight", func() {
		It("returns the expected datum", func() {
			Expect(physical.NewFlight()).To(Equal(&physical.Flight{}))
		})
	})

	Context("Flight", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *physical.Flight), expectedErrors ...error) {
					datum := NewFlight()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *physical.Flight) {},
				),
				Entry("count missing",
					func(datum *physical.Flight) { datum.Count = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/count"),
				),
				Entry("count out of range (lower)",
					func(datum *physical.Flight) { datum.Count = pointer.Int(-1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 10000), "/count"),
				),
				Entry("count in range (lower)",
					func(datum *physical.Flight) { datum.Count = pointer.Int(0) },
				),
				Entry("count in range (upper)",
					func(datum *physical.Flight) { datum.Count = pointer.Int(10000) },
				),
				Entry("count out of range (upper)",
					func(datum *physical.Flight) { datum.Count = pointer.Int(10001) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(10001, 0, 10000), "/count"),
				),
				Entry("multiple errors",
					func(datum *physical.Flight) { datum.Count = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/count"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *physical.Flight)) {
					for _, origin := range structure.Origins() {
						datum := NewFlight()
						mutator(datum)
						expectedDatum := CloneFlight(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *physical.Flight) {},
				),
				Entry("does not modify the datum; count missing",
					func(datum *physical.Flight) { datum.Count = nil },
				),
			)
		})
	})
})
