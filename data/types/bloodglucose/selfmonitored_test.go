package bloodglucose

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("Selfmonitored", func() {

	var bgObj = TestingDatumBase()
	bgObj["type"] = "smbg"
	bgObj["value"] = 5.5
	bgObj["units"] = "mmol/l"

	var processing validate.ErrorProcessing

	Context("smbg from obj", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
		})

		It("returns a bolus if the obj is valid", func() {
			selfMonitored := BuildSelfMonitored(bgObj, processing)
			var bgType *SelfMonitored
			Expect(selfMonitored).To(BeAssignableToTypeOf(bgType))
			Expect(processing.HasErrors()).To(BeFalse())
		})

	})
})
