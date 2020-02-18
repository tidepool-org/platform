package dosingdecision_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/dosingdecision"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/test"

	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewRecommendedBasal() *dosingdecision.RecommendedBasal {
	datum := dosingdecision.NewRecommendedBasal()

	datum.UnitsPerHour = pointer.FromFloat64(test.RandomFloat64FromRange(dosingdecision.MinUnitsPerHour, dosingdecision.MaxUnitsPerHour))
	datum.Duration = pointer.FromFloat64(test.RandomFloat64FromRange(dosingdecision.MinDuration, dosingdecision.MaxDuration))
	return datum
}

var _ = Describe("RecommendedBasal", func() {
	Context("Target", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",

				func(mutator func(datum *dosingdecision.RecommendedBasal), expectedErrors ...error) {
					datum := NewRecommendedBasal()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dosingdecision.RecommendedBasal) {},
				),
				Entry("Duration below Minimum",
					func(datum *dosingdecision.RecommendedBasal) {
						datum.Duration = pointer.FromFloat64(dosingdecision.MinDuration - 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dosingdecision.MinDuration-1, dosingdecision.MinDuration, dosingdecision.MaxDuration), "/duration"),
				),
				Entry("Duration above Maximum",
					func(datum *dosingdecision.RecommendedBasal) {
						datum.Duration = pointer.FromFloat64(dosingdecision.MaxDuration + 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dosingdecision.MaxDuration+1, dosingdecision.MinDuration, dosingdecision.MaxDuration), "/duration"),
				),
				Entry("Units per hour below Minimum",
					func(datum *dosingdecision.RecommendedBasal) {
						datum.UnitsPerHour = pointer.FromFloat64(dosingdecision.MinUnitsPerHour - 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dosingdecision.MinUnitsPerHour-1, dosingdecision.MinUnitsPerHour, dosingdecision.MaxUnitsPerHour), "/unitsPerHour"),
				),
				Entry("Units per hour above Maximum",
					func(datum *dosingdecision.RecommendedBasal) {
						datum.UnitsPerHour = pointer.FromFloat64(dosingdecision.MaxUnitsPerHour + 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dosingdecision.MaxUnitsPerHour+1, dosingdecision.MinUnitsPerHour, dosingdecision.MaxUnitsPerHour), "/unitsPerHour"),
				),
			)
		})
	})
})
