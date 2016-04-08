package device

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("DeviceEvent", func() {

	var processing validate.ErrorProcessing
	BeforeEach(func() {
		processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
	})

	Context("base", func() {

		Context("alarm subType", func() {
			var deviceEventObj = fixtures.TestingDatumBase()
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

			var deviceEventObj = fixtures.TestingDatumBase()
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
