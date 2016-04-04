package data

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"

	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("DeviceEvent", func() {

	var deviceEventObj = testingDatumBase()
	deviceEventObj["type"] = "deviceEvent"
	deviceEventObj["subType"] = "alarm"

	var processing = validate.ErrorProcessing{BasePath: "0/deviceEvent", ErrorsArray: validate.NewErrorsArray()}

	Context("can be built from obj", func() {
		It("should return a basal if the obj is valid", func() {
			deviceEvent := BuildDeviceEvent(deviceEventObj, processing)
			var deviceEventType *DeviceEvent
			Expect(deviceEvent).To(BeAssignableToTypeOf(deviceEventType))
		})
		It("should produce no error when valid", func() {
			BuildDeviceEvent(deviceEventObj, processing)
			Expect(processing.HasErrors()).To(BeFalse())
		})
	})
})
