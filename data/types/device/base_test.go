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

	Context("base", func() {

		Context("alarm subType", func() {
			var deviceEventObj = fixtures.TestingDatumBase()
			deviceEventObj["type"] = "deviceEvent"
			deviceEventObj["subType"] = "alarm"
			deviceEventObj["alarmType"] = "low_insulin"

			It("returns a Alarm if the obj is valid", func() {
				Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
			})
		})

		Context("calibration subType", func() {

			var deviceEventObj = fixtures.TestingDatumBase()
			deviceEventObj["type"] = "deviceEvent"
			deviceEventObj["subType"] = "calibration"
			deviceEventObj["value"] = 3.0
			deviceEventObj["units"] = "mg/dL"

			It("returns a Calibration if the obj is valid", func() {
				Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
			})
		})

	})
})
