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

	Context("timeChange", func() {

		var deviceEventObj = fixtures.TestingDatumBase()
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
			Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
		})

		Context("validation", func() {})
	})
})
