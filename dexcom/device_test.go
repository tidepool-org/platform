package dexcom_test

import (
	. "github.com/onsi/ginkgo"
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

	Describe("Validate", func() {
		Describe("requires", func() {
			It("lastUploadDate", func() {
				device := test.RandomDevice()
				device.LastUploadDate = nil
				validator := validator.New()
				device.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			})
			It("transmitterGeneration", func() {
				device := test.RandomDevice()
				device.TransmitterGeneration = nil
				validator := validator.New()
				device.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			})
			It("displayDevice", func() {
				device := test.RandomDevice()
				device.DisplayDevice = nil
				validator := validator.New()
				device.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			})
			It("alertSchedules", func() {
				device := test.RandomDevice()
				device.AlertScheduleList = nil
				validator := validator.New()
				device.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			})
		})
		Describe("does not require", func() {
			It("transmitterId", func() {
				device := test.RandomDevice()
				device.TransmitterID = nil
				validator := validator.New()
				device.Validate(validator)
				Expect(validator.Error()).ToNot(HaveOccurred())
			})
			It("displayApp", func() {
				device := test.RandomDevice()
				device.DisplayApp = nil
				validator := validator.New()
				device.Validate(validator)
				Expect(validator.Error()).ToNot(HaveOccurred())
			})
		})
	})
})
