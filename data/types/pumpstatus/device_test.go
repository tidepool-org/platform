package pumpstatus_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataTypesPumpStatus "github.com/tidepool-org/platform/data/types/pumpstatus"
	dataTypesPumpStatusTest "github.com/tidepool-org/platform/data/types/pumpstatus/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Device", func() {
	It("DeviceIDLengthMaximum is expected", func() {
		Expect(dataTypesPumpStatus.DeviceIDLengthMaximum).To(Equal(1000))
	})

	It("DeviceIDLengthMinimum is expected", func() {
		Expect(dataTypesPumpStatus.DeviceIDLengthMinimum).To(Equal(1))
	})

	It("DeviceNameLengthMaximum is expected", func() {
		Expect(dataTypesPumpStatus.DeviceNameLengthMaximum).To(Equal(1000))
	})

	It("DeviceNameLengthMinimum is expected", func() {
		Expect(dataTypesPumpStatus.DeviceNameLengthMinimum).To(Equal(1))
	})

	It("DeviceManufacturerLengthMaximum is expected", func() {
		Expect(dataTypesPumpStatus.DeviceManufacturerLengthMaximum).To(Equal(1000))
	})

	It("DeviceManufacturerLengthMinimum is expected", func() {
		Expect(dataTypesPumpStatus.DeviceManufacturerLengthMinimum).To(Equal(1))
	})

	It("DeviceModelLengthMaximum is expected", func() {
		Expect(dataTypesPumpStatus.DeviceModelLengthMaximum).To(Equal(1000))
	})

	It("DeviceModelLengthMinimum is expected", func() {
		Expect(dataTypesPumpStatus.DeviceModelLengthMinimum).To(Equal(1))
	})

	It("DeviceVersionLengthMaximum is expected", func() {
		Expect(dataTypesPumpStatus.DeviceVersionLengthMaximum).To(Equal(100))
	})

	It("DeviceVersionLengthMinimum is expected", func() {
		Expect(dataTypesPumpStatus.DeviceVersionLengthMinimum).To(Equal(1))
	})

	Context("ParseDevice", func() {
		// TODO
	})

	Context("NewDevice", func() {
		It("is successful", func() {
			Expect(dataTypesPumpStatus.NewDevice()).To(Equal(&dataTypesPumpStatus.Device{}))
		})
	})

	Context("Device", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(mutator func(datum *dataTypesPumpStatus.Device), expectedErrors ...error) {
					datum := dataTypesPumpStatusTest.RandomDevice()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesPumpStatus.Device) {},
				),
				Entry("id length below minimum",
					func(datum *dataTypesPumpStatus.Device) { datum.ID = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(0, 1, 1000), "/id"),
				),
				Entry("id length above maximum",
					func(datum *dataTypesPumpStatus.Device) {
						datum.ID = pointer.FromString(test.RandomStringFromRange(1001, 1001))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(1001, 1, 1000), "/id"),
				),
				Entry("name length below minimum",
					func(datum *dataTypesPumpStatus.Device) { datum.Name = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(0, 1, 1000), "/name"),
				),
				Entry("name length above maximum",
					func(datum *dataTypesPumpStatus.Device) {
						datum.Name = pointer.FromString(test.RandomStringFromRange(1001, 1001))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(1001, 1, 1000), "/name"),
				),

				Entry("manufacturer length below minimum",
					func(datum *dataTypesPumpStatus.Device) { datum.Manufacturer = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(0, 1, 1000), "/manufacturer"),
				),
				Entry("manufacturer length above maximum",
					func(datum *dataTypesPumpStatus.Device) {
						datum.Manufacturer = pointer.FromString(test.RandomStringFromRange(1001, 1001))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(1001, 1, 1000), "/manufacturer"),
				),
				Entry("model length below minimum",
					func(datum *dataTypesPumpStatus.Device) { datum.Model = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(0, 1, 1000), "/model"),
				),
				Entry("model length above maximum",
					func(datum *dataTypesPumpStatus.Device) {
						datum.Model = pointer.FromString(test.RandomStringFromRange(1001, 1001))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(1001, 1, 1000), "/model"),
				),
				Entry("firmwareVersion length below minimum",
					func(datum *dataTypesPumpStatus.Device) { datum.FirmwareVersion = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(0, 1, 100), "/firmwareVersion"),
				),
				Entry("firmwareVersion length above maximum",
					func(datum *dataTypesPumpStatus.Device) {
						datum.FirmwareVersion = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(101, 1, 100), "/firmwareVersion"),
				),
				Entry("hardwareVersion length below minimum",
					func(datum *dataTypesPumpStatus.Device) { datum.HardwareVersion = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(0, 1, 100), "/hardwareVersion"),
				),
				Entry("hardwareVersion length above maximum",
					func(datum *dataTypesPumpStatus.Device) {
						datum.HardwareVersion = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(101, 1, 100), "/hardwareVersion"),
				),
				Entry("softwareVersion length below minimum",
					func(datum *dataTypesPumpStatus.Device) { datum.SoftwareVersion = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(0, 1, 100), "/softwareVersion"),
				),
				Entry("softwareVersion length above maximum",
					func(datum *dataTypesPumpStatus.Device) {
						datum.SoftwareVersion = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(101, 1, 100), "/softwareVersion"),
				),
				Entry("multiple errors",
					func(datum *dataTypesPumpStatus.Device) {
						datum.ID = pointer.FromString("")
						datum.Name = pointer.FromString("")
						datum.Manufacturer = pointer.FromString("")
						datum.Model = pointer.FromString("")
						datum.FirmwareVersion = pointer.FromString("")
						datum.HardwareVersion = pointer.FromString("")
						datum.SoftwareVersion = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(0, 1, 1000), "/id"),
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(0, 1, 1000), "/name"),
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(0, 1, 1000), "/manufacturer"),
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(0, 1, 1000), "/model"),
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(0, 1, 100), "/firmwareVersion"),
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(0, 1, 100), "/hardwareVersion"),
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(0, 1, 100), "/softwareVersion"),
				),
			)
		})
	})
})
