package pumpstatus_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/test"

	"github.com/tidepool-org/platform/data/types/pumpstatus"

	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func RandomBattery() *pumpstatus.Battery {
	datum := *pumpstatus.NewBattery()
	datum.Unit = pointer.FromString("grams")
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(pumpstatus.MinBatteryPercentage, pumpstatus.MaxBatteryPercentage))
	return &datum
}

var _ = Describe("Battery", func() {

	Context("Battery", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",

				func(mutator func(datum *pumpstatus.Battery), expectedErrors ...error) {
					datum := RandomBattery()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pumpstatus.Battery) {},
				),
				Entry("Unit missing",
					func(datum *pumpstatus.Battery) { datum.Unit = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/unit"),
				),
				Entry("Value missing",
					func(datum *pumpstatus.Battery) { datum.Value = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("Value belove Minimum",
					func(datum *pumpstatus.Battery) {
						datum.Value = pointer.FromFloat64(pumpstatus.MinBatteryPercentage - 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(pumpstatus.MinBatteryPercentage-1, pumpstatus.MinBatteryPercentage, pumpstatus.MaxBatteryPercentage), "/value"),
				),
				Entry("Value above Maximum",
					func(datum *pumpstatus.Battery) {
						datum.Value = pointer.FromFloat64(pumpstatus.MaxBatteryPercentage + 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(pumpstatus.MaxBatteryPercentage+1, pumpstatus.MinBatteryPercentage, pumpstatus.MaxBatteryPercentage), "/value"),
				),
				Entry("Multiple Errors",
					func(datum *pumpstatus.Battery) {
						datum.Unit = nil
						datum.Value = pointer.FromFloat64(pumpstatus.MaxBatteryPercentage + 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/unit"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(pumpstatus.MaxBatteryPercentage+1, pumpstatus.MinBatteryPercentage, pumpstatus.MaxBatteryPercentage), "/value"),
				),
			)
		})
	})
})
