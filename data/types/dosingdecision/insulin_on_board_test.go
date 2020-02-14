package dosingdecision_test

import (
	"math"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/dosingdecision"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/test"

	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	//errorsTest "github.com/tidepool-org/platform/errors/test"
	//structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func RandomInsulinOnBoard() *dosingdecision.InsulinOnBoard {
	datum := dosingdecision.NewInsulinOnBoard()

	datum.StartDate = pointer.FromString(test.FutureNearTime().Format(time.RFC3339Nano))
	datum.Value = pointer.FromFloat64(math.Abs(test.RandomFloat64())) // Any positive number
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
				//Entry("end date not after start date",
				//	func(datum *dosingdecision.InsulinOnBoard) {
				//		datum.StartDate = pointer.FromString(test.PastFarTime().Format(time.RFC3339Nano))
				//	},
				//	errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(test.PastFarTime(), test.PastNearTime()), "/startDate"),
				//),
			)
		})
	})

})
