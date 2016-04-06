package device

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"

	"github.com/tidepool-org/platform/validate"
)

func TestingDatumBase() map[string]interface{} {
	return map[string]interface{}{
		"userId":           "b676436f60",
		"groupId":          "43099shgs55",
		"uploadId":         "upid_b856b0e6e519",
		"deviceTime":       "2014-06-11T06:00:00.000Z",
		"time":             "2014-06-11T06:00:00.000Z",
		"timezoneOffset":   0,
		"conversionOffset": 0,
		"clockDriftOffset": 0,
		"deviceId":         "InsOmn-111111111",
	}
}

var _ = Describe("DeviceEvent", func() {

	var processing validate.ErrorProcessing
	BeforeEach(func() {
		processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
	})

	Context("base", func() {

		Context("alarm subType", func() {
			var deviceEventObj = TestingDatumBase()
			deviceEventObj["type"] = "deviceEvent"
			deviceEventObj["subType"] = "alarm"
			deviceEventObj["alarmType"] = "low_insulin"

			It("returns a Alarm if the obj is valid", func() {
				deviceEvent := Build(deviceEventObj, processing)
				var deviceEventType *Alarm
				Expect(deviceEvent).To(BeAssignableToTypeOf(deviceEventType))
				Expect(processing.HasErrors()).To(BeFalse())
			})
		})

		Context("calibration subType", func() {

			var deviceEventObj = TestingDatumBase()
			deviceEventObj["type"] = "deviceEvent"
			deviceEventObj["subType"] = "calibration"
			deviceEventObj["value"] = 3.0
			deviceEventObj["units"] = "mg/dL"

			It("returns a Calibration if the obj is valid", func() {
				deviceEvent := Build(deviceEventObj, processing)
				var deviceEventType *Calibration
				Expect(deviceEvent).To(BeAssignableToTypeOf(deviceEventType))
				Expect(processing.HasErrors()).To(BeFalse())
			})
		})

	})
})
