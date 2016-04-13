package device

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/data/types"
)

var _ = Describe("DeviceEvent", func() {

	var helper *types.TestingHelper

	BeforeEach(func() {
		helper = types.NewTestingHelper()
	})

	Context("calibration", func() {

		var deviceEventObj = fixtures.TestingDatumBase()
		deviceEventObj["type"] = "deviceEvent"
		deviceEventObj["subType"] = "calibration"
		deviceEventObj["value"] = 3.0
		deviceEventObj["units"] = "mg/dL"

		It("returns a Calibration if the obj is valid", func() {
			Expect(helper.ValidDataType(Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
		})

		Context("validation", func() {})
	})
})
