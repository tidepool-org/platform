package pumpstatus_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/pumpstatus"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/test"

	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func RandomPumpStatus() *pumpstatus.PumpStatus {
	datum := *pumpstatus.NewPumpStatus()
	datum.PumpBatteryChargeRemaining = pointer.FromFloat64(test.RandomFloat64FromRange(pumpstatus.MinPumpChargeRemaining, pumpstatus.MaxPumpChargeRemaining))

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
				Entry("PumpBatteryChargeRemaining below Minimum",
					func(datum *pumpstatus.PumpStatus) {
						datum.PumpBatteryChargeRemaining = pointer.FromFloat64(pumpstatus.MinPumpChargeRemaining - 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(pumpstatus.MinPumpChargeRemaining-1, pumpstatus.MinPumpChargeRemaining, pumpstatus.MaxPumpChargeRemaining), "/pumpBatteryChargeRemaining"),
				),
				Entry("PumpBatteryChargeRemaining above Maximum",
					func(datum *pumpstatus.PumpStatus) {
						datum.PumpBatteryChargeRemaining = pointer.FromFloat64(pumpstatus.MaxPumpChargeRemaining + 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(pumpstatus.MaxPumpChargeRemaining+1, pumpstatus.MinPumpChargeRemaining, pumpstatus.MaxPumpChargeRemaining), "/pumpBatteryChargeRemaining"),
				),
			)
		})
	})
})
