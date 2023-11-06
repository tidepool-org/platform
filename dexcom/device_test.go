package dexcom_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/dexcom/test"
	"github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Device", func() {

	It("DeviceDisplayDevices returns expected", func() {
		Expect(dexcom.DeviceDisplayDevices()).To(Equal([]string{"android", "iOS", "receiver", "shareReceiver", "touchscreenReceiver"}))
		Expect(dexcom.DeviceDisplayDevices()).To(Equal([]string{
			dexcom.DeviceDisplayDeviceAndroid,
			dexcom.DeviceDisplayDeviceIOS,
			dexcom.DeviceDisplayDeviceReceiver,
			dexcom.DeviceDisplayDeviceShareReceiver,
			dexcom.DeviceDisplayDeviceTouchscreenReceiver,
		}))
	})

	It("DeviceTransmitterGenerations returns expected", func() {
		Expect(dexcom.DeviceTransmitterGenerations()).To(Equal([]string{"unknown", "g4", "g5", "g6", "g6 pro", "g6+", "dexcomPro", "g7"}))
		Expect(dexcom.DeviceTransmitterGenerations()).To(Equal([]string{
			dexcom.DeviceTransmitterGenerationUnknown,
			dexcom.DeviceTransmitterGenerationG4,
			dexcom.DeviceTransmitterGenerationG5,
			dexcom.DeviceTransmitterGenerationG6,
			dexcom.DeviceTransmitterGenerationG6Pro,
			dexcom.DeviceTransmitterGenerationG6Plus,
			dexcom.DeviceTransmitterGenerationPro,
			dexcom.DeviceTransmitterGenerationG7,
		}))
	})

	Describe("Validate", func() {
		DescribeTable("errors when",
			func(setupDeviceFunc func() *dexcom.Device) {
				testDevice := setupDeviceFunc()
				validator := validator.New()
				testDevice.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			},
			Entry("required lastUploadDate is not set", func() *dexcom.Device {
				device := test.RandomDevice()
				device.LastUploadDate = nil
				return device
			}),
			Entry("required alertSchedules is not set", func() *dexcom.Device {
				device := test.RandomDevice()
				device.AlertScheduleList = nil
				return device
			}),
			Entry("required transmitterGeneration is not set", func() *dexcom.Device {
				device := test.RandomDevice()
				device.TransmitterGeneration = nil
				return device
			}),
			Entry("required displayDevice is not set", func() *dexcom.Device {
				device := test.RandomDevice()
				device.DisplayDevice = nil
				return device
			}),
		)
		DescribeTable("does not error when",
			func(setupDeviceFunc func() *dexcom.Device) {
				testDevice := setupDeviceFunc()
				validator := validator.New()
				testDevice.Validate(validator)
				Expect(validator.Error()).ToNot(HaveOccurred())
			},
			Entry("transmitterID is not set", func() *dexcom.Device {
				device := test.RandomDevice()
				device.TransmitterID = nil
				return device
			}),
			Entry("displayApp is not set", func() *dexcom.Device {
				device := test.RandomDevice()
				device.DisplayApp = nil
				return device
			}),
		)
	})
})
