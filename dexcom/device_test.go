package dexcom_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dexcom"
)

var _ = Describe("Device", func() {
	It("DeviceDisplayDeviceAndroid is expected", func() {
		Expect(dexcom.DeviceDisplayDeviceAndroid).To(Equal("android"))
	})

	It("DeviceDisplayDeviceIOS is expected", func() {
		Expect(dexcom.DeviceDisplayDeviceIOS).To(Equal("iOS"))
	})

	It("DeviceDisplayDeviceReceiver is expected", func() {
		Expect(dexcom.DeviceDisplayDeviceReceiver).To(Equal("receiver"))
	})

	It("DeviceDisplayDeviceShareReceiver is expected", func() {
		Expect(dexcom.DeviceDisplayDeviceShareReceiver).To(Equal("shareReceiver"))
	})

	It("DeviceDisplayDeviceTouchscreenReceiver is expected", func() {
		Expect(dexcom.DeviceDisplayDeviceTouchscreenReceiver).To(Equal("touchscreenReceiver"))
	})

	It("DeviceTransmitterGenerationG4 is expected", func() {
		Expect(dexcom.DeviceTransmitterGenerationG4).To(Equal("g4"))
	})

	It("DeviceTransmitterGenerationG5 is expected", func() {
		Expect(dexcom.DeviceTransmitterGenerationG5).To(Equal("g5"))
	})

	It("DeviceTransmitterGenerationG6 is expected", func() {
		Expect(dexcom.DeviceTransmitterGenerationG6).To(Equal("g6"))
	})

	It("DeviceTransmitterGenerationG6Pro is expected", func() {
		Expect(dexcom.DeviceTransmitterGenerationG6Pro).To(Equal("g6 pro"))
	})

	It("DeviceTransmitterGenerationG7 is expected", func() {
		Expect(dexcom.DeviceTransmitterGenerationG7).To(Equal("g7"))
	})

	It("DeviceDisplayDevices returns expected", func() {
		Expect(dexcom.DeviceDisplayDevices()).To(Equal([]string{"android", "iOS", "receiver", "shareReceiver", "touchscreenReceiver"}))
	})

	It("DeviceTransmitterGenerations returns expected", func() {
		Expect(dexcom.DeviceTransmitterGenerations()).To(Equal([]string{"g4", "g5", "g6", "g6 pro", "g7"}))
	})
})
