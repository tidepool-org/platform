package device

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("DeviceEvent", func() {

	var processing validate.ErrorProcessing

	Context("status", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0/deviceEvent", ErrorsArray: validate.NewErrorsArray()}
		})

		var deviceEventObj = fixtures.TestingDatumBase()
		deviceEventObj["type"] = "deviceEvent"
		deviceEventObj["subType"] = "status"
		deviceEventObj["status"] = "suspended"
		deviceEventObj["reason"] = map[string]string{"suspended": "automatic"}
		It("returns a Status if the obj is valid", func() {
			deviceEvent := Build(deviceEventObj, processing)
			var deviceEventType *Status
			Expect(deviceEvent).To(BeAssignableToTypeOf(deviceEventType))
			Expect(processing.HasErrors()).To(BeFalse())
		})

		Context("validation", func() {})
	})
})
