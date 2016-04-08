package device

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("DeviceEvent", func() {

	var processing validate.ErrorProcessing

	Context("prime", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
		})

		var deviceEventObj = fixtures.TestingDatumBase()
		deviceEventObj["type"] = "deviceEvent"
		deviceEventObj["subType"] = "prime"
		deviceEventObj["primeTarget"] = "cannula"

		It("returns a PrimeDeviceEvent if the obj is valid", func() {
			deviceEvent := Build(deviceEventObj, processing)
			var deviceEventType *Prime
			Expect(deviceEvent).To(BeAssignableToTypeOf(deviceEventType))
			Expect(processing.HasErrors()).To(BeFalse())
		})

		Context("validation", func() {})
	})

})
