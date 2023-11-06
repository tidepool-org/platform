package physical_test

import (
	. "github.com/onsi/ginkgo/v2"
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

func NewStep() *physical.Step {
	datum := physical.NewStep()
	datum.Count = pointer.FromInt(test.RandomIntFromRange(0, 100000))
	return datum
}

func CloneStep(datum *physical.Step) *physical.Step {
	if datum == nil {
		return nil
	}
	clone := physical.NewStep()
	clone.Count = pointer.CloneInt(datum.Count)
	return clone
}

var _ = Describe("Step", func() {
	It("StepCountMaximum is expected", func() {
		Expect(physical.StepCountMaximum).To(Equal(100000))
	})

	It("StepCountMinimum is expected", func() {
		Expect(physical.StepCountMinimum).To(Equal(0))
	})

	Context("ParseStep", func() {
		// TODO
	})

	Context("NewStep", func() {
		It("returns the expected datum", func() {
			Expect(physical.NewStep()).To(Equal(&physical.Step{}))
		})
	})

	Context("Step", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *physical.Step), expectedErrors ...error) {
					datum := NewStep()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *physical.Step) {},
				),
				Entry("count missing",
					func(datum *physical.Step) { datum.Count = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/count"),
				),
				Entry("count out of range (lower)",
					func(datum *physical.Step) { datum.Count = pointer.FromInt(-1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 100000), "/count"),
				),
				Entry("count in range (lower)",
					func(datum *physical.Step) { datum.Count = pointer.FromInt(0) },
				),
				Entry("count in range (upper)",
					func(datum *physical.Step) { datum.Count = pointer.FromInt(100000) },
				),
				Entry("count out of range (upper)",
					func(datum *physical.Step) { datum.Count = pointer.FromInt(100001) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(100001, 0, 100000), "/count"),
				),
				Entry("multiple errors",
					func(datum *physical.Step) { datum.Count = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/count"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *physical.Step)) {
					for _, origin := range structure.Origins() {
						datum := NewStep()
						mutator(datum)
						expectedDatum := CloneStep(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *physical.Step) {},
				),
				Entry("does not modify the datum; count missing",
					func(datum *physical.Step) { datum.Count = nil },
				),
			)
		})
	})
})
