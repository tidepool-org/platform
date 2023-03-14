package dexcom_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dexcom"
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
		Expect(dexcom.DeviceTransmitterGenerations()).To(Equal([]string{"unknown", "g4", "g5", "g6", "g6+", "dexcomPro", "g7"}))
		Expect(dexcom.DeviceTransmitterGenerations()).To(Equal([]string{
			dexcom.DeviceTransmitterGenerationUnknown,
			dexcom.DeviceTransmitterGenerationG4,
			dexcom.DeviceTransmitterGenerationG5,
			dexcom.DeviceTransmitterGenerationG6,
			dexcom.DeviceTransmitterGenerationG6Plus,
			dexcom.DeviceTransmitterGenerationPro,
			dexcom.DeviceTransmitterGenerationG7,
		}))
	})
})
