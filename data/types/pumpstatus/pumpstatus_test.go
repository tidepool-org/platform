package pumpstatus_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"

	"github.com/tidepool-org/platform/data/types/pumpstatus"
	"github.com/tidepool-org/platform/structure"

	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func RandomPumpStatus() *pumpstatus.PumpStatus {
	datum := *pumpstatus.NewPumpStatus()
	datum.ReservoirRemaining = pointer.FromFloat64(test.RandomFloat64FromRange(pumpstatus.MinReservoirRemaining, pumpstatus.MaxReservoirRemaining))

	return &datum
}

var _ = Describe("PumpStatus", func() {
	Context("BasalDeliveryState", func() {
		Context("Validate", func() {
			DescribeTable("return the expected results when the input",

				func(mutator func(datum *pumpstatus.PumpStatus), expectedErrors ...error) {
					datum := RandomPumpStatus()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pumpstatus.PumpStatus) {},
				),
				Entry("ReservoirRemaing belove Minimum",
					func(datum *pumpstatus.PumpStatus) {
						datum.ReservoirRemaining = pointer.FromFloat64(pumpstatus.MinReservoirRemaining - 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(pumpstatus.MinReservoirRemaining-1, pumpstatus.MinReservoirRemaining, pumpstatus.MaxReservoirRemaining), "/reservoirRemaining"),
				),
				Entry("ReservoirRemaing Value above Maximum",
					func(datum *pumpstatus.PumpStatus) {
						datum.ReservoirRemaining = pointer.FromFloat64(pumpstatus.MaxReservoirRemaining + 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(pumpstatus.MaxReservoirRemaining+1, pumpstatus.MinReservoirRemaining, pumpstatus.MaxReservoirRemaining), "/reservoirRemaining"),
				),
			)
		})
	})
})
