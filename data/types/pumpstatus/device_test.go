package pumpstatus_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/pumpstatus"

	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func RandomDevice() *pumpstatus.Device {
	datum := *pumpstatus.NewDevice()
	datum.DeviceID = pointer.FromString("deviceID")
	datum.Model = pointer.FromString("model")
	datum.Manufacturer = pointer.FromString("manufacturer")
	return &datum
}

var _ = Describe("Device", func() {

	Context("Device", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",

				func(mutator func(datum *pumpstatus.Device), expectedErrors ...error) {
					datum := RandomDevice()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pumpstatus.Device) {},
				),
				Entry("DeviceID missing",
					func(datum *pumpstatus.Device) { datum.DeviceID = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/deviceID"),
				),
				Entry("DeviceID not long enough",
					func(datum *pumpstatus.Device) { datum.DeviceID = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(0, pumpstatus.MinDeviceIDLength, pumpstatus.MaxDeviceIDLength), "/deviceID"),
				),
				Entry("Model missing",
					func(datum *pumpstatus.Device) { datum.Model = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/model"),
				),
				Entry("Model not long enough",
					func(datum *pumpstatus.Device) { datum.Model = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(0, pumpstatus.MinModelLength, pumpstatus.MaxModelLength), "/model"),
				),
				Entry("Manufacturer missing",
					func(datum *pumpstatus.Device) { datum.Manufacturer = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/manufacturer"),
				),
				Entry("Manufacturer not long enough",
					func(datum *pumpstatus.Device) { datum.Manufacturer = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(0, pumpstatus.MinManufacturerLength, pumpstatus.MaxManufacturerLength), "/manufacturer"),
				),
				Entry("Multiple Errors",
					func(datum *pumpstatus.Device) {
						datum.DeviceID = nil
						datum.Model = nil
						datum.Manufacturer = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/deviceID"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/model"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/manufacturer"),
				),
			)
		})
	})
})
