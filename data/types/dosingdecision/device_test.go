package dosingdecision_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/dosingdecision"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func RandomDevice() *dosingdecision.Device {
	datum := dosingdecision.NewDevice()

	datum.DeviceID = pointer.FromString("DeviceID")
	datum.Manufacturer = pointer.FromString("Manufacturer")
	datum.Model = pointer.FromString("Model")
	return datum
}

var _ = Describe("Device", func() {
	Context("Target", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",

				func(mutator func(datum *dosingdecision.Device), expectedErrors ...error) {
					datum := RandomDevice()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dosingdecision.Device) {},
				),
				Entry("Invalid DeviceID",
					func(datum *dosingdecision.Device) {
						datum.DeviceID = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotGreaterThanOrEqualTo(0, 1), "/deviceID"),
				),
				Entry("Invalid Manufacturer",
					func(datum *dosingdecision.Device) {
						datum.Manufacturer = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotGreaterThanOrEqualTo(0, 1), "/manufacturer"),
				),
				Entry("Invalid Model",
					func(datum *dosingdecision.Device) {
						datum.Model = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotGreaterThanOrEqualTo(0, 1), "/model"),
				),
			)
		})
	})
})
