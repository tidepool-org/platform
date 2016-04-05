package data

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"

	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("DeviceEvent", func() {

	var processing validate.ErrorProcessing

	Context("calibration", func() {

		var deviceEventObj = testingDatumBase()
		deviceEventObj["type"] = "deviceEvent"
		deviceEventObj["subType"] = "calibration"
		deviceEventObj["value"] = 3.0
		deviceEventObj["units"] = "mg/dL"

		It("returns a CalibrationDeviceEvent if the obj is valid", func() {
			deviceEvent := BuildDeviceEvent(deviceEventObj, processing)
			var deviceEventType *CalibrationDeviceEvent
			Expect(deviceEvent).To(BeAssignableToTypeOf(deviceEventType))
			Expect(processing.HasErrors()).To(BeFalse())
		})

		Context("validation", func() {})
	})
	Context("status", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0/deviceEvent", ErrorsArray: validate.NewErrorsArray()}
		})

		var deviceEventObj = testingDatumBase()
		deviceEventObj["type"] = "deviceEvent"
		deviceEventObj["subType"] = "status"
		deviceEventObj["status"] = "suspended"
		deviceEventObj["reason"] = map[string]string{"suspended": "automatic"}
		It("returns a StatusDeviceEvent if the obj is valid", func() {
			deviceEvent := BuildDeviceEvent(deviceEventObj, processing)
			var deviceEventType *StatusDeviceEvent
			Expect(deviceEvent).To(BeAssignableToTypeOf(deviceEventType))
			Expect(processing.HasErrors()).To(BeFalse())
		})

		Context("validation", func() {})
	})
	Context("alarm", func() {
		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0/deviceEvent", ErrorsArray: validate.NewErrorsArray()}
		})
		var deviceEventObj = testingDatumBase()
		deviceEventObj["type"] = "deviceEvent"
		deviceEventObj["subType"] = "alarm"
		deviceEventObj["alarmType"] = "low_insulin"

		It("returns a AlarmDeviceEvent if the obj is valid", func() {
			deviceEvent := BuildDeviceEvent(deviceEventObj, processing)
			var deviceEventType *AlarmDeviceEvent
			Expect(deviceEvent).To(BeAssignableToTypeOf(deviceEventType))
			Expect(processing.HasErrors()).To(BeFalse())
		})

		Context("validation", func() {})
	})
	Context("prime", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0/deviceEvent", ErrorsArray: validate.NewErrorsArray()}
		})

		var deviceEventObj = testingDatumBase()
		deviceEventObj["type"] = "deviceEvent"
		deviceEventObj["subType"] = "prime"
		deviceEventObj["primeTarget"] = "cannula"

		It("returns a PrimeDeviceEvent if the obj is valid", func() {
			deviceEvent := BuildDeviceEvent(deviceEventObj, processing)
			var deviceEventType *PrimeDeviceEvent
			Expect(deviceEvent).To(BeAssignableToTypeOf(deviceEventType))
			Expect(processing.HasErrors()).To(BeFalse())
		})

		Context("validation", func() {})
	})
	Context("timeChange", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0/deviceEvent", ErrorsArray: validate.NewErrorsArray()}
		})

		var deviceEventObj = testingDatumBase()
		deviceEventObj["type"] = "deviceEvent"
		deviceEventObj["subType"] = "timeChange"
		deviceEventObj["change"] = map[string]interface{}{
			"from":     "2015-03-08T12:02:00",
			"to":       "2015-03-08T13:00:00",
			"agent":    "manual",
			"reasons":  []string{"to_daylight_savings", "correction"},
			"timezone": "US/Pacific",
		}

		It("returns a TimeChangeDeviceEvent if the obj is valid", func() {
			deviceEvent := BuildDeviceEvent(deviceEventObj, processing)
			var deviceEventType *TimeChangeDeviceEvent
			Expect(deviceEvent).To(BeAssignableToTypeOf(deviceEventType))
			Expect(processing.HasErrors()).To(BeFalse())
		})

		Context("validation", func() {})
	})
})
