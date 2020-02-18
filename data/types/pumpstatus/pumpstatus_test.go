package pumpstatus_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/pumpstatus"
	"github.com/tidepool-org/platform/structure"

	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
)

func RandomPumpStatus() *pumpstatus.PumpStatus {
	datum := *pumpstatus.NewPumpStatus()

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
			)
		})
	})
})
