package device

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"

	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("DeviceEvent", func() {

	var processing validate.ErrorProcessing

	Context("timeChange", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
		})

		var deviceEventObj = TestingDatumBase()
		deviceEventObj["type"] = "deviceEvent"
		deviceEventObj["subType"] = "timeChange"
		deviceEventObj["change"] = map[string]interface{}{
			"from":     "2015-03-08T12:02:00",
			"to":       "2015-03-08T13:00:00",
			"agent":    "manual",
			"reasons":  []string{"to_daylight_savings", "correction"},
			"timezone": "US/Pacific",
		}

		It("returns a TimeChange if the obj is valid", func() {
			deviceEvent := Build(deviceEventObj, processing)
			var deviceEventType *TimeChange
			Expect(deviceEvent).To(BeAssignableToTypeOf(deviceEventType))
			Expect(processing.HasErrors()).To(BeFalse())
		})

		Context("validation", func() {})
	})
})
