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

func RandomInsulinOnBoard() *dosingdecision.InsulinOnBoard {
	datum := dosingdecision.NewInsulinOnBoard()

	datum.StartDate = pointer.FromString(test.FutureNearTime().Format(time.RFC3339Nano))
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dosingdecision.MinInsulinOnBoard, dosingdecision.MaxInsulinOnBoard)) // Any positive number
	return datum
}

var _ = Describe("InsulinOnBoard", func() {
	Context("Target", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",

				func(mutator func(datum *dosingdecision.InsulinOnBoard), expectedErrors ...error) {
					datum := RandomInsulinOnBoard()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dosingdecision.InsulinOnBoard) {},
				),
				Entry("Start Date missing",
					func(datum *dosingdecision.InsulinOnBoard) { datum.StartDate = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/startDate"),
				),
				Entry("Value missing",
					func(datum *dosingdecision.InsulinOnBoard) { datum.Value = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("Value below Minimum",
					func(datum *dosingdecision.InsulinOnBoard) { datum.Value = pointer.FromFloat64(dosingdecision.MinInsulinOnBoard - 1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dosingdecision.MinInsulinOnBoard-1, dosingdecision.MinInsulinOnBoard, dosingdecision.MaxInsulinOnBoard), "/value"),
				),
				Entry("Value above Maximum",
					func(datum *dosingdecision.InsulinOnBoard) { datum.Value = pointer.FromFloat64(dosingdecision.MaxInsulinOnBoard + 1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dosingdecision.MaxInsulinOnBoard+1, dosingdecision.MinInsulinOnBoard, dosingdecision.MaxInsulinOnBoard), "/value"),
				),
			)
		})
	})

})
