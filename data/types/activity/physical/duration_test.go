package physical_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"math"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/activity/physical"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func ValueRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case physical.UnitsHours:
			return 0.0, 24.0
		case physical.UnitsMinutes:
			return 0.0, 24.0 * 60.0
		case physical.UnitsSeconds:
			return 0.0, 24.0 * 60.0 * 60.0
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}

func NewDuration() *physical.Duration {
	datum := physical.NewDuration()
	datum.Units = pointer.String(test.RandomStringFromArray(physical.Units()))
	datum.Value = pointer.Float64(test.RandomFloat64FromRange(ValueRangeForUnits(datum.Units)))
	return datum
}

func CloneDuration(datum *physical.Duration) *physical.Duration {
	if datum == nil {
		return nil
	}
	clone := physical.NewDuration()
	clone.Units = test.CloneString(datum.Units)
	clone.Value = test.CloneFloat64(datum.Value)
	return clone
}

var _ = Describe("Duration", func() {
	It("UnitsHours is expected", func() {
		Expect(physical.UnitsHours).To(Equal("hours"))
	})

	It("UnitsMinutes is expected", func() {
		Expect(physical.UnitsMinutes).To(Equal("minutes"))
	})

	It("UnitsSeconds is expected", func() {
		Expect(physical.UnitsSeconds).To(Equal("seconds"))
	})

	It("ValueMinimum is expected", func() {
		Expect(physical.ValueMinimum).To(Equal(0))
	})

	It("Units returns expected", func() {
		Expect(physical.Units()).To(Equal([]string{"hours", "minutes", "seconds"}))
	})

	Context("ParseDuration", func() {
		// TODO
	})

	Context("NewDuration", func() {
		It("returns the expected datum", func() {
			Expect(physical.NewDuration()).To(Equal(&physical.Duration{}))
		})
	})

	Context("Duration", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *physical.Duration), expectedErrors ...error) {
					datum := NewDuration()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *physical.Duration) {},
				),
				Entry("units missing",
					func(datum *physical.Duration) { datum.Units = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid",
					func(datum *physical.Duration) { datum.Units = pointer.String("invalid") },
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours", "minutes", "seconds"}), "/units"),
				),
				Entry("units hours",
					func(datum *physical.Duration) { datum.Units = pointer.String("hours") },
				),
				Entry("units minutes",
					func(datum *physical.Duration) { datum.Units = pointer.String("minutes") },
				),
				Entry("units seconds",
					func(datum *physical.Duration) { datum.Units = pointer.String("seconds") },
				),
				Entry("value missing",
					func(datum *physical.Duration) { datum.Value = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("value out of range (lower)",
					func(datum *physical.Duration) { datum.Value = pointer.Float64(0.0) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotGreaterThan(0, 0), "/value"),
				),
				Entry("value in range (lower)",
					func(datum *physical.Duration) { datum.Value = pointer.Float64(0.1) },
				),
				Entry("multiple errors",
					func(datum *physical.Duration) {
						datum.Units = nil
						datum.Value = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *physical.Duration)) {
					for _, origin := range structure.Origins() {
						datum := NewDuration()
						mutator(datum)
						expectedDatum := CloneDuration(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *physical.Duration) {},
				),
				Entry("does not modify the datum; units missing",
					func(datum *physical.Duration) { datum.Units = nil },
				),
				Entry("does not modify the datum; units hours",
					func(datum *physical.Duration) { datum.Units = pointer.String("hours") },
				),
				Entry("does not modify the datum; units minutes",
					func(datum *physical.Duration) { datum.Units = pointer.String("minutes") },
				),
				Entry("does not modify the datum; units seconds",
					func(datum *physical.Duration) { datum.Units = pointer.String("seconds") },
				),
				Entry("does not modify the datum; value missing",
					func(datum *physical.Duration) { datum.Value = nil },
				),
			)
		})
	})
})
