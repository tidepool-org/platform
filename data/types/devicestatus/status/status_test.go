package status_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/devicestatus/status"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/structure"
	//errorsTest "github.com/tidepool-org/platform/errors/test"
	//structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewStatusArray() *status.TypeStatusArray {
	batteryStatus := NewStatus()
	batteryStatus.Battery = NewBattery()

	reservoirRemainingStatus := NewStatus()
	reservoirRemainingStatus.ReservoirRemaining = NewReservoirRemaining()

	signalStrengthStatus := NewStatus()
	signalStrengthStatus.SignalStrength = NewSignalStrength()

	datum := status.TypeStatusArray{batteryStatus, reservoirRemainingStatus, signalStrengthStatus}
	return &datum
}

func NewStatus() *status.Status {
	datum := *status.NewStatus()
	return &datum
}

func CloneStatusArray(datum *status.TypeStatusArray) *status.TypeStatusArray {
	if datum == nil {
		return nil
	}
	clone := status.NewParseStatusArray()
	return clone
}

var _ = Describe("Status", func() {

	Context("Status", func() {
		Context("Validate", func() {
			DescribeTable("return the expected results when the input",

				func(mutator func(datum *status.TypeStatusArray), expectedErrors ...error) {
					datum := NewStatusArray()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *status.TypeStatusArray) {},
				),
				Entry("Missing Unit on Battery",
					func(datum *status.TypeStatusArray) {
						(*datum)[0].Battery = nil
					},
				),
			)
		})

	})
})
