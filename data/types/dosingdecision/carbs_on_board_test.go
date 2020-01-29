package dosingdecision_test

import (
	"time"

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

func NewCarbsOnBoard() *dosingdecision.CarbsOnBoard {
	datum := dosingdecision.NewCarbsOnBoard()

	datum.Quantity = pointer.FromFloat64(test.RandomFloat64FromRange(dosingdecision.MinCarbsOnBoard, dosingdecision.MaxCarbsOnBoard))
	datum.StartDate = pointer.FromString(test.FutureNearTime().Format(time.RFC3339Nano))
	datum.EndDate = pointer.FromString(test.FutureFarTime().Format(time.RFC3339Nano))
	return datum
}

var _ = Describe("CarbsOnBoard", func() {
	Context("Target", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",

				func(mutator func(datum *dosingdecision.CarbsOnBoard), expectedErrors ...error) {
					datum := NewCarbsOnBoard()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dosingdecision.CarbsOnBoard) {},
				),
				Entry("CarbsOnBoard below Minimum",
					func(datum *dosingdecision.CarbsOnBoard) {
						datum.Quantity = pointer.FromFloat64(dosingdecision.MinCarbsOnBoard - 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dosingdecision.MinCarbsOnBoard-1, dosingdecision.MinCarbsOnBoard, dosingdecision.MaxCarbsOnBoard), "/quantity"),
				),
				Entry("CarbsOnBoard above Maximum",
					func(datum *dosingdecision.CarbsOnBoard) {
						datum.Quantity = pointer.FromFloat64(dosingdecision.MaxCarbsOnBoard + 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dosingdecision.MaxCarbsOnBoard+1, dosingdecision.MinCarbsOnBoard, dosingdecision.MaxCarbsOnBoard), "/quantity"),
				),
				Entry("start date time invalid",
					func(datum *dosingdecision.CarbsOnBoard) {
						datum.StartDate = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339Nano), "/startDate"),
				),
				Entry("end date time invalid",
					func(datum *dosingdecision.CarbsOnBoard) {
						datum.EndDate = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339Nano), "/endDate"),
				),
				Entry("end date not after start date",
					func(datum *dosingdecision.CarbsOnBoard) {
						datum.EndDate = pointer.FromString(test.PastFarTime().Format(time.RFC3339Nano))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(test.PastFarTime(), test.FutureNearTime()), "/endDate"),
				),
				Entry("end date equals start date - no error",
					func(datum *dosingdecision.CarbsOnBoard) {
						datum.EndDate = datum.StartDate
					},
				),
			)
		})
	})
})
