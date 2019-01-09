package physical_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/activity/physical"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewLap() *physical.Lap {
	datum := physical.NewLap()
	datum.Count = pointer.FromInt(test.RandomIntFromRange(0, 10000))
	datum.Distance = NewDistance()
	return datum
}

func CloneLap(datum *physical.Lap) *physical.Lap {
	if datum == nil {
		return nil
	}
	clone := physical.NewLap()
	clone.Count = pointer.CloneInt(datum.Count)
	clone.Distance = CloneDistance(datum.Distance)
	return clone
}

var _ = Describe("Lap", func() {
	It("LapCountMaximum is expected", func() {
		Expect(physical.LapCountMaximum).To(Equal(10000))
	})

	It("LapCountMinimum is expected", func() {
		Expect(physical.LapCountMinimum).To(Equal(0))
	})

	Context("ParseLap", func() {
		// TODO
	})

	Context("NewLap", func() {
		It("returns the expected datum", func() {
			Expect(physical.NewLap()).To(Equal(&physical.Lap{}))
		})
	})

	Context("Lap", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *physical.Lap), expectedErrors ...error) {
					datum := NewLap()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *physical.Lap) {},
				),
				Entry("count missing",
					func(datum *physical.Lap) { datum.Count = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/count"),
				),
				Entry("count out of range (lower)",
					func(datum *physical.Lap) { datum.Count = pointer.FromInt(-1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 10000), "/count"),
				),
				Entry("count in range (lower)",
					func(datum *physical.Lap) { datum.Count = pointer.FromInt(0) },
				),
				Entry("count in range (upper)",
					func(datum *physical.Lap) { datum.Count = pointer.FromInt(10000) },
				),
				Entry("count out of range (upper)",
					func(datum *physical.Lap) { datum.Count = pointer.FromInt(10001) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(10001, 0, 10000), "/count"),
				),
				Entry("distance missing",
					func(datum *physical.Lap) { datum.Distance = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/distance"),
				),
				Entry("distance invalid",
					func(datum *physical.Lap) { datum.Distance.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/distance/units"),
				),
				Entry("distance valid",
					func(datum *physical.Lap) { datum.Distance = NewDistance() },
				),
				Entry("multiple errors",
					func(datum *physical.Lap) {
						datum.Count = nil
						datum.Distance = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/count"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/distance"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *physical.Lap)) {
					for _, origin := range structure.Origins() {
						datum := NewLap()
						mutator(datum)
						expectedDatum := CloneLap(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *physical.Lap) {},
				),
				Entry("does not modify the datum; count missing",
					func(datum *physical.Lap) { datum.Count = nil },
				),
				Entry("does not modify the datum; distance missing",
					func(datum *physical.Lap) { datum.Distance = nil },
				),
			)
		})
	})
})
