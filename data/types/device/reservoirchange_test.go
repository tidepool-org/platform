package device_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fixtures "github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/device"
)

var _ = Describe("DeviceEvent", func() {

	var helper *types.TestingHelper

	BeforeEach(func() {
		helper = types.NewTestingHelper()
	})
	Context("reservoirChange", func() {

		var deviceEventObj = fixtures.TestingDatumBase()
		deviceEventObj["type"] = "deviceEvent"
		deviceEventObj["subType"] = "reservoirChange"
		deviceEventObj["status"] = "suspended"

		It("returns a ReservoirChange if the obj is valid", func() {
			Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
		})

		Context("validation", func() {})
	})

})
