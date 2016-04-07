package bloodglucose

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("Continuous", func() {
	var bgObj = TestingDatumBase()
	bgObj["type"] = "cbg"
	bgObj["value"] = 5.5
	bgObj["units"] = "mmol/l"
	bgObj["isig"] = 6.5

	var processing validate.ErrorProcessing

	Context("cbg from obj", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
		})

		It("returns a bolus if the obj is valid", func() {
			continuous := BuildContinuous(bgObj, processing)
			var bgType *Continuous
			Expect(continuous).To(BeAssignableToTypeOf(bgType))
			Expect(processing.HasErrors()).To(BeFalse())
		})

	})
})
